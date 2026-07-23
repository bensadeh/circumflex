package article

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

const (
	// A referenced element may itself hold a <use>, so each pass resolves one
	// level of indirection. Four covers real documents — a matplotlib plot
	// needs one — while bounding a self-referential chain, which resolves
	// forever.
	maxUseDepth = 4

	// Instances spliced and bytes the rewrite may reach: backstops against a
	// <use> graph that fans out exponentially, not working limits. The plot
	// that motivated this holds 195 instances and roughly triples in size.
	maxUseExpansions = 8192
	maxUseBytes      = 8 << 20
)

// useTargets are the elements a <use> is expanded against: the shapes canvas
// can draw, plus the group that collects them. <symbol> and nested <svg> are
// deliberately absent — each establishes a viewport of its own from its
// width/height/viewBox, and placing a clone wrongly is worse than the current
// behavior, where the instance simply goes undrawn.
var useTargets = map[string]bool{
	"circle":   true,
	"ellipse":  true,
	"g":        true,
	"line":     true,
	"path":     true,
	"polygon":  true,
	"polyline": true,
	"rect":     true,
	"text":     true,
}

// useReserved are the <use> attributes that describe the reference rather than
// the instance: they name the target, place it (folded into the wrapper's
// transform), or size a viewport only the target kinds we decline establish.
// Everything else — style, fill, class — is presentation the clone has to
// inherit, so it rides the wrapper.
var useReserved = map[string]bool{
	"height":     true,
	"href":       true,
	"id":         true,
	"transform":  true,
	"width":      true,
	"x":          true,
	"xlink:href": true,
	"y":          true,
}

// svgUnpainted are elements whose content defines something rather than draws
// it. canvas implements none of them and skips only what <defs> encloses, so
// it already paints the children of a clip path written outside one; expanding
// an instance in there would add ink to a drawing that is wrong already.
// <defs> itself is absent: canvas drops that subtree whole, so an instance
// inside one is inert either way, and the clone it lands in gets resolved
// wherever it is spliced.
var svgUnpainted = map[string]bool{
	"clipPath":       true,
	"filter":         true,
	"linearGradient": true,
	"marker":         true,
	"mask":           true,
	"pattern":        true,
	"radialGradient": true,
	"symbol":         true,
}

// svgSpan is a byte range in the source document.
type svgSpan struct {
	start, end int
}

// svgAttr is one attribute as it was written, delimiters stripped.
type svgAttr struct {
	name, value string
}

// svgUse is one <use> instance: where it sits in the source, the id it names,
// and what its expansion carries onto the wrapping group.
type svgUse struct {
	span      svgSpan
	href      string
	x, y      string
	transform string
	attrs     []svgAttr
}

// expandSVGUse rewrites every <use> instance into an inlined copy of the
// element it references, wrapped in the group the spec defines the instance to
// be equivalent to. canvas implements no part of <use>: it never reads
// xlink:href and has no case for the element, so each instance is dropped with
// no error and a non-empty canvas — nothing downstream can tell that half the
// drawing is missing. Matplotlib's default svg.fonttype ("path") emits every
// glyph as a <path> under <defs> plus one <use> per occurrence, so its plots
// otherwise rasterize with their bars and gridlines intact and every label,
// tick and title gone.
//
// It returns nil when there is nothing to expand or the source resists
// rewriting, so the caller falls back to the untouched bytes and lands exactly
// where it would have without any of this.
func expandSVGUse(data []byte) []byte {
	if !bytes.Contains(data, []byte("<use")) {
		return nil
	}

	budget := maxUseExpansions
	out, expanded := data, false

	for range maxUseDepth {
		next := expandUsePass(out, &budget)
		if next == nil {
			break
		}

		out, expanded = next, true
	}

	if !expanded {
		return nil
	}

	return out
}

