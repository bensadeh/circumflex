package article

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// mathSpans renders a MathML element. LaTeXML (behind arXiv's HTML papers)
// keeps the original TeX in the alttext attribute or an annotation child;
// converting that beats flattening the MathML layout tree, whose presentation
// nodes double the content ("n" plus its annotation reads as "nn").
func mathSpans(n *html.Node, format inlineFormat) []span {
	tex := strings.TrimSpace(ansi.Strip(attr(n, "alttext")))
	if tex == "" {
		tex = strings.TrimSpace(texAnnotation(n))
	}

	text := mathMLText(n)
	if tex != "" {
		text = latexToUnicode(tex)
	}

	if text == "" {
		return nil
	}

	return []span{{text: text, format: format}}
}

func texAnnotation(n *html.Node) string {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && nodeAtom(c) == atom.Annotation &&
			strings.Contains(attr(c, "encoding"), "tex") {
			return ansi.Strip(nodeText(c))
		}
	}

	return ""
}

// mathMLText is the fallback for MathML without TeX source: the concatenated
// presentation text, minus annotations, which duplicate it in other formats.
func mathMLText(n *html.Node) string {
	var sb strings.Builder

	var visit func(*html.Node)

	visit = func(c *html.Node) {
		switch c.Type {
		case html.TextNode:
			sb.WriteString(ansi.Strip(c.Data))

		case html.ElementNode:
			if a := nodeAtom(c); a == atom.Annotation || a == atom.AnnotationXml {
				return
			}

			for gc := range c.ChildNodes() {
				visit(gc)
			}

		default:
		}
	}

	for c := range n.ChildNodes() {
		visit(c)
	}

	return strings.TrimSpace(collapseWhitespace(sb.String()))
}

func nodeText(n *html.Node) string {
	var sb strings.Builder

	for c := range n.Descendants() {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}

	return sb.String()
}

// Pages that render math client-side ship raw LaTeX in their HTML; the
// script include is the signal that $-delimited source is meant as math.
func usesMathRenderer(body []byte) bool {
	page := strings.ToLower(string(body))

	return strings.Contains(page, "mathjax") || strings.Contains(page, "katex")
}

func convertMath(blocks []block) {
	for i := range blocks {
		b := &blocks[i]

		switch b.kind {
		case blockParagraph, blockQuote, blockImage:
			convertMathSpans(b.spans)

		case blockHeading:
			b.text = convertMathText(b.text)

		case blockList:
			for j := range b.items {
				convertMathSpans(b.items[j].spans)
			}

		case blockTable:
			for _, row := range b.rows {
				for k := range row {
					row[k] = convertMathText(row[k])
				}
			}

		case blockCode, blockDivider, blockVerbatim:
		}
	}
}

// Inline code spans keep their dollar signs: shell and template snippets use
// $ far more often than math does.
func convertMathSpans(spans []span) {
	for i := range spans {
		if spans[i].format == formatCode {
			continue
		}

		spans[i].text = convertMathText(spans[i].text)
	}
}

// The inline pattern mirrors MathJax's pairing rule: the content must not
// touch its delimiters with whitespace, so "cost $5 and $10" cannot pair.
var (
	displayMathPattern = regexp.MustCompile(`\$\$([^$]+)\$\$|\\\[(?s:.+?)\\\]`)
	inlineMathPattern  = regexp.MustCompile(`\$([^\s$](?:[^$]*[^\s$])?)\$|\\\((?s:.+?)\\\)`)
)

func convertMathText(text string) string {
	if !strings.ContainsRune(text, '$') && !strings.Contains(text, `\(`) && !strings.Contains(text, `\[`) {
		return text
	}

	text = displayMathPattern.ReplaceAllStringFunc(text, func(m string) string {
		return latexToUnicode(stripMathDelimiters(m))
	})

	return inlineMathPattern.ReplaceAllStringFunc(text, func(m string) string {
		inner := stripMathDelimiters(m)
		if strings.HasPrefix(m, `\(`) || looksLikeMath(inner) {
			return latexToUnicode(inner)
		}

		return m
	})
}

