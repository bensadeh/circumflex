package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v3"
)

// lispLexers names the chroma lexers whose grammar is s-expressions. They tag
// the same code differently — Clojure marks every head symbol NameFunction
// where Emacs Lisp marks ordinary symbols NameVariable, and a :keyword is a
// builtin to one lexer, a symbol literal to another, a plain variable to a
// third — so retypeLisp settles them onto one scheme before anything is
// colored.
var lispLexers = map[string]struct{}{
	"EmacsLisp":   {},
	"Common Lisp": {},
	"Scheme":      {},
	"Clojure":     {},
	"Racket":      {},
	"NewLisp":     {},
	"Fennel":      {},
}

// retypeLisp rebalances a Lisp token stream. Left alone it paints a block one
// hue: every identifier in the language is a symbol, the lexers tag symbols
// NameVariable, and the shared map colors that like a shell $var. Plain is the
// right default for a symbol, and the color belongs on what a Lisp reader
// scans for instead — the special form opening each expression, the keywords,
// the name a definition introduces, and the quoted data.
func retypeLisp(tokens []chroma.Token) {
	for i := range tokens {
		tok := &tokens[i]

		switch {
		// A :keyword is one self-evaluating constant whichever lexer read it.
		case isLispKeyword(tok.Value) && !tok.Type.InCategory(chroma.Comment) &&
			tok.Type != chroma.LiteralString:
			tok.Type = chroma.NameConstant

		// Ordinary symbols. The lexers reserve NameFunction and NameBuiltin for
		// vocabulary they actually recognize, so what is left is user naming.
		case tok.Type == chroma.NameVariable:
			tok.Type = chroma.Name

		// Special forms and macros — setq, defun, use-package. Emacs Lisp files
		// them beside its builtin functions, which stay NameFunction.
		case tok.Type == chroma.NameBuiltin:
			tok.Type = chroma.Keyword

		// Lambda-list markers — &rest, &optional, and Clojure's bare & — bound
		// a parameter list rather than open a form, and Emacs gives them the
		// type face for it.
		case strings.HasPrefix(tok.Value, "&") &&
			(tok.Type == chroma.Operator || tok.Type == chroma.KeywordPseudo):
			tok.Type = chroma.NameClass

		// The cons dot and the quoting marks are structure, not the escape
		// hatches the shared map paints red.
		case tok.Type == chroma.Operator:
			tok.Type = chroma.Punctuation
		}
	}

	markModernForms(tokens)
	markLispDefinitions(tokens)
}

// markModernForms supplies the macros chroma's Emacs Lisp vocabulary predates.
// The list is deliberately confined to forms Emacs itself gives the keyword
// face: functions the lexer also omits — add-hook, add-to-list, advice-add —
// are left plain on purpose, because that is how Emacs renders them too.
func markModernForms(tokens []chroma.Token) {
	for i := range tokens {
		tok := &tokens[i]

		if tok.Type != chroma.Name && tok.Type != chroma.NameVariable {
			continue
		}

		if _, ok := modernElispForms[strings.TrimSpace(tok.Value)]; ok {
			tok.Type = chroma.Keyword
		}
	}
}

var modernElispForms = map[string]struct{}{
	"setopt": {}, "setq-local": {}, "named-let": {}, "while-let": {},
	"when-let": {}, "when-let*": {}, "if-let": {}, "if-let*": {}, "and-let*": {},
	"thread-first": {}, "thread-last": {}, "with-suppressed-warnings": {},
}

// isLispKeyword reports a self-evaluating :keyword. A string holding one keeps
// its quotes in the token value, so the colon never leads there.
func isLispKeyword(value string) bool {
	return len(value) > 1 && value[0] == ':' && isASCIILetter(value[1])
}

// markLispDefinitions colors the symbol a definition introduces — the one name
// in the form a reader is looking for. No lexer in the family marks it: Emacs
// Lisp tags the defun's name exactly as it tags every call site, so the head of
// the form is the only thing that can say the next symbol is being defined.
// Scheme writes the name inside a nested paren, hence skipping punctuation.
func markLispDefinitions(tokens []chroma.Token) {
	defining := false

	for i := range tokens {
		tok := &tokens[i]

		if tok.Type == chroma.Text || tok.Type == chroma.TextWhitespace ||
			tok.Type == chroma.Punctuation || tok.Type.InCategory(chroma.Comment) {
			continue
		}

		if defining {
			if tok.Type == chroma.Name || tok.Type == chroma.NameFunction {
				tok.Type = chroma.NameFunction
			}

			defining = false

			continue
		}

		// Clojure's lexer carries the trailing space into the token value.
		_, defining = lispDefiners[strings.TrimSpace(tok.Value)]
	}
}

// lispDefiners is an explicit set rather than a def prefix test: an ordinary
// symbol like default-directory would otherwise hand the function color to
// whatever symbol followed it.
var lispDefiners = map[string]struct{}{
	"defun": {}, "defsubst": {}, "defmacro": {}, "defvar": {}, "defvar-local": {},
	"defconst": {}, "defcustom": {}, "defface": {}, "defgroup": {},
	"define-minor-mode": {}, "define-derived-mode": {}, "cl-defun": {},
	"cl-defmacro": {}, "cl-defmethod": {}, "cl-defstruct": {},

	"defparameter": {}, "defconstant": {}, "defclass": {}, "defmethod": {},
	"defgeneric": {}, "defstruct": {}, "defpackage": {}, "define-condition": {},

	"define": {}, "define-syntax": {}, "define-record-type": {}, "define-values": {},

	"defn": {}, "defn-": {}, "defrecord": {}, "defprotocol": {}, "defmulti": {},
	"defmethod-": {}, "defmacro-": {}, "deftype": {},
}
