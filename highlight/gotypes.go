package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v3"
)

// enrichGoTypes colors the type positions Go's grammar pins down at token
// level. Go can't join capitalizedTypeLangs — every exported identifier is
// capitalized, so fields and calls would take the type hue wholesale — but a
// few positions are unambiguous in gofmt'd code: the name declared after the
// type keyword, the element type behind map and chan, and the sigil-glued
// expression a declared name or closing paren introduces (Fset *token.FileSet,
// Requires []*Analyzer, func f() *User — two adjacent identifiers only ever
// mean declaration-then-type). Only capitalized segments promote, keeping
// package qualifiers (token.Pos) and lowercase locals plain, and the glue
// rules leave expressions alone: a * b keeps its spaces where a type's sigils
// never have any, and a deref's * follows an operator or keyword, never a
// name. Composite literals, conversions and func-type parameters stay plain —
// telling those apart needs a parser, and a wrongly typed name is worse than
// an uncolored one.
func enrichGoTypes(tokens []chroma.Token) {
	sig := significantIndices(tokens)

	for j := 0; j < len(sig); j++ {
		tok := tokens[sig[j]]

		switch {
		case isKeywordValue(tok, "type"):
			// The declared name is the one promotion that skips the
			// capitalization gate: type isWrapper struct{} names a type too.
			if j+1 < len(sig) && tokens[sig[j+1]].Type == chroma.NameOther &&
				spacedSameLine(tokens, sig[j], sig[j+1]) {
				tokens[sig[j+1]].Type = chroma.NameClass
				j = walkTypeExpr(tokens, sig, j+2, sig[j+1]) - 1
			}

		case isKeywordValue(tok, "map") || isKeywordValue(tok, "chan"):
			j = walkTypeExpr(tokens, sig, j, -1) - 1

		case startsTypeExpr(tok) && j > 0 && typeContextBefore(tokens, sig, j):
			j = walkTypeExpr(tokens, sig, j, -1) - 1
		}
	}
}

// walkTypeExpr consumes a glued type expression from sig[j], promoting its
// capitalized names, and returns the first index it declined. prev is the raw
// index of the last consumed token, or -1 when the trigger already vouched
// for the first one; chan is the one keyword a space may follow.
func walkTypeExpr(tokens []chroma.Token, sig []int, j, prev int) int {
	for j < len(sig) {
		idx := sig[j]

		if prev >= 0 && idx != prev+1 && !chanGap(tokens, prev, idx) {
			return j
		}

		tok := &tokens[idx]

		switch {
		case tok.Type == chroma.Punctuation && (bracketsOnly(tok.Value) || tok.Value == "."):
		case tok.Type == chroma.Operator && (tok.Value == "*" || tok.Value == "..."):
		case isKeywordValue(*tok, "map") || isKeywordValue(*tok, "chan"):
		case tok.Type == chroma.KeywordType:
		case tok.Type.InCategory(chroma.LiteralNumber): // array lengths: [3]Vec
		case tok.Type == chroma.NameOther:
			if startsUpper(tok.Value) {
				tok.Type = chroma.NameClass
			}
		default:
			return j
		}

		prev = idx
		j++
	}

	return j
}

func startsTypeExpr(t chroma.Token) bool {
	return t.Type == chroma.NameOther ||
		(t.Type == chroma.Operator && (t.Value == "*" || t.Value == "...")) ||
		(t.Type == chroma.Punctuation && bracketsOnly(t.Value))
}

// typeContextBefore reports the positions whose next word opens a type: a
// preceding identifier (field, parameter, receiver, var declaration) or a
// closing paren (return type). The whitespace requirement is load-bearing
// both ways — two identifiers touch in no expression, and unformatted a*b
// arrives glued where a declaration's name and type never do.
func typeContextBefore(tokens []chroma.Token, sig []int, j int) bool {
	prev := tokens[sig[j-1]]

	if !spacedSameLine(tokens, sig[j-1], sig[j]) {
		return false
	}

	return prev.Type == chroma.NameOther ||
		(prev.Type == chroma.Punctuation && strings.HasSuffix(prev.Value, ")"))
}

func isKeywordValue(t chroma.Token, value string) bool {
	return t.Type.InCategory(chroma.Keyword) && t.Value == value
}

func bracketsOnly(s string) bool {
	return s != "" && strings.TrimLeft(s, "[]") == ""
}

// spacedSameLine reports whitespace, but no line break, between two raw
// token indices.
func spacedSameLine(tokens []chroma.Token, a, b int) bool {
	if b <= a+1 {
		return false
	}

	for _, t := range tokens[a+1 : b] {
		if strings.Contains(t.Value, "\n") {
			return false
		}
	}

	return true
}

// chanGap allows the single space chan writes before its element type.
func chanGap(tokens []chroma.Token, prev, idx int) bool {
	return isKeywordValue(tokens[prev], "chan") && idx == prev+2 &&
		tokens[prev+1].Type == chroma.TextWhitespace &&
		!strings.Contains(tokens[prev+1].Value, "\n")
}
