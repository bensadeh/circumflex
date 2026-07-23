package highlight

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v3"
	"github.com/alecthomas/chroma/v3/lexers"

	"github.com/bensadeh/circumflex/style"
)

// capitalizedTypeLangs names languages whose naming convention makes a
// capitalized identifier a type, trait, or constructor, but whose lexer
// emits them as plain Names. Haskell shares the convention yet needs no
// entry — its lexer already tags them; C# is excluded because it
// capitalizes methods too.
var capitalizedTypeLangs = map[string]struct{}{
	"Rust":   {},
	"Java":   {},
	"Kotlin": {},
}

// allCapsConstLangs names languages where an ALL_CAPS identifier is a
// constant by convention: C's macros, PEP 8, and the Rust/Java/Kotlin style
// guides. JavaScript is excluded — JSON, URL and friends collide.
var allCapsConstLangs = map[string]struct{}{
	"C":      {},
	"C++":    {},
	"Python": {},
	"Rust":   {},
	"Java":   {},
	"Kotlin": {},
}

// resolve looks lang up by lexer name or alias only. chroma's Get also
// falls back to matching "filename.<lang>" against extension globs, which
// routes declared junk to confidently wrong lexers (mod → AMPL, es →
// Erlang).
func resolve(lang string) chroma.Lexer {
	if lang == "" {
		return nil
	}

	lexer := lexers.Get(lang)
	if lexer == nil {
		return nil
	}

	cfg := lexer.Config()
	if strings.EqualFold(cfg.Name, lang) {
		return lexer
	}

	for _, alias := range cfg.Aliases {
		if strings.EqualFold(alias, lang) {
			return lexer
		}
	}

	return nil
}

// Resolves reports whether lang names a real lexer, so declared-language
// extraction can reject page noise before it suppresses the guesser.
func Resolves(lang string) bool { return resolve(lang) != nil }

// Code renders source through chroma when lang names a lexer, or returns ""
// so the caller falls back to its unhighlighted treatment.
func Code(text, lang string) string {
	lexer := resolve(lang)
	if lexer == nil {
		return ""
	}

	iterated, err := chroma.Coalesce(lexer).Tokenise(nil, text)
	if err != nil {
		return ""
	}

	tokens := slices.Collect(iterated)

	name := lexer.Config().Name
	_, capTypes := capitalizedTypeLangs[name]
	_, capsConsts := allCapsConstLangs[name]

	tokens = retypeByLanguage(tokens, name)

	var retypeTOML func(*chroma.Token)
	if name == "TOML" {
		retypeTOML = tomlRetyper()
	}

	// Java and Kotlin capitalize acronym classes (URL, UUID, IO); their
	// true constants always carry an underscore.
	capsNeedUnderscore := name == "Java" || name == "Kotlin"

	enrichCoarseNames(tokens)

	var sb strings.Builder

	for _, token := range tokens {
		if retypeTOML != nil {
			retypeTOML(&token)
		}

		// NameBuiltin joins Name here so Rust's std types (String, Vec,
		// Option, Box) take the type hue like a primitive str does, instead
		// of the builtin/literal one — chroma pre-tags them where a
		// user-defined type would arrive as a plain Name. Java and Kotlin
		// emit no NameBuiltin, so this is Rust in practice; its variant
		// constructors (Some, Ok) ride along as capitalized names.
		if token.Type == chroma.Name || token.Type == chroma.NameBuiltin {
			switch {
			case token.Type == chroma.Name && capsConsts && isAllCaps(token.Value) &&
				(!capsNeedUnderscore || strings.Contains(token.Value, "_")):
				token.Type = chroma.NameConstant

			case capTypes && startsUpper(token.Value):
				token.Type = chroma.NameClass
			}
		}

		// Rust lifetimes give them the type hue like the types they bound,
		// not the attribute/function or builtin one. Named lifetimes ('a)
		// arrive as NameAttribute; 'static arrives as NameBuiltin. A leading
		// apostrophe is theirs alone — char literals lex as a string.
		if strings.HasPrefix(token.Value, "'") &&
			(token.Type == chroma.NameAttribute || token.Type == chroma.NameBuiltin) {
			token.Type = chroma.NameClass
		}

		// The reference & before a lifetime is the one sigil chroma tags
		// KeywordPseudo rather than Operator; fold it back so every & and *
		// takes the operator color with the rest.
		if token.Type == chroma.KeywordPseudo && isRefSigil(token.Value) {
			token.Type = chroma.Operator
		}

		sb.WriteString(styleToken(token))
	}

	// Lexers append a trailing newline the source never had.
	return strings.TrimRight(sb.String(), "\n")
}

