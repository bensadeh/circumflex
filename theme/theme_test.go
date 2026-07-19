package theme

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "named color", input: "magenta", valid: true},
		{name: "bright named color", input: "bright_cyan", valid: true},
		{name: "named color with whitespace", input: " red ", valid: true},
		{name: "ansi 256", input: "219", valid: true},
		{name: "ansi 256 upper bound", input: "255", valid: true},
		{name: "hex six digits", input: "#ff5500", valid: true},
		{name: "hex three digits", input: "#f50", valid: true},
		{name: "empty means terminal default", input: "", valid: true},
		{name: "misspelled name", input: "magneta", valid: false},
		{name: "ansi 256 out of range", input: "256", valid: false},
		{name: "negative number", input: "-1", valid: false},
		{name: "malformed hex", input: "#ff55", valid: false},
		{name: "non-hex characters", input: "#zzzzzz", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, ok := parseColor(tt.input)

			assert.Equal(t, tt.valid, ok)
			assert.Equal(t, c, ParseColor(tt.input))

			if !tt.valid {
				assert.Equal(t, lipgloss.NoColor{}, c)
			}
		})
	}
}

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	t.Parallel()

	loaded, err := Load(filepath.Join(t.TempDir(), "theme.toml"))

	require.NoError(t, err)
	assert.Equal(t, Default(), loaded)
}

func TestLoad_AppliesOverrides(t *testing.T) {
	t.Parallel()

	path := writeTheme(t, `
[headline]
ask_hn = "#ff5500"

[indent]
cycle = ["red", "blue"]
`)

	loaded, err := Load(path)

	require.NoError(t, err)
	assert.Equal(t, "#ff5500", loaded.Headline.AskHN)
	assert.Equal(t, []string{"red", "blue"}, loaded.Indent.Cycle)
	assert.Equal(t, Default().App.Primary, loaded.App.Primary, "unset keys keep their defaults")
}

func TestLoad_CodeGroupOverrides(t *testing.T) {
	t.Parallel()

	path := writeTheme(t, `
[code]
keyword = "bright_red"
`)

	loaded, err := Load(path)

	require.NoError(t, err)
	assert.Equal(t, "bright_red", loaded.Code.Keyword)
	assert.Equal(t, "green", loaded.Code.String, "the other groups keep their defaults")
}

func TestLoad_ReportsUnrecognizedColors(t *testing.T) {
	t.Parallel()

	path := writeTheme(t, `
[headline]
ask_hn = "blu"

[indent]
cycle = ["red", "magneta"]
`)

	_, err := Load(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `headline.ask_hn = "blu"`)
	assert.Contains(t, err.Error(), `indent.cycle = "magneta"`)
}

func TestLoad_RejectsUnknownKeys(t *testing.T) {
	t.Parallel()

	path := writeTheme(t, `
[headline]
ask_hm = "#ff5500"
`)

	_, err := Load(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ask_hm")
}

func TestLoad_DefaultThemeIsValid(t *testing.T) {
	t.Parallel()

	assert.Empty(t, Default().invalidColors())
}

func writeTheme(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "theme.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
