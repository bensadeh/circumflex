package highlight

import "strings"

// The detectors in guess.go ask whether a token appears somewhere in the text.
// That reaches a language only when it reserves words worth searching for, and
// some don't: every Lisp dialect shares one grammar and almost no keywords, and
// a Nix expression is attribute sets and punctuation nearly all the way down.
// The detectors here read a block's structure instead — how its lines open, how
// its delimiters nest, how its bindings terminate — and corroborate the shape
// with the vocabulary that rides on it.

// sexpOpeners are the characters a line of s-expressions legitimately starts
// with: a form, a closing run left on its own line, a comment, a keyword
// argument, a quoted datum, a vector, or a string continued from the line above.
const sexpOpeners = "()';`:[]\""

// sexpShaped reports whether nearly every non-blank line opens or continues an
// s-expression. Continuation lines are merely indented, so the gate is loose by
// itself — the corroborating signals in isLisp carry the precision.
func sexpShaped(lines []string) bool {
	var total, fits int

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" {
			continue
		}

		total++

		if l[0] == ' ' || l[0] == '\t' || strings.ContainsRune(sexpOpeners, rune(t[0])) {
			fits++
		}
	}

	return total > 0 && fits*10 >= total*9
}

// isLisp requires the shape and an opening form, then wants two corroborating
// signals. Neither half stands alone: a numbered list and a block of
// parenthesized citations both satisfy the shape and open with a paren, and
// nested parens alone are ordinary in every language with function calls.
func isLisp(text string, lines []string) bool {
	if !sexpShaped(lines) || !lispHeadForm(lines) {
		return false
	}

	return atLeastTwo(
		lispQuote(text),
		strings.Contains(text, "))"),
		containsAny(text, lispForms),
		lispKeyword(text),
	)
}

// lispHeadForm reports a line opening a form: a paren followed by a symbol
// name. Required rather than counted, since a block of indented prose in
// parentheses passes the shape gate but never opens one.
func lispHeadForm(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if len(t) > 1 && t[0] == '(' && isASCIILetter(t[1]) {
			return true
		}
	}

	return false
}

// lispQuote reports a quoted symbol or list — 'name, `(form). The quote has to
// open a token and must not close again, which keeps a contraction (don't) and
// a quoted word (the 'foo' option) from reading as a datum.
func lispQuote(text string) bool {
	for i := range len(text) {
		if (text[i] != '\'' && text[i] != '`') || !opensToken(text, i) || i+1 >= len(text) {
			continue
		}

		if text[i+1] == '(' {
			return true
		}

		end := symbolEnd(text, i+1)
		if end > i+1 && (end >= len(text) || (text[end] != '\'' && text[end] != '`')) {
			return true
		}
	}

	return false
}

// lispKeyword reports a self-evaluating keyword in argument position — :ensure,
// (:isHttpEnabled. The colon leads its token, where YAML and JSON trail one
// behind the key instead.
func lispKeyword(text string) bool {
	for i := range len(text) {
		if text[i] != ':' || !opensToken(text, i) {
			continue
		}

		if i+1 < len(text) && isASCIILetter(text[i+1]) {
			return true
		}
	}

	return false
}

func opensToken(text string, i int) bool {
	if i == 0 {
		return true
	}

	switch text[i-1] {
	case ' ', '\t', '\n', '(', '[':
		return true
	default:
		return false
	}
}

func symbolEnd(text string, i int) int {
	for i < len(text) && (isASCIILetter(text[i]) || text[i] == '-' || text[i] == '_' ||
		text[i] == '/' || (text[i] >= '0' && text[i] <= '9')) {
		i++
	}

	return i
}

var lispForms = []string{
	"(defun ", "(defmacro ", "(defvar ", "(defparameter ", "(defconst",
	"(define ", "(defn ", "(defmethod ", "(defclass ", "(defpackage ",
	"(lambda ", "(lambda(", "(setq ", "(setf ", "(let ", "(let*", "(let(",
	"(cond ", "(cond(", "(progn", "(mapcar ", "(require ", "(provide ",
	"(when ", "(unless ", "(cons ", "(apply ", "(funcall ",
}

// lispDialect names the dialect of a block already confirmed as Lisp. chroma
// carries a lexer per dialect and the code box prints its name, so a block that
// shows no dialect resolves to the generic "lisp": the Common Lisp lexer colors
// the grammar they all share, and Label prints "Lisp" rather than claiming a
// dialect the block never evidenced. Common Lisp is tested before Emacs Lisp
// because the two share setq and defun — only the markers below separate them.
func lispDialect(text string) string {
	switch {
	case containsAny(text, clojureMarkers):
		return "clojure"

	case containsAny(text, schemeMarkers):
		return "scheme"

	case containsAny(text, commonLispMarkers):
		return "common-lisp"

	case containsAny(text, emacsLispMarkers):
		return "elisp"

	default:
		return "lisp"
	}
}

