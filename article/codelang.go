package article

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/bensadeh/circumflex/highlight"
)

// codeLangAttr carries a <pre> block's declared language across readability,
// which strips the class attributes the language conventions ride on.
const codeLangAttr = "data-clx-lang"

// preserveCodeLang pins each <pre> block's declared language to codeLangAttr
// before readability runs. Rouge and GitHub put the language class on a
// wrapper element rather than the <pre> itself, so nearby ancestors count.
func preserveCodeLang(doc *html.Node) {
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || nodeAtom(n) != atom.Pre {
			continue
		}

		lang := codeLang(n)
		for anc, depth := n.Parent, 0; lang == "" && anc != nil && depth < 3; anc, depth = anc.Parent, depth+1 {
			lang = classLang(attr(anc, "class"))
		}

		if lang != "" {
			n.Attr = append(n.Attr, html.Attribute{Key: codeLangAttr, Val: lang})
		}
	}
}

// codeLang returns the declared language of a <pre> block: the attribute
// preserveCodeLang pinned, Hugo's data-lang, or a language class on the
// element itself or a <code> child (goldmark's fenced-block output).
func codeLang(pre *html.Node) string {
	if lang := langFromNode(pre); lang != "" {
		return lang
	}

	for n := range pre.Descendants() {
		if n.Type == html.ElementNode && nodeAtom(n) == atom.Code {
			if lang := langFromNode(n); lang != "" {
				return lang
			}
		}
	}

	return ""
}

func langFromNode(n *html.Node) string {
	if lang := attr(n, codeLangAttr); lang != "" {
		return normalizeLang(lang)
	}

	if lang := attr(n, "data-lang"); lang != "" {
		return normalizeLang(lang)
	}

	return classLang(attr(n, "class"))
}

// classLang extracts a language from the class conventions highlighters use:
// language-go (WHATWG spec, Prism, goldmark), lang-go (highlight.js), and
// GitHub's highlight-source-go.
func classLang(class string) string {
	for name := range strings.FieldsSeq(class) {
		for _, prefix := range []string{"language-", "lang-", "highlight-source-"} {
			if rest, ok := strings.CutPrefix(name, prefix); ok && rest != "" {
				return normalizeLang(rest)
			}
		}
	}

	return ""
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// No real language alias is this long; a longer value is page noise.
	if len(lang) > 30 {
		return ""
	}

	// "No language" spelled as one: Jekyll stamps language-plaintext on
	// every unlabeled fence, which would otherwise suppress the guesser,
	// drop the faint fallback, and label the box "plaintext".
	switch lang {
	case "plaintext", "text", "plain", "txt", "nohighlight", "none":
		return ""
	}

	// A declaration only counts when it names a real lexer: i18n wrapper
	// classes like lang-en would suppress the guesser, and lang-es would
	// reach the Erlang lexer through chroma's extension fallback.
	if !highlight.Resolves(lang) {
		return ""
	}

	return lang
}
