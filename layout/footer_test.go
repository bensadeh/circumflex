package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFooterSections(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ab              cd", FooterSections(18, "ab", "cd"),
		"first flush left, last flush right")

	assert.Equal(t, "ab     mid     cd", FooterSections(17, "ab", "mid", "cd"),
		"slack shared equally between the gaps")

	assert.Equal(t, "abcdef ghijkl", FooterSections(10, "abcdef", "ghijkl"),
		"overrunning labels keep a single space")

	assert.Equal(t, "ab              cd", FooterSections(18, "ab", "", "cd"),
		"empty sections are skipped")

	assert.Equal(t, "  \x1b[1mab\x1b[0m            cd", FooterSections(18, "  \x1b[1mab\x1b[0m", "cd"),
		"labels are measured by display width")
}
