package article

import (
	"bytes"
	"context"
	"encoding/xml"
	"image"
	"image/png"
	nurl "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	// Blank imports register the decoders that image.Decode dispatches to.
	_ "image/gif"
	_ "image/jpeg"

	_ "github.com/gen2brain/avif" // libavif via wazero; Hugo/Cloudflare pipelines emit AVIF
	_ "golang.org/x/image/webp"   // WordPress and others increasingly serve WebP

	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/version"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	maxImages         = 256 // safety valve against pathological pages, not a working limit
	imageConcurrency  = 8
	imageFetchTimeout = 8 * time.Second
	minImageWidth     = 24 // skip tracking pixels and vertical rules
	minImageHeight    = 32 // also short strips: badges (28px in the tallest shields style), dividers
)

// maxImagePx bounds an image's long side in pixels: the copy a Kitty-graphics
// terminal composites, and so equally the resolution a downloaded source must
// cover. The terminal stretches the copy across the placement's device pixels
// — a full content column on a retina display runs ~1500–2000 of them, and any
// shortfall is upscaled into blur. 2048 keeps the copy on the downscale side of
// that; kittyPNG never inflates smaller sources.
const maxImagePx = 2048

// maxArticleDecodedBytes bounds what an article's images decode to in the
// terminal's Kitty-graphics storage together. kitty and Ghostty quota that
// storage at 320MB and evict oldest-first when it overflows; transmission
// follows document order, so the evicted images are the top-of-page figures
// the reader is most likely viewing (an arXiv paper with 47 full-HD appendix
// screenshots decodes to ~390MB). The margin below the quota is deliberate:
// storage is shared with everything else the session has transmitted and is
// never freed mid-session.
const maxArticleDecodedBytes = 256 * 1000 * 1000

// kittyImage is one image block's terminal-side life: the PNG the terminal
// receives, the ID its placeholder cells dereference, and the placement
// geometry — what the last render wants against what the terminal holds.
// want mutates on render and sent when PendingKittyWork hands the delta to
// the reader; both stay on the update goroutine.
type kittyImage struct {
	png                []byte
	id                 int
	sent               bool
	sentCols, sentRows int
	wantCols, wantRows int
}

// fetchImages downloads and decodes the image blocks in place, resolving
// relative sources against base. Failures leave block.kitty nil, so rendering
// falls back to the text label; images displayed below the size floors are
// marked decorative instead, so rendering can drop them. Only the first
// maxImages are fetched.
func fetchImages(ctx context.Context, blocks []block, base *nurl.URL) {
	var targets []int

	for i := range blocks {
		// Figures fetch like any image: whether the terminal composites the
		// high-res pixels that make them legible isn't known until the
		// graphics probe answers, well after parse.
		if blocks[i].kind == blockImage && blocks[i].imageURL != "" {
			targets = append(targets, i)
			if len(targets) == maxImages {
				break
			}
		}
	}

	if len(targets) == 0 {
		return
	}

	client := resty.New()
	defer func() { _ = client.Close() }()

	client.SetTimeout(imageFetchTimeout)
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetLogger(discardLogger{})

	g := new(errgroup.Group)
	g.SetLimit(imageConcurrency)

	for _, i := range targets {
		g.Go(func() error {
			img, raw, viewBox := fetchImage(ctx, client, base, blocks[i].imageURL)

			// A raster's bounds are its identity; a vector's raster is drawn
			// at the Kitty fidelity ceiling regardless of size, so vectors
			// are judged and sized by their on-page geometry instead — which
			// survives even when rasterization fails, so a badge canvas cannot
			// draw is still recognized as one rather than kept as a label.
			var size image.Point
			if img != nil {
				size = img.Bounds().Size()
			}

			if viewBox != (image.Point{}) {
				size = svgDisplaySize(viewBox, blocks[i].dispWidth)
			}

			if size != (image.Point{}) && (size.X < minImageWidth || size.Y < minImageHeight) {
				blocks[i].decorative = true

				return nil
			}

			if img == nil {
				return nil
			}

			// Materialize the intrinsic-width fallback so imageCols still
			// sizes from the on-page geometry.
			if blocks[i].dispWidth <= 0 {
				blocks[i].dispWidth = size.X
			}

			// Only the PNG outlives the fetch: the terminal holds the pixels,
			// and rendering needs the raster's dimensions alone to keep the
			// cell grid's aspect ratio honest.
			blocks[i].kitty = newKittyImage(img, raw)
			blocks[i].imgSize = img.Bounds().Size()

			return nil
		})
	}

	_ = g.Wait()

	shrinkToBudget(blocks, maxArticleDecodedBytes)
}