// retypeByLanguage runs the passes that only make sense for one language,
// before the shared rules that follow apply to every token stream alike.
func retypeByLanguage(tokens []chroma.Token, name string) []chroma.Token {
	if _, ok := lispLexers[name]; ok {
		retypeLisp(tokens)

		return tokens
	}

	if name == "Bash Session" {
		tokens = relexContinuations(tokens)
	}

	if name == "Bash" || name == "Bash Session" {
		tokens = splitShellFlags(tokens)

		// The \ line continuation is shell machinery, cyan like builtins
		// and $vars — not an escape hatch; in-string escapes keep red.
		for i, token := range tokens {
			if token.Type == chroma.LiteralStringEscape && token.Value == "\\\n" {
				tokens[i].Type = chroma.NameBuiltin
			}
		}
	}

	return tokens
}

// styleToken styles each line of a token separately: the rounded box splices
// border cells between lines, so a style spanning a newline would bleed into
// the frame and lose its color on the continuation line.
func styleToken(token chroma.Token) string {
	styleFn := tokenStyle(token.Type)
	if isCommentProse(token.Type) {
		styleFn = styleCommentText
	}

	if styleFn == nil {
		return token.Value
	}

	parts := strings.Split(token.Value, "\n")
	for i, part := range parts {
		if part != "" {
			parts[i] = styleFn(part)
		}
	}

	return strings.Join(parts, "\n")
}

// tokenStyles maps token types to colors, from categories down to exact
// types — chroma.Lookup falls back exact → sub-category → category, so the
// specific entries override their parents. Anything unlisted (operators,
// punctuation, plain names, program output) renders in the terminal's
// default text style on purpose.
var tokenStyles = map[chroma.TokenType]func(string) string{
	chroma.Comment:       style.CodeComment,
	chroma.Keyword:       style.CodeKeyword,
	chroma.LiteralString: style.CodeString,
	chroma.Literal:       style.CodeLiteral,

	// Preprocessor directives are typed as comments but read as keywords.
	chroma.CommentPreproc:     style.CodeKeyword,
	chroma.CommentPreprocFile: style.CodeString,

	chroma.OperatorWord: style.CodeKeyword, // is, not, and
	chroma.KeywordType:  style.CodeKeyword, // primitives (str, i32, int) are builtin vocabulary

	// Type hue is left for named types — user structs and the capitalized
	// std types promoted into NameClass; a primitive reads as a keyword.
	chroma.NameClass: style.CodeType,

	chroma.NameFunction:      style.CodeFunction,
	chroma.NameFunctionMagic: style.CodeFunction, // println!, vec!
	chroma.NameDecorator:     style.CodeFunction,

	// Red marks escape hatches and sharp edges: operators including the &
	// and * sigils, escapes and interpolation standing apart inside the
	// string color, regexes out of plain strings, entities as html's escape
	// mechanism, exception names on error paths. Word operators (is, not,
	// and) keep the keyword color; punctuation stays plain. Lexer Error
	// tokens deliberately stay plain, red would read as broken.
	chroma.Operator:              style.CodeEscape,
	chroma.LiteralStringEscape:   style.CodeEscape,
	chroma.LiteralStringInterpol: style.CodeEscape,
	chroma.LiteralStringRegex:    style.CodeEscape,
	chroma.NameEntity:            style.CodeEscape,
	chroma.NameException:         style.CodeEscape,

	chroma.NameTag:              style.CodeKeyword,  // html tags, json/yaml keys, toml headers
	chroma.NameAttribute:        style.CodeFunction, // html attributes, ini/properties/toml keys, shell flags
	chroma.NameBuiltin:          style.CodeLiteral,
	chroma.NameConstant:         style.CodeLiteral,
	chroma.NameVariable:         style.CodeLiteral,
	chroma.NameVariableClass:    style.CodeLiteral,
	chroma.NameVariableGlobal:   style.CodeLiteral,
	chroma.NameVariableInstance: style.CodeLiteral,

	chroma.GenericPrompt:     style.CodeKeyword, // console $, output stays plain
	chroma.GenericInserted:   style.CodeString,
	chroma.GenericDeleted:    style.CodeDeleted,
	chroma.GenericHeading:    style.Faint,
	chroma.GenericSubheading: style.Faint,
}

