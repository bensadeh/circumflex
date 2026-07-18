package article

import (
	"bytes"
	"context"
	"image"
	"image/png"
	nurl "net/url"
	"time"

	// Blank imports register the decoders that image.Decode dispatches to.
	_ "image/gif"
	_ "image/jpeg"

	_ "github.com/gen2brain/avif" // libavif via wazero; Hugo/Cloudflare pipelines emit AVIF
	_ "golang.org/x/image/webp"   // WordPress and others increasingly serve WebP

	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/version"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	maxImages         = 256 // safety valve against pathological pages, not a working limit
	imageConcurrency  = 8
	imageFetchTimeout = 8 * time.Second
	minImageDimension = 24   // skip tracking pixels and tiny icons
	maxRetainedPx     = 512  // decoded images are downscaled to fit this box; see boundImage
	maxKittyPx        = 1024 // the high-res copy a Kitty-graphics terminal displays; see kittyPNG
	maxSVGRasterPx    = 2048 // a vector has no intrinsic resolution; bound the temporary raster
)

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
// relative sources against base. Failures leave block.img nil, so rendering
// falls back to the text label; images below minImageDimension are marked
// decorative instead, so rendering can drop them. Only the first maxImages
// are fetched.
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
			img, raw := fetchImage(ctx, client, base, blocks[i].imageURL)
			if img == nil {
				return nil
			}

			if bounds := img.Bounds(); bounds.Dx() < minImageDimension || bounds.Dy() < minImageDimension {
				blocks[i].decorative = true

				return nil
			}

			// Materialize the intrinsic-width fallback before downscaling
			// so imageCols still sizes from the original resolution.
			if blocks[i].dispWidth <= 0 {
				blocks[i].dispWidth = img.Bounds().Dx()
			}

			blocks[i].kitty = newKittyImage(img, raw)
			blocks[i].img = boundImage(img)

			return nil
		})
	}

	_ = g.Wait()
}

// boundImage downscales img to fit within maxRetainedPx in both dimensions.
// Rendering samples at most the content column's width in cells and
// maxImageRows*2 pixel rows — far below this box — so nothing visible is
// lost, while a full-size photo would otherwise hold tens of megabytes of
// pixels for a handful of glyphs.
func boundImage(img image.Image) image.Image {
	bounds := img.Bounds()

	width, height := bounds.Dx(), bounds.Dy()
	if width <= maxRetainedPx && height <= maxRetainedPx {
		return img
	}

	scale := float64(maxRetainedPx) / float64(max(width, height))
	dst := image.NewRGBA(image.Rect(0, 0,
		max(1, int(float64(width)*scale+0.5)),
		max(1, int(float64(height)*scale+0.5))))

	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, xdraw.Src, nil)

	return dst
}

func fetchImage(ctx context.Context, client *resty.Client, base *nurl.URL, rawURL string) (image.Image, []byte) {
	ref, err := nurl.Parse(rawURL)
	if err != nil {
		return nil, nil
	}

	target := base.ResolveReference(ref)

	req := client.R().SetContext(ctx)
	if referer := refererFor(base, target); referer != "" {
		req.SetHeader("Referer", referer)
	}

	resp, err := req.Get(target.String())
	if err != nil || resp.StatusCode() >= 400 {
		return nil, nil
	}

	img, _, err := image.Decode(bytes.NewReader(resp.Bytes()))
	if err != nil {
		return decodeSVG(resp.Bytes()), resp.Bytes()
	}

	return img, resp.Bytes()
}

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

// newKittyImage retains the high-resolution copy a Kitty-graphics terminal
// displays in place of half-block art. The terminal-global image ID is
// claimed here, at fetch time, so every render and walk-back of this
// article agrees on it. Retention is unconditional — the standalone article
// command parses before the terminal is probed, so gating on the probe here
// would leave it permanently low-res.
func newKittyImage(img image.Image, raw []byte) *kittyImage {
	data := kittyPNG(img, raw)
	if data == nil {
		return nil
	}

	return &kittyImage{png: data, id: graphics.AllocID()}
}

// kittyPNG bounds img to maxKittyPx and encodes it as the PNG the terminal
// will receive. Sources that already are a PNG within bounds pass through
// byte-for-byte. The downscale uses Catmull-Rom: this is the copy whose
// entire point is fidelity, and fetch goroutines have the time.
func kittyPNG(img image.Image, raw []byte) []byte {
	bounds := img.Bounds()

	if bytes.HasPrefix(raw, pngMagic) && bounds.Dx() <= maxKittyPx && bounds.Dy() <= maxKittyPx {
		return raw
	}

	if bounds.Dx() > maxKittyPx || bounds.Dy() > maxKittyPx {
		scale := float64(maxKittyPx) / float64(max(bounds.Dx(), bounds.Dy()))
		dst := image.NewRGBA(image.Rect(0, 0,
			max(1, int(float64(bounds.Dx())*scale+0.5)),
			max(1, int(float64(bounds.Dy())*scale+0.5))))

		xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, xdraw.Src, nil)

		img = dst
	}

	var buf bytes.Buffer
	if png.Encode(&buf, img) != nil {
		return nil
	}

	return buf.Bytes()
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

// decodeSVG rasterizes an SVG at its viewBox size, so downstream treats it
// exactly like a decoded raster image (dispWidth from bounds, boundImage
// downscale). Text elements are not supported by oksvg and are dropped —
// invisible at terminal resolution anyway.
func decodeSVG(data []byte) (img image.Image) {
	// oksvg panics on some malformed path data; a broken SVG should fall
	// back to the text label like any other undecodable image.
	defer func() {
		if recover() != nil {
			img = nil
		}
	}()

	// Mermaid diagrams mark inline styles "!important"; oksvg's color parser
	// rejects the suffix and the whole parse fails with it.
	data = bytes.ReplaceAll(data, []byte("!important"), []byte(""))

	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil || icon.ViewBox.W <= 0 || icon.ViewBox.H <= 0 {
		return nil
	}

	scale := 1.0
	if long := max(icon.ViewBox.W, icon.ViewBox.H); long > maxSVGRasterPx {
		scale = maxSVGRasterPx / long
	}

	width := max(1, int(icon.ViewBox.W*scale+0.5))
	height := max(1, int(icon.ViewBox.H*scale+0.5))

	icon.SetTarget(0, 0, float64(width), float64(height))

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	icon.Draw(rasterx.NewDasher(width, height, rasterx.NewScannerGV(width, height, rgba, rgba.Bounds())), 1.0)

	return rgba
}