// expandUsePass splices every instance it can resolve in one walk of the
// document, returning nil when it resolves none — the source is then either
// free of expandable instances or beyond what this can rewrite.
func expandUsePass(data []byte, budget *int) []byte {
	ids, uses, ok := scanSVGUse(data)
	if !ok {
		return nil
	}

	var out bytes.Buffer

	prev, spliced := 0, 0

	for _, use := range uses {
		target, known := ids[use.href]

		// An instance nested inside one already spliced was copied out with
		// its parent; splicing it again would run backwards through the
		// source. The copy is resolved on the next pass.
		if !known || *budget <= 0 || use.span.start < prev {
			continue
		}

		wrapper, ok := useWrapper(use)
		if !ok {
			continue
		}

		out.Write(data[prev:use.span.start])
		out.WriteString(wrapper)
		out.Write(data[target.start:target.end])
		out.WriteString("</g>")

		prev = use.span.end
		spliced++
		*budget--

		if out.Len() > maxUseBytes {
			return nil
		}
	}

	if spliced == 0 {
		return nil
	}

	out.Write(data[prev:])

	return out.Bytes()
}

// useWrapper builds the opening tag of the group an instance stands for: the
// attributes that are the instance's own, and a transform that places it — the
// element's own transform first, then its offset, the order the generated
// group carries them in.
func useWrapper(use svgUse) (string, bool) {
	x, ok := useOffset(use.x)
	if !ok {
		return "", false
	}

	y, ok := useOffset(use.y)
	if !ok {
		return "", false
	}

	transform := use.transform

	if x != 0 || y != 0 {
		offset := "translate(" + formatSVGNumber(x) + " " + formatSVGNumber(y) + ")"
		transform = strings.TrimSpace(transform + " " + offset)
	}

	var b strings.Builder

	b.WriteString("<g")

	if transform != "" && !writeSVGAttr(&b, "transform", transform) {
		return "", false
	}

	for _, a := range use.attrs {
		if !writeSVGAttr(&b, a.name, a.value) {
			return "", false
		}
	}

	b.WriteString(">")

	return b.String(), true
}

// writeSVGAttr re-emits an attribute lifted from the source. Its value keeps
// whatever escaping it arrived with, so only the delimiter has to be chosen; a
// value holding both delimiters is declined rather than escaped, since leaving
// that one instance undrawn is the behavior without any of this.
func writeSVGAttr(b *strings.Builder, name, value string) bool {
	quote := `"`

	if strings.Contains(value, `"`) {
		if strings.Contains(value, "'") {
			return false
		}

		quote = "'"
	}

	b.WriteString(" ")
	b.WriteString(name)
	b.WriteString("=")
	b.WriteString(quote)
	b.WriteString(value)
	b.WriteString(quote)

	return true
}

// useOffset reads an instance's x/y placement. Only unitless user-space
// numbers are taken: a percentage resolves against the viewport and a physical
// unit against the DPI, neither of which composes into a transform.
func useOffset(v string) (float64, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, true
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}

	return f, true
}

