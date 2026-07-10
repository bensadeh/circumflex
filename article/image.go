package article

import (
	"bytes"
	"context"
	"image"
	nurl "net/url"
	"time"

	// Blank imports register the decoders that image.Decode dispatches to.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp" // WordPress and others increasingly serve WebP

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
	maxSVGRasterPx    = 2048 // a vector has no intrinsic resolution; bound the temporary raster
)

// fetchImages downloads and decodes the image blocks in place, resolving
// relative sources against base. Failures leave block.img nil, so rendering
// falls back to the text label; images below minImageDimension are marked
// decorative instead, so rendering can drop them. Only the first maxImages
// are fetched.
func fetchImages(ctx context.Context, blocks []block, base *nurl.URL) {
	var targets []int

	for i := range blocks {
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
			img := fetchImage(ctx, client, base, blocks[i].imageURL)
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

func fetchImage(ctx context.Context, client *resty.Client, base *nurl.URL, rawURL string) image.Image {
	ref, err := nurl.Parse(rawURL)
	if err != nil {
		return nil
	}

	resp, err := client.R().SetContext(ctx).Get(base.ResolveReference(ref).String())
	if err != nil || resp.StatusCode() >= 400 {
		return nil
	}

	img, _, err := image.Decode(bytes.NewReader(resp.Bytes()))
	if err != nil {
		return decodeSVG(resp.Bytes())
	}

	return img
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
