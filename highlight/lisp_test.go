package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
)

// The reference for these expectations is the Emacs Lisp in the wild that
// prompted them: an org-exported config post, whose own htmlize spans show
// which symbols Emacs itself fontifies and which it leaves alone.
func TestCode_ElispFollowsEmacsFontLock(t *testing.T) {
	t.Parallel()

	src := "(use-package eglot\n  :ensure nil\n  :hook ((scala-ts-mode . eglot-ensure)))\n" +
		"(setopt eglot-code-action-indications nil) ;; noise\n" +
		"(defun my/fix (orig-fn &rest args)\n  (apply orig-fn args))"

	out := Code(src, "elisp")

	assert.Contains(t, out, style.CodeKeyword("use-package"), "special forms take the keyword hue")
	assert.Contains(t, out, style.CodeKeyword("defun"))
	assert.Contains(t, out, style.CodeKeyword("setopt"), "chroma's vocabulary predates setopt")
	assert.Contains(t, out, style.CodeFunction("my/fix"), "the name a defun introduces")
	assert.Contains(t, out, style.CodeFunction("apply"), "builtin functions stay function-colored")
	assert.Contains(t, out, style.CodeLiteral(":ensure"), "keywords are self-evaluating constants")
	assert.Contains(t, out, style.CodeLiteral("nil"))
	assert.Contains(t, out, style.CodeType("&rest"), "lambda-list markers take Emacs' type face")
	assert.Contains(t, out, style.CodeComment(";; noise"))
}

// Every identifier in a Lisp is a symbol, so coloring symbols paints the whole
// block one hue — the state this scheme replaced.
func TestCode_ElispLeavesOrdinarySymbolsPlain(t *testing.T) {
	t.Parallel()

	out := Code("(use-package eglot\n  :hook ((scala-ts-mode . eglot-ensure)))", "elisp")

	for _, symbol := range []string{"eglot", "scala-ts-mode", "eglot-ensure"} {
		assert.NotContains(t, out, style.CodeLiteral(symbol))
		assert.NotContains(t, out, style.CodeKeyword(symbol))
		assert.NotContains(t, out, style.CodeFunction(symbol))
	}
}

// The cons dot separates a pair; red marks escape hatches, and reading it as
// one made every alist in a config look like an error.
func TestCode_ConsDotIsNotAnEscape(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{"elisp", "lisp", "scheme"} {
		out := Code("(add-to-list 'modes '(scala-ts-mode . eglot-ensure))", lang)

		assert.NotContains(t, out, style.CodeEscape("."), lang)
	}
}

func TestCode_LispDialectsShareOneScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lang    string
		src     string
		defined string
	}{
		{"elisp", "(defun greet (name)\n  (message \"hi %s\" name))", "greet"},
		{"lisp", "(defun greet (name)\n  (format t \"hi ~a\" name))", "greet"},
		{"clojure", "(defn greet [name]\n  (println \"hi\" name))", "greet"},
		{"scheme", "(define (greet name)\n  (display name))", "greet"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()

			out := Code(tt.src, tt.lang)

			assert.Contains(t, out, style.CodeFunction(tt.defined),
				"every dialect names what its definition introduces")
			assert.Equal(t, tt.src, ansi.Strip(out), "coloring never edits the source")
		})
	}
}

// A :keyword reaches the shared map as a builtin, a symbol literal or a plain
// variable depending on which lexer read it; a reader sees one thing.
func TestCode_KeywordsAgreeAcrossDialects(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{"elisp", "lisp", "clojure", "scheme"} {
		out := Code("(config :enabled true :depth 3)", lang)

		assert.Contains(t, out, style.CodeLiteral(":enabled"), lang)
		assert.Contains(t, out, style.CodeLiteral(":depth"), lang)
	}
}

func TestCode_QuotedSymbolsReadAsData(t *testing.T) {
	t.Parallel()

	out := Code("(add-to-list 'eglot-server-programs '(scala-ts-mode))", "elisp")

	assert.Contains(t, out, style.CodeString("'eglot-server-programs"))
}