func formatSVGNumber(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// useScanner collects what a splice needs from one walk of the document: where
// each referenceable element sits, and every instance in document order.
type useScanner struct {
	l    *xml.Lexer
	z    *parse.Input
	ids  map[string]svgSpan
	uses []svgUse
	open []svgOpenElem

	// depth of open svgUnpainted elements; instances below one are left alone.
	unpainted int
}

// svgOpenElem is an element whose end tag has not been reached yet. use is the
// index of the instance it is, or -1.
type svgOpenElem struct {
	name  string
	id    string
	start int
	use   int
}

// scanSVGUse walks the document once, recording the byte range of every
// element that can be referenced and of every <use>. It declines a document
// whose tags do not balance: the ranges are computed from the nesting, so a
// range there could bound something other than the element it names.
func scanSVGUse(data []byte) (map[string]svgSpan, []svgUse, bool) {
	z := parse.NewInput(bytes.NewReader(data))
	defer z.Restore()

	s := useScanner{l: xml.NewLexer(z), z: z, ids: map[string]svgSpan{}}

	for {
		tt, tag := s.l.Next()

		if tt == xml.ErrorToken {
			if !errors.Is(s.l.Err(), io.EOF) || len(s.open) != 0 {
				return nil, nil, false
			}

			return s.ids, s.uses, true
		}

		// A processing instruction carries attributes but no content and does
		// not nest, so it never joins the open stack.
		if tt == xml.StartTagPIToken {
			if closeTT, _ := readSVGAttrs(s.l); closeTT == xml.ErrorToken {
				return nil, nil, false
			}

			continue
		}

		if tt == xml.StartTagToken {
			if !s.openTag(tag) {
				return nil, nil, false
			}

			continue
		}

		if tt == xml.EndTagToken && !s.closeTag() {
			return nil, nil, false
		}
	}
}

// openTag records a start tag, resolving it immediately when the tag closes
// itself and deferring to its end tag otherwise. tag is the raw lexeme, whose
// length places the element's first byte.
func (s *useScanner) openTag(tag []byte) bool {
	elem := svgOpenElem{
		name:  string(s.l.Text()),
		start: s.z.Offset() - len(tag),
		use:   -1,
	}

	closeTT, attrs := readSVGAttrs(s.l)
	if closeTT == xml.ErrorToken {
		return false
	}

	for _, a := range attrs {
		if a.name == "id" && useTargets[elem.name] {
			elem.id = a.value
		}
	}

	if elem.name == "use" && s.unpainted == 0 {
		elem.use = len(s.uses)
		s.uses = append(s.uses, newSVGUse(attrs))
	}

	if closeTT != xml.StartTagCloseVoidToken {
		if svgUnpainted[elem.name] {
			s.unpainted++
		}

		s.open = append(s.open, elem)

		return true
	}

	s.resolve(elem, svgSpan{elem.start, s.z.Offset()})

	return true
}

// closeTag matches an end tag against the innermost open element and gives it
// the span the pair encloses.
func (s *useScanner) closeTag() bool {
	if len(s.open) == 0 {
		return false
	}

	elem := s.open[len(s.open)-1]
	s.open = s.open[:len(s.open)-1]

	if elem.name != string(s.l.Text()) {
		return false
	}

	if svgUnpainted[elem.name] {
		s.unpainted--
	}

	s.resolve(elem, svgSpan{elem.start, s.z.Offset()})

	return true
}

func (s *useScanner) resolve(elem svgOpenElem, span svgSpan) {
	if elem.id != "" {
		s.ids[elem.id] = span
	}

	if elem.use >= 0 {
		s.uses[elem.use].span = span
	}
}

// newSVGUse sorts an instance's attributes into the reference it makes and the
// presentation its expansion inherits.
func newSVGUse(attrs []svgAttr) svgUse {
	use := svgUse{}
	hasHref := false

	for _, a := range attrs {
		switch a.name {
		case "href":
			use.href, hasHref = localSVGRef(a.value), true
		case "xlink:href":
			if !hasHref {
				use.href = localSVGRef(a.value)
			}
		case "x":
			use.x = a.value
		case "y":
			use.y = a.value
		case "transform":
			use.transform = a.value
		}

		if !useReserved[a.name] {
			use.attrs = append(use.attrs, a)
		}
	}

	return use
}

// localSVGRef reduces a reference to the id it names, keeping only
// same-document fragments: a reference into another file would have to be
// fetched, and this runs on bytes already in hand.
func localSVGRef(v string) string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "#") {
		return ""
	}

	return v[1:]
}

// readSVGAttrs consumes the attributes of the tag the lexer stands inside,
// returning them with the token that closed the tag. A valueless attribute is
// skipped rather than ending the run, so the walk stays in step with the token
// stream and the spans it derives from it stay true.
func readSVGAttrs(l *xml.Lexer) (xml.TokenType, []svgAttr) {
	var attrs []svgAttr

	for {
		tt, _ := l.Next()
		if tt != xml.AttributeToken {
			return tt, attrs
		}

		val := l.AttrVal()
		if len(val) < 2 {
			continue
		}

		attrs = append(attrs, svgAttr{name: string(l.Text()), value: string(val[1 : len(val)-1])})
	}
}