// shrinkToBudget resizes an article's images to fit budget bytes of decoded
// pixels together. The long-side ceiling that fits is shared: images above it
// shrink to it, images below keep every pixel — small body figures already sit
// at or under their display size, while oversized screenshots carry the slack.
func shrinkToBudget(blocks []block, budget int) {
	var (
		targets []int
		sizes   []image.Point
	)

	total := 0

	for i := range blocks {
		k := blocks[i].kitty
		if k == nil {
			continue
		}

		size := pngSize(k.png, blocks[i].imgSize)
		targets = append(targets, i)
		sizes = append(sizes, size)
		total += 4 * size.X * size.Y
	}

	if total <= budget {
		return
	}

	ceiling := sharedCeiling(sizes, budget)

	g := new(errgroup.Group)
	g.SetLimit(imageConcurrency)

	for n, i := range targets {
		if scaledSize(sizes[n], ceiling) == sizes[n] {
			continue
		}

		g.Go(func() error {
			shrinkKitty(&blocks[i], ceiling)

			return nil
		})
	}

	_ = g.Wait()
}

// pngSize reads the terminal copy's dimensions from its header; the block's
// imgSize records the source raster, which a kittyPNG downscale may have
// tightened.
func pngSize(data []byte, fallback image.Point) image.Point {
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fallback
	}

	return image.Pt(cfg.Width, cfg.Height)
}