// (def is deliberately absent: Arc and other lisps define with it too, and
// a block showing nothing more Clojure than that stays generic.
var clojureMarkers = []string{
	"(defn ", "(defn-", "(ns ", "(fn [", "(let [", "(loop [",
	"(doseq [", "(for [", "(:require", "(defproject", "->>", "#(", "(defrecord ",
}

var schemeMarkers = []string{
	"(define ", "(define-", "(set! ", "(letrec", "(call/cc",
	"(call-with-current-continuation", "(display ", "(newline)", "#t", "#f",
}

var commonLispMarkers = []string{
	"(defpackage ", "(in-package ", "(defparameter ", "(defclass ",
	"(defgeneric ", "(defstruct ", "(defconstant ", "(declaim ",
	"(format t ", "(format nil ", "(loop for ", "(handler-case", "(defvar *",
}

var emacsLispMarkers = []string{
	"(use-package ", "(add-hook ", "(add-to-list ", "(setq-default ", "(setopt ",
	"(defcustom ", "(interactive", "(kbd ", "(require '", "(provide '",
	"(with-eval-after-load ", "(advice-add ", "(global-set-key ", "(define-key ",
	"(defface ", "(defvar-local ", "(defadvice ", "(save-excursion",
	"(eval-after-load", "(message ", "(buffer-", "(point)", "lexical-binding",
	"\"C-c", "\"C-x", "\"M-x", "-mode)", "-hook",
}

// isNix keys on the attribute set: an expression that opens a brace or binds
// through let ... in, with bindings a semicolon terminates. HCL and Gradle also
// assign with =, but neither closes a binding with a semicolon; JSON and
// JavaScript object literals bind with : instead; and a typed declaration
// (int x = 5;) puts a space inside the name, which nixBindings rejects.
func isNix(text string, lines []string) bool {
	trimmed := strings.TrimSpace(text)

	opensAttrSet := trimmed != "" && trimmed[0] == '{'
	if !opensAttrSet && !nixLetIn(lines) {
		return false
	}

	if nixBindings(lines) < 2 {
		return false
	}

	return atLeastTwo(
		strings.Contains(text, "inherit "),
		nixWith(lines),
		strings.Contains(text, "rec {"),
		containsAny(text, nixVocabulary),
		nixLetIn(lines),
	)
}

// nixBindings counts name = value; lines, the attribute set's one universal
// form. The name must be a bare identifier or dotted path — a space inside it
// means a type or a declaration keyword, and the language is not Nix.
func nixBindings(lines []string) int {
	count := 0

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasSuffix(t, ";") {
			continue
		}

		name, _, ok := strings.Cut(t, " = ")
		if !ok || name == "" {
			continue
		}

		if strings.TrimLeft(name, nixNameChars) == "" {
			count++
		}
	}

	return count
}

const nixNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.\"'"

// nixWith reports Nix's with expression — with pkgs; — whose semicolon
// terminator separates it from Python's with block and JavaScript's with
// statement, both of which open a body instead.
func nixWith(lines []string) bool {
	for _, l := range lines {
		_, rest, ok := strings.Cut(l, "with ")
		if !ok {
			continue
		}

		scope, _, ok := strings.Cut(rest, ";")
		if !ok || scope == "" {
			continue
		}

		if strings.TrimLeft(scope, nixNameChars) == "" {
			return true
		}
	}

	return false
}

func nixLetIn(lines []string) bool {
	var hasLet, hasIn bool

	for _, l := range lines {
		t := strings.TrimSpace(l)

		if t == "let" || strings.HasPrefix(t, "let ") {
			hasLet = true
		}

		if t == "in" || strings.HasPrefix(t, "in ") || strings.HasPrefix(t, "in{") {
			hasIn = true
		}
	}

	return hasLet && hasIn
}

var nixVocabulary = []string{
	"mkShell", "mkDerivation", "buildInputs", "nativeBuildInputs", "stdenv",
	"nixpkgs", "pkgs.", "callPackage", "fetchurl", "fetchFromGitHub",
	"devShells", "nixosConfigurations", "homeConfigurations", "flake-utils",
	"writeShellScriptBin", "lib.mkIf", "lib.mkDefault",
}
