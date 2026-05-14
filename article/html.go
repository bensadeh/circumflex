package article

import (
	"bytes"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"
)

func convertToMarkdown(article string) (string, error) {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)

	// <a>, <b>, <strong> are unwrapped to plain text; <i>, <em> are wrapped
	// in CLX-ITALIC markers that the renderer later turns into ANSI.
	for _, tag := range []string{"a", "b", "strong"} {
		conv.Register.RendererFor(tag, converter.TagTypeInline, renderUnwrapped, converter.PriorityEarly)
	}

	for _, tag := range []string{"i", "em"} {
		conv.Register.RendererFor(tag, converter.TagTypeInline, renderItalic, converter.PriorityEarly)
	}

	return conv.ConvertString(article)
}

func renderChildText(ctx converter.Context, n *html.Node) string {
	var buf bytes.Buffer
	ctx.RenderChildNodes(ctx, &buf, n)

	return strings.TrimSpace(buf.String())
}

func renderUnwrapped(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	_, _ = w.WriteString(renderChildText(ctx, n))

	return converter.RenderSuccess
}

func renderItalic(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	_, _ = w.WriteString(italicStart + renderChildText(ctx, n) + italicStop)

	return converter.RenderSuccess
}