// sharedCeiling finds the largest long-side ceiling that fits the images into
// budget together.
func sharedCeiling(sizes []image.Point, budget int) int {
	lo, hi := 1, maxImagePx

	for lo < hi {
		mid := (lo + hi + 1) / 2

		total := 0

		for _, size := range sizes {
			s := scaledSize(size, mid)
			total += 4 * s.X * s.Y
		}

		if total <= budget {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	return lo
}

// shrinkKitty re-derives the block's terminal copy at a tighter ceiling from
// the PNG in hand — the source bytes are gone by now, and a second small
// downscale of an already-bounded copy costs no visible fidelity. imgSize
// stays the source's: it exists to carry the aspect ratio, which shrinking
// preserves.
func shrinkKitty(b *block, ceiling int) {
	img, err := png.Decode(bytes.NewReader(b.kitty.png))
	if err != nil {
		return
	}

	var buf bytes.Buffer
	if png.Encode(&buf, downscaleTo(img, ceiling)) != nil {
		return
	}

	b.kitty.png = buf.Bytes()
}

// fetchImage downloads and decodes one image. A non-zero viewBox identifies
// the source as a vector, whose decoded bounds are raster fidelity rather
// than intrinsic size; it can arrive without an image when the vector
// declares its geometry but canvas cannot draw it.
func fetchImage(ctx context.Context, client *resty.Client, base *nurl.URL, rawURL string) (image.Image, []byte, image.Point) {
	ref, err := nurl.Parse(rawURL)
	if err != nil {
		return nil, nil, image.Point{}
	}

	target := base.ResolveReference(ref)

	req := client.R().SetContext(ctx)
	if referer := refererFor(base, target); referer != "" {
		req.SetHeader("Referer", referer)
	}

	resp, err := req.Get(target.String())
	if err != nil || resp.StatusCode() >= 400 {
		return nil, nil, image.Point{}
	}

	img, _, err := image.Decode(bytes.NewReader(resp.Bytes()))
	if err != nil {
		svg, viewBox := decodeSVG(resp.Bytes())

		return svg, resp.Bytes(), viewBox
	}

	return img, resp.Bytes(), image.Point{}
}

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

// newKittyImage retains the copy a Kitty-graphics terminal composites. The
// terminal-global image ID is claimed here, at fetch time, so every render
// and walk-back of this article agrees on it. Retention is unconditional —
// the standalone article command parses before the terminal is probed, so
// gating on the probe here would leave it permanently label-only.
func newKittyImage(img image.Image, raw []byte) *kittyImage {
	data := kittyPNG(img, raw)
	if data == nil {
		return nil
	}

	return &kittyImage{png: data, id: graphics.AllocID()}
}

// kittyPNG bounds img to maxImagePx and encodes it as the PNG the terminal
// will receive. Sources that already are a PNG within bounds pass through
// byte-for-byte. The downscale uses Catmull-Rom: this is the copy whose
// entire point is fidelity, and fetch goroutines have the time.
func kittyPNG(img image.Image, raw []byte) []byte {
	bounds := img.Bounds()

	if bytes.HasPrefix(raw, pngMagic) && bounds.Dx() <= maxImagePx && bounds.Dy() <= maxImagePx {
		return raw
	}

	var buf bytes.Buffer
	if png.Encode(&buf, downscaleTo(img, maxImagePx)) != nil {
		return nil
	}

	return buf.Bytes()
}

// downscaleTo bounds img to ceiling on its long side with Catmull-Rom,
// keeping the aspect; within bounds it passes through untouched.
func downscaleTo(img image.Image, ceiling int) image.Image {
	bounds := img.Bounds()

	target := scaledSize(bounds.Size(), ceiling)
	if target == bounds.Size() {
		return img
	}

	dst := image.NewRGBA(image.Rect(0, 0, target.X, target.Y))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, xdraw.Src, nil)

	return dst
}

// scaledSize is size bounded to ceiling on its long side, aspect kept.
func scaledSize(size image.Point, ceiling int) image.Point {
	long := max(size.X, size.Y)
	if long <= ceiling {
		return size
	}

	scale := float64(ceiling) / float64(long)

	return image.Pt(
		max(1, int(float64(size.X)*scale+0.5)),
		max(1, int(float64(size.Y)*scale+0.5)))
}

// refererFor mirrors the browser default referrer policy
// (strict-origin-when-cross-origin): the full page URL same-origin, the bare
// origin cross-origin, nothing on an https→http downgrade. Hotlink-protected
// hosts (e.g. fabiensanglard.net) 403 image requests without it.
func refererFor(page, target *nurl.URL) string {
	if page.Scheme == "https" && target.Scheme != "https" {
		return ""
	}

	if page.Scheme == target.Scheme && page.Host == target.Host {
		full := *page
		full.User = nil
		full.Fragment = ""

		return full.String()
	}

	return page.Scheme + "://" + page.Host + "/"
}

// svgMu serializes SVG rasterization: canvas mutates process-global font state
// while laying out text and is not documented as concurrency-safe, whereas
// images fetch on up to imageConcurrency goroutines.
var svgMu sync.Mutex

// decodeSVG rasterizes an SVG at the fidelity ceiling — a vector has no
// intrinsic resolution, so the raster is drawn at maxImagePx on the long side
// regardless of the size the page gives it. canvas renders text and markers,
// so diagram labels and arrowheads survive the trip to pixels. The returned
// viewBox is the only intrinsic geometry the file has;
// its units are arbitrary (Discord's logo measures 126×20 — millimeters), so
// sizing keys off it or the declared display width, never off the raster
// (canvas restates the geometry in its own millimetre space). An SVG that
// declares no box canvas could still rasterize is dropped to its label: there
// is no on-page size to render against, and sizing off the ~2048px raster would
// blow a small icon up to the full column.
func decodeSVG(data []byte) (img image.Image, viewBox image.Point) {
	viewBox = svgDeclaredBox(data)
	if viewBox == (image.Point{}) {
		return nil, image.Point{}
	}

	return rasterizeSVG(data), viewBox
}

// rasterizeSVG draws the SVG, working down from the copy that renders most to
// the copy that renders at all. The instanced copy goes first because canvas
// draws no part of <use> (see expandSVGUse); it is a rewrite of the source, so
// a copy that will not draw costs only the attempt, and the original still
// gets the passes it would have had. The last of those drops the text: canvas
// panics when a <text> element names a font the host cannot load (getFontFace
// → LoadSystemFont), so on a box missing the declared or default family the
// retry renders the shapes alone — oksvg's text-less output — rather than
// losing the whole figure to its label.
func rasterizeSVG(data []byte) image.Image {
	if instanced := expandSVGUse(data); instanced != nil {
		if img := drawSVG(instanced); img != nil {
			return img
		}
	}

	if img := drawSVG(data); img != nil {
		return img
	}

	if stripped := stripSVGText(data); stripped != nil {
		return drawSVG(stripped)
	}

	return nil
}

// drawSVG parses and rasterizes the SVG, returning nil on any failure. canvas
// keeps drawing past the first construct it flags (an unknown unit, a bad
// color) and hands back a usable canvas alongside the error, so the error is
// non-fatal — only an empty canvas or a panic (malformed path, missing font)
// leaves no image.
func drawSVG(data []byte) (img image.Image) {
	// canvas panics on malformed path data and unloadable fonts; a broken SVG
	// should fall back like any other undecodable image.
	defer func() {
		if recover() != nil {
			img = nil
		}
	}()

	// Mermaid diagrams mark inline styles "!important"; canvas keeps the
	// declaration but drops the value it qualifies, painting the shape its
	// default black — stripping the suffix restores the intended fill.
	data = bytes.ReplaceAll(data, []byte("!important"), []byte(""))

	svgMu.Lock()
	defer svgMu.Unlock()

	c, _ := canvas.ParseSVG(bytes.NewReader(data))
	if c == nil || c.W <= 0 || c.H <= 0 {
		return nil
	}

	resolution := canvas.Resolution(float64(maxImagePx) / max(c.W, c.H))

	return rasterizer.Draw(c, resolution, canvas.DefaultColorSpace)
}

var svgTextElem = regexp.MustCompile(`(?is)<text\b[^>]*>.*?</text>|<text\b[^>]*/>`)

// stripSVGText removes <text> elements so an SVG whose text canvas cannot lay
// out still rasterizes its shapes. It returns nil when there is nothing to
// strip, so the caller skips a redundant second parse.
func stripSVGText(data []byte) []byte {
	if !bytes.Contains(data, []byte("<text")) {
		return nil
	}

	return svgTextElem.ReplaceAll(data, nil)
}

// svgDeclaredBox reads the root <svg> element's own geometry — the viewBox,
// or the width/height attributes when there is none — without rasterizing
// anything. Matches the viewBox canvas rasterizes against, so a vector judged
// this way lands where a drawable copy of it would.
func svgDeclaredBox(data []byte) image.Point {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false

	for {
		tok, err := dec.Token()
		if err != nil {
			return image.Point{}
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		if start.Name.Local != "svg" {
			return image.Point{}
		}

		var width, height float64

		for _, a := range start.Attr {
			switch a.Name.Local {
			case "viewBox":
				if f := strings.Fields(a.Value); len(f) == 4 {
					w, errW := strconv.ParseFloat(f[2], 64)
					h, errH := strconv.ParseFloat(f[3], 64)

					if errW == nil && errH == nil && w > 0 && h > 0 {
						return image.Pt(int(w+0.5), int(h+0.5))
					}
				}
			case "width":
				width = pixelLength(a.Value)
			case "height":
				height = pixelLength(a.Value)
			}
		}

		if width > 0 && height > 0 {
			return image.Pt(int(width+0.5), int(height+0.5))
		}

		return image.Point{}
	}
}

// pixelLength parses a bare or px-suffixed SVG length; percentages and other
// units carry no pixel meaning here and yield 0.
func pixelLength(v string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(v), "px"), 64)
	if err != nil || f <= 0 {
		return 0
	}

	return f
}

// svgDisplaySize is the size a vector occupies on the page: the declared
// width attribute scaled through the viewBox aspect when the page gives one,
// the bare viewBox otherwise — arbitrary units, but the only signal left.
func svgDisplaySize(viewBox image.Point, dispWidth int) image.Point {
	if dispWidth <= 0 {
		return viewBox
	}

	return image.Pt(dispWidth, max(1, (dispWidth*viewBox.Y+viewBox.X/2)/viewBox.X))
}