// relexContinuations repairs the console lexer's handling of \-continued
// commands: it flips to output state at the first newline, so the later
// lines of the command lex as plain GenericOutput. Those lines re-lex with
// the bash lexer to keep their flags, strings and escapes colored.
func relexContinuations(tokens []chroma.Token) []chroma.Token {
	bash := lexers.Get("bash")
	if bash == nil {
		return tokens
	}

	out := make([]chroma.Token, 0, len(tokens))

	for i, token := range tokens {
		continued := i > 0 && tokens[i-1].Type == chroma.LiteralStringEscape &&
			strings.HasSuffix(tokens[i-1].Value, "\\\n")
		if token.Type != chroma.GenericOutput || !continued {
			out = append(out, token)

			continue
		}

		lines := strings.SplitAfter(token.Value, "\n")

		cut := 0
		for cut < len(lines) {
			line := strings.TrimSuffix(lines[cut], "\n")
			cut++

			if !strings.HasSuffix(line, `\`) {
				break
			}
		}

		command, err := chroma.Coalesce(bash).Tokenise(nil, strings.Join(lines[:cut], ""))
		if err != nil {
			out = append(out, token)

			continue
		}

		out = slices.AppendSeq(out, command)

		if rest := strings.Join(lines[cut:], ""); rest != "" {
			out = append(out, chroma.Token{Type: chroma.GenericOutput, Value: rest})
		}
	}

	return out
}

// shellFlag matches -v and --verbose words at word start, so the dashes in
// URLs, dates and file names stay plain. The bare -- end-of-options
// separator counts too; a bare - (stdin placeholder) does not.
var shellFlag = regexp.MustCompile(`(?:^|\s)(--?[A-Za-z0-9][A-Za-z0-9_-]*|--)`)

// splitShellFlags carves flag words out of shell Text tokens — the lexer
// leaves command arguments undifferentiated, but flags name parameters, so
// they take the key hue via NameAttribute.
func splitShellFlags(tokens []chroma.Token) []chroma.Token {
	out := make([]chroma.Token, 0, len(tokens))

	for _, token := range tokens {
		if token.Type != chroma.Text {
			out = append(out, token)

			continue
		}

		pos := 0

		for _, m := range shellFlag.FindAllStringSubmatchIndex(token.Value, -1) {
			start, end := m[2], m[3]
			if start > pos {
				out = append(out, chroma.Token{Type: chroma.Text, Value: token.Value[pos:start]})
			}

			out = append(out, chroma.Token{Type: chroma.NameAttribute, Value: token.Value[start:end]})
			pos = end
		}

		if pos < len(token.Value) {
			out = append(out, chroma.Token{Type: chroma.Text, Value: token.Value[pos:]})
		}
	}

	return out
}

// tomlRetyper returns a stateful retyper for TOML's undifferentiated
// NameOther tokens: a word inside line-leading [ ] brackets names a table,
// any other bare word is a key — valid TOML never bares a word in value
// position. The lexer's flat grammar leaves all of them as NameOther, and
// other lexers use NameOther for plain identifiers, hence the scoping.
func tomlRetyper() func(*chroma.Token) {
	lineStart, inHeader := true, false

	return func(token *chroma.Token) {
		switch {
		case token.Type == chroma.Punctuation && lineStart && strings.HasPrefix(token.Value, "["):
			inHeader = true

		case token.Type == chroma.Punctuation && inHeader && strings.Contains(token.Value, "]"):
			inHeader = false

		case token.Type == chroma.NameOther && inHeader:
			token.Type = chroma.NameTag

		case token.Type == chroma.NameOther:
			token.Type = chroma.NameAttribute
		}

		switch {
		case token.Type == chroma.Text && strings.Contains(token.Value, "\n"):
			lineStart, inHeader = true, false
		case token.Type != chroma.Text:
			lineStart = false
		}
	}
}

func tokenStyle(t chroma.TokenType) func(string) string {
	fn, _ := chroma.Lookup(tokenStyles, t)

	return fn
}

// enrichCoarseNames fills in roles that coarse lexers (JavaScript,
// TypeScript, Go) leave as undifferentiated NameOther — the token chroma
// emits when it declines to classify a name. It reads only syntactic
// adjacency, never guesses, and touches NameOther alone: a lexer that
// tagged a name Name/NameFunction/NameBuiltin already decided, so Rust and
// friends pass through unchanged.
//
//   - a name immediately before ( is a call or definition -> NameFunction
//   - a name in { } member position before : is a key -> NameTag, the same
//     hue json and yaml keys already take
//
// The bracket stack is what keeps the key rule honest: it fires only with
// an open { on top and a preceding { or , , so a ternary branch (preceded
// by ?) or a typed parameter (inside ( ) never qualifies.
func enrichCoarseNames(tokens []chroma.Token) {
	sig := make([]int, 0, len(tokens))

	for i, t := range tokens {
		if t.Type == chroma.Text || t.Type == chroma.TextWhitespace ||
			t.Type.InCategory(chroma.Comment) {
			continue
		}

		sig = append(sig, i)
	}

	var stack []byte

	for j, idx := range sig {
		tok := tokens[idx]

		if tok.Type == chroma.NameOther {
			var next chroma.Token
			if j+1 < len(sig) {
				next = tokens[sig[j+1]]
			}

			switch {
			case strings.HasPrefix(next.Value, "("):
				tokens[idx].Type = chroma.NameFunction

			case next.Value == ":" && topIs(stack, '{') && j > 0 &&
				endsWithMemberSep(tokens[sig[j-1]].Value):
				tokens[idx].Type = chroma.NameTag
			}
		}

		if tok.Type == chroma.Punctuation {
			for _, c := range tok.Value {
				switch c {
				case '{', '(', '[':
					stack = append(stack, byte(c))
				case '}', ')', ']':
					if n := len(stack); n > 0 {
						stack = stack[:n-1]
					}
				}
			}
		}
	}
}

func topIs(stack []byte, b byte) bool {
	return len(stack) > 0 && stack[len(stack)-1] == b
}

func endsWithMemberSep(s string) bool {
	if s == "" {
		return false
	}

	c := s[len(s)-1]

	return c == '{' || c == ','
}

// isRefSigil reports an operator built only from & and *, the sigils of
// Rust's borrows, raw pointers and dereferences (&, *, &*, &&). Mixed
// operators like *= and arithmetic never qualify.
func isRefSigil(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r != '&' && r != '*' {
			return false
		}
	}

	return true
}

func startsUpper(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)

	return unicode.IsUpper(r)
}

func isAllCaps(s string) bool {
	return utf8.RuneCountInString(s) >= 2 && s == strings.ToUpper(s) && s != strings.ToLower(s)
}

func isCommentProse(t chroma.TokenType) bool {
	return t.InCategory(chroma.Comment) &&
		t != chroma.CommentPreproc && t != chroma.CommentPreprocFile
}

var todoMarker = regexp.MustCompile(`\b(?:TODO|FIXME|XXX)\b`)

// styleCommentText paints debt markers red inside otherwise-quiet comments.
// Chroma never emits CommentSpecial for them, so the split happens here.
func styleCommentText(part string) string {
	locs := todoMarker.FindAllStringIndex(part, -1)
	if locs == nil {
		return style.CodeComment(part)
	}

	var b strings.Builder

	pos := 0

	for _, loc := range locs {
		if loc[0] > pos {
			b.WriteString(style.CodeComment(part[pos:loc[0]]))
		}

		b.WriteString(style.CodeEscape(part[loc[0]:loc[1]]))
		pos = loc[1]
	}

	if pos < len(part) {
		b.WriteString(style.CodeComment(part[pos:]))
	}

	return b.String()
}
