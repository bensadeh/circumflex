package article

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatexToUnicode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{`(-g_x)^c \bmod N = (-1)^c \cdot g_x^c \bmod N.`, "(-𝑔ₓ)ᶜ mod 𝑁 = (-1)ᶜ · 𝑔ₓᶜ mod 𝑁."},
		{`(\mathbb{Z}/n\mathbb{Z})^*`, "(ℤ/𝑛ℤ)*"},
		{`\approx 9.2 \times 10^{18}`, "≈ 9.2 × 10¹⁸"},
		{`\delta \cdot \lfloor 15/8 \rfloor`, "δ · ⌊15/8⌋"},
		{`(-1)^{\text{even}} = 1`, "(-1)ᵉᵛᵉⁿ = 1"},
		{`(-1)^{c} = (-1)^c`, "(-1)ᶜ = (-1)ᶜ"},
		{`\mathsf{pk}_a = g^{\mathsf{sk}_a} - \mathsf{pk}_v`, "pkₐ = 𝑔^(skₐ) - pkᵥ"},
		{`\log_g(-g_x) = \log_h(h_x)`, "log_𝑔(-𝑔ₓ) = logₕ(ℎₓ)"},
		{`\frac{a+b}{2}`, "(𝑎+𝑏)/2"},
		{`\frac{15}{8}`, "15/8"},
		{`\sqrt{2}`, "√2"},
		{`\sqrt{a+b}`, "√(𝑎+𝑏)"},
		{`a \equiv b \pmod{N}`, "𝑎 ≡ 𝑏 (mod 𝑁)"},
		{`\{1, 3, 5\}`, "{1, 3, 5}"},
		{`2^{53}`, "2⁵³"},
		{`\mathcal{O}(n \log n)`, "𝒪(𝑛 log 𝑛)"},
		{`\hat{x} + \vec{v}`, "𝑥̂ + 𝑣⃗"},
		{`\unknowncmd{arg}`, "unknowncmd𝑎𝑟𝑔"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, latexToUnicode(tt.in))
		})
	}
}

func TestConvertMathText_InlineGuards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"latex syntax converts", "solve $x^2 = 4$ now", "solve 𝑥² = 4 now"},
		{"single symbols convert", "when $c$ is even", "when 𝑐 is even"},
		{"bare numbers convert", "a delta of $120$ here", "a delta of 120 here"},
		{"threshold notation", "$t$-of-$n$ sharing", "𝑡-of-𝑛 sharing"},
		{"display math converts", "so $$x = 1$$ holds", "so 𝑥 = 1 holds"},
		{"escaped parens always convert", `bound \(a \le b\) here`, "bound 𝑎 ≤ 𝑏 here"},
		{"prices stay", "was $99 and is now $79 today", "was $99 and is now $79 today"},
		{"price ranges stay", "spend $5-$10 on lunch", "spend $5-$10 on lunch"},
		{"prose between dollars stays", "add $5 and then $10 more", "add $5 and then $10 more"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, convertMathText(tt.in))
		})
	}
}

func TestConvertMath_SkipsCodeSpans(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>run <code>echo $$HOME$$</code> to get $x^2$</p>`)
	convertMath(blocks)

	require.Len(t, blocks, 1)
	assert.Equal(t, "run echo $$HOME$$ to get 𝑥²", blocks[0].plainText(),
		"dollar signs in inline code must survive untouched")
}

func TestUsesMathRenderer(t *testing.T) {
	t.Parallel()

	assert.True(t, usesMathRenderer([]byte(`<script src="/js/MathJax.js"></script>`)))
	assert.True(t, usesMathRenderer([]byte(`<link href="katex.min.css">`)))
	assert.False(t, usesMathRenderer([]byte(`<p>plain page about $5 prices</p>`)))
}