func stripMathDelimiters(m string) string {
	width := 1
	if strings.HasPrefix(m, "$$") || strings.HasPrefix(m, `\`) {
		width = 2
	}

	return strings.TrimSpace(m[width : len(m)-width])
}

var mathNumber = regexp.MustCompile(`^[0-9]+(/[0-9]+)?$`)

// looksLikeMath separates $x^2$ from prose dollar amounts: single-$ content
// must show LaTeX syntax, be a lone symbol, or be a bare number or fraction.
func looksLikeMath(s string) bool {
	if strings.ContainsAny(s, `\^_=`) {
		return true
	}

	if utf8.RuneCountInString(s) == 1 {
		return true
	}

	return mathNumber.MatchString(s)
}

// latexToUnicode converts the common LaTeX vocabulary to Unicode: symbol
// commands, Greek letters, font wrappers, fractions, roots, and sub- and
// superscripts. Anything it cannot map degrades to a readable plain form
// (g^{\mathsf{sk}_a} becomes g^(skₐ)) rather than raw source.
func latexToUnicode(src string) string {
	s := &texScanner{src: src}

	var out strings.Builder

	for {
		out.WriteString(s.sequence())

		if s.pos >= len(s.src) {
			break
		}

		s.pos++ // stray closing brace
	}

	return bracketSpaces.Replace(strings.Join(strings.Fields(out.String()), " "))
}

var bracketSpaces = strings.NewReplacer(
	"⌊ ", "⌊", " ⌋", "⌋",
	"⌈ ", "⌈", " ⌉", "⌉",
	"⟨ ", "⟨", " ⟩", "⟩",
)

// textMode suppresses the variables-are-italic rule inside upright wrappers
// like \text and \mathsf.
type texScanner struct {
	src      string
	pos      int
	textMode bool
}

func (s *texScanner) sequence() string {
	var out strings.Builder

	for s.pos < len(s.src) {
		switch c := s.src[s.pos]; c {
		case '}':
			return out.String()

		case '{':
			s.pos++
			out.WriteString(s.sequence())
			s.skipByte('}')

		case '\\':
			out.WriteString(s.command())

		case '^':
			s.pos++
			out.WriteString(superscript(s.arg()))

		case '_':
			s.pos++
			out.WriteString(subscript(s.arg()))

		case '~', '&':
			s.pos++

			out.WriteByte(' ')

		case '\'':
			s.pos++

			out.WriteString("′")

		default:
			r, size := utf8.DecodeRuneInString(s.src[s.pos:])
			s.pos += size

			if !s.textMode {
				r = mathItalicRune(r)
			}

			out.WriteRune(r)
		}
	}

	return out.String()
}

func (s *texScanner) skipByte(b byte) {
	if s.pos < len(s.src) && s.src[s.pos] == b {
		s.pos++
	}
}

func (s *texScanner) command() string {
	s.pos++ // the backslash

	if s.pos >= len(s.src) {
		return ""
	}

	start := s.pos
	for s.pos < len(s.src) && isTexLetter(s.src[s.pos]) {
		s.pos++
	}

	if s.pos == start {
		c := s.src[s.pos]
		s.pos++

		switch c {
		case '\\', ',', ';', ':':
			return " "
		case '!':
			return ""
		default:
			return string(c)
		}
	}

	return s.expand(s.src[start:s.pos])
}

func isTexLetter(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func (s *texScanner) expand(name string) string {
	switch name {
	case "frac", "tfrac", "dfrac":
		num, den := s.arg(), s.arg()

		return parenthesizeTerm(num) + "/" + parenthesizeTerm(den)

	case "sqrt":
		return s.root()

	case "pmod":
		return " (mod " + s.arg() + ")"

	case "bmod", "mod":
		return " mod "

	case "mathbb":
		return mapRunes(s.uprightArg(), doubleStruckRune)

	case "mathcal":
		return mapRunes(s.uprightArg(), scriptStyleRune)

	case "left", "right", "big", "Big", "bigg", "Bigg",
		"bigl", "bigr", "Bigl", "Bigr", "biggl", "biggr", "Biggl", "Biggr":
		// \left. and \right. are invisible delimiters; the dot goes with them.
		s.skipByte('.')

		return ""

	case "phantom", "vphantom", "hphantom":
		s.arg()

		return ""
	}

	if texWrappers[name] {
		return s.uprightArg()
	}

	if accent, ok := texAccents[name]; ok {
		arg := s.arg()
		if utf8.RuneCountInString(arg) == 1 {
			return arg + string(accent)
		}

		return arg
	}

	if symbol, ok := texSymbols[name]; ok {
		return symbol
	}

	return name // unknown command: keep its name, drop the backslash
}

func (s *texScanner) arg() string {
	for s.pos < len(s.src) && s.src[s.pos] == ' ' {
		s.pos++
	}

	if s.pos >= len(s.src) {
		return ""
	}

	switch s.src[s.pos] {
	case '{':
		s.pos++
		inner := s.sequence()
		s.skipByte('}')

		return inner

	case '\\':
		return s.command()
	}

	r, size := utf8.DecodeRuneInString(s.src[s.pos:])
	s.pos += size

	if !s.textMode {
		r = mathItalicRune(r)
	}

	return string(r)
}

func (s *texScanner) uprightArg() string {
	prev := s.textMode
	s.textMode = true
	arg := s.arg()
	s.textMode = prev

	return arg
}

func (s *texScanner) root() string {
	radical := "√"

	if s.pos < len(s.src) && s.src[s.pos] == '[' {
		if end := strings.IndexByte(s.src[s.pos:], ']'); end >= 0 {
			index := s.src[s.pos+1 : s.pos+end]
			s.pos += end + 1

			switch index {
			case "3":
				radical = "∛"
			case "4":
				radical = "∜"
			default:
				radical = superscript(index) + "√"
			}
		}
	}

	arg := s.arg()
	if strings.ContainsAny(arg, " +-·×÷/=") {
		return radical + "(" + arg + ")"
	}

	return radical + arg
}

func parenthesizeTerm(s string) string {
	if strings.ContainsAny(s, " +-·×÷/=,") {
		return "(" + s + ")"
	}

	return s
}

func superscript(text string) string {
	if converted, ok := runesToScript(text, superscriptRunes); ok {
		return converted
	}

	// ^* and ^′ read as-is in plain notation: g^* prints as g*.
	if text == "*" || text == "′" {
		return text
	}

	return plainScript("^", text)
}

func subscript(text string) string {
	if converted, ok := runesToScript(text, subscriptRunes); ok {
		return converted
	}

	return plainScript("_", text)
}

// runesToScript maps sub/superscript text to Unicode only when every rune has
// a form. The forms are upright, so italic variables map back to ASCII first:
// ^{c} and ^c both come out as ᶜ.
func runesToScript(text string, mapping map[rune]rune) (string, bool) {
	var sb strings.Builder

	for _, r := range text {
		m, ok := mapping[unMathItalicRune(r)]
		if !ok {
			return "", false
		}

		sb.WriteRune(m)
	}

	return sb.String(), true
}

// plainScript is the marker-notation fallback: x^(a+b), log_𝑔.
func plainScript(marker, text string) string {
	if text == "" {
		return ""
	}

	if utf8.RuneCountInString(text) == 1 {
		return marker + text
	}

	return marker + "(" + text + ")"
}

func mapRunes(s string, mapping func(rune) rune) string {
	var sb strings.Builder

	for _, r := range s {
		sb.WriteRune(mapping(r))
	}

	return sb.String()
}

// Raw letters in math mode are variables, which LaTeX sets in italic;
// command output like \log or \text{…} stays upright.
func mathItalicRune(r rune) rune {
	switch {
	case r == 'h':
		return 'ℎ' // the Planck constant predates the math italic block
	case r >= 'A' && r <= 'Z':
		return 0x1D434 + r - 'A'
	case r >= 'a' && r <= 'z':
		return 0x1D44E + r - 'a'
	}

	return r
}

func unMathItalicRune(r rune) rune {
	switch {
	case r == 'ℎ':
		return 'h'
	case r >= 0x1D434 && r <= 0x1D44D:
		return 'A' + (r - 0x1D434)
	case r >= 0x1D44E && r <= 0x1D467:
		return 'a' + (r - 0x1D44E)
	}

	return r
}

// Letters predating the Unicode mathematical alphanumeric blocks live in
// Letterlike Symbols, so the contiguous ranges have holes to special-case.
func doubleStruckRune(r rune) rune {
	switch r {
	case 'C':
		return 'ℂ'
	case 'H':
		return 'ℍ'
	case 'N':
		return 'ℕ'
	case 'P':
		return 'ℙ'
	case 'Q':
		return 'ℚ'
	case 'R':
		return 'ℝ'
	case 'Z':
		return 'ℤ'
	}

	switch {
	case r >= 'A' && r <= 'Z':
		return 0x1D538 + r - 'A'
	case r >= 'a' && r <= 'z':
		return 0x1D552 + r - 'a'
	case r >= '0' && r <= '9':
		return 0x1D7D8 + r - '0'
	}

	return r
}

func scriptStyleRune(r rune) rune {
	switch r {
	case 'B':
		return 'ℬ'
	case 'E':
		return 'ℰ'
	case 'F':
		return 'ℱ'
	case 'H':
		return 'ℋ'
	case 'I':
		return 'ℐ'
	case 'L':
		return 'ℒ'
	case 'M':
		return 'ℳ'
	case 'R':
		return 'ℛ'
	case 'e':
		return 'ℯ'
	case 'g':
		return 'ℊ'
	case 'o':
		return 'ℴ'
	}

	switch {
	case r >= 'A' && r <= 'Z':
		return 0x1D49C + r - 'A'
	case r >= 'a' && r <= 'z':
		return 0x1D4B6 + r - 'a'
	}

	return r
}

var texWrappers = map[string]bool{
	"text": true, "textrm": true, "textit": true, "textbf": true, "textsf": true,
	"texttt": true, "textnormal": true, "mbox": true, "hbox": true,
	"mathrm": true, "mathit": true, "mathbf": true, "mathsf": true, "mathtt": true,
	"operatorname": true, "boldsymbol": true, "bm": true,
}

var texAccents = map[string]rune{
	"hat": 0x0302, "widehat": 0x0302, "tilde": 0x0303, "widetilde": 0x0303,
	"bar": 0x0304, "overline": 0x0305, "vec": 0x20D7, "dot": 0x0307,
	"ddot": 0x0308, "check": 0x030C, "breve": 0x0306, "acute": 0x0301, "grave": 0x0300,
}

var texSymbols = map[string]string{
	// operators and relations
	"times": "×", "cdot": "·", "div": "÷", "pm": "±", "mp": "∓",
	"approx": "≈", "sim": "∼", "simeq": "≃", "cong": "≅", "equiv": "≡", "propto": "∝",
	"neq": "≠", "ne": "≠", "leq": "≤", "le": "≤", "geq": "≥", "ge": "≥",
	"ll": "≪", "gg": "≫", "in": "∈", "notin": "∉", "ni": "∋",
	"subset": "⊂", "subseteq": "⊆", "supset": "⊃", "supseteq": "⊇",
	"cup": "∪", "cap": "∩", "setminus": "∖", "emptyset": "∅", "varnothing": "∅",
	"forall": "∀", "exists": "∃", "nexists": "∄",
	"neg": "¬", "lnot": "¬", "land": "∧", "wedge": "∧", "lor": "∨", "vee": "∨",
	"oplus": "⊕", "ominus": "⊖", "otimes": "⊗", "odot": "⊙",
	"circ": "∘", "bullet": "•", "star": "⋆", "ast": "*",
	"sum": "∑", "prod": "∏", "int": "∫", "oint": "∮",
	"partial": "∂", "nabla": "∇", "infty": "∞",
	"to": "→", "rightarrow": "→", "leftarrow": "←", "gets": "←", "leftrightarrow": "↔",
	"Rightarrow": "⇒", "Leftarrow": "⇐", "Leftrightarrow": "⇔",
	"implies": "⇒", "iff": "⇔", "mapsto": "↦",
	"langle": "⟨", "rangle": "⟩", "lfloor": "⌊", "rfloor": "⌋", "lceil": "⌈", "rceil": "⌉",
	"ldots": "…", "cdots": "⋯", "dots": "…", "vdots": "⋮", "ddots": "⋱",
	"prime": "′", "perp": "⊥", "parallel": "∥", "mid": "|", "angle": "∠",
	"ell": "ℓ", "hbar": "ℏ", "Re": "ℜ", "Im": "ℑ", "aleph": "ℵ", "wp": "℘",
	"surd": "√", "dagger": "†", "ddagger": "‡", "checkmark": "✓",

	// greek
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ",
	"epsilon": "ε", "varepsilon": "ε", "zeta": "ζ", "eta": "η",
	"theta": "θ", "vartheta": "ϑ", "iota": "ι", "kappa": "κ",
	"lambda": "λ", "mu": "μ", "nu": "ν", "xi": "ξ",
	"pi": "π", "varpi": "ϖ", "rho": "ρ", "varrho": "ϱ",
	"sigma": "σ", "varsigma": "ς", "tau": "τ", "upsilon": "υ",
	"phi": "φ", "varphi": "φ", "chi": "χ", "psi": "ψ", "omega": "ω",
	"Gamma": "Γ", "Delta": "Δ", "Theta": "Θ", "Lambda": "Λ", "Xi": "Ξ",
	"Pi": "Π", "Sigma": "Σ", "Upsilon": "Υ", "Phi": "Φ", "Psi": "Ψ", "Omega": "Ω",

	// named functions keep their name
	"log": "log", "ln": "ln", "lg": "lg", "exp": "exp",
	"sin": "sin", "cos": "cos", "tan": "tan", "cot": "cot", "sec": "sec", "csc": "csc",
	"arcsin": "arcsin", "arccos": "arccos", "arctan": "arctan",
	"sinh": "sinh", "cosh": "cosh", "tanh": "tanh",
	"min": "min", "max": "max", "inf": "inf", "sup": "sup", "lim": "lim",
	"gcd": "gcd", "det": "det", "dim": "dim", "deg": "deg", "ker": "ker", "Pr": "Pr",

	// spacing
	"quad": " ", "qquad": " ", "thinspace": " ", "enspace": " ",

	// style and positioning switches with no plain-text counterpart
	"displaystyle": "", "textstyle": "", "scriptstyle": "", "scriptscriptstyle": "",
	"limits": "", "nolimits": "",
}
