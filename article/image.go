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

	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	maxImages         = 24
	imageConcurrency  = 8
	imageFetchTimeout = 8 * time.Second
	minImageDimension = 24 // skip tracking pixels and tiny icons
)

// fetchImages downloads and decodes the image blocks in place, resolving
// relative sources against base. Failures leave block.img nil, so rendering
// falls back to the text label. Only the first maxImages are fetched.
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
			if img := fetchImage(ctx, client, base, blocks[i].imageURL); img != nil {
				blocks[i].img = img
			}

			return nil
		})
	}

	_ = g.Wait()
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
		return nil
	}

	if bounds := img.Bounds(); bounds.Dx() < minImageDimension || bounds.Dy() < minImageDimension {
		return nil
	}

	return img
}
