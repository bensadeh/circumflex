package settings

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/graphics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestLoadConfig_MissingFileIsEmpty(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "config.toml"))
	require.NoError(t, err)
	assert.Equal(t, &FileConfig{}, cfg)
}

func TestLoadConfig_AllKeys(t *testing.T) {
	path := writeConfigFile(t, `
comment_width = 60
article_width = 90
indent = 2
history = false
nerdfonts = true
graphics = "always"
show_images_on_open = true
categories = ["top", "new"]
pages = 2
wide_view = "always"
`)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	config := Default()
	require.NoError(t, cfg.Apply(config))

	assert.Equal(t, 60, config.CommentWidth)
	assert.Equal(t, 90, config.ArticleWidth)
	assert.Equal(t, 2, config.Indent)
	assert.True(t, config.DoNotMarkSubmissionsAsRead)
	assert.True(t, config.EnableNerdFonts)
	assert.True(t, config.ShowImagesOnOpen)
	assert.Equal(t, graphics.ModeAlways, config.Graphics)
	assert.Equal(t, "top,new", config.Categories)
	assert.Equal(t, 2, config.PageMultiplier)
	assert.Equal(t, 0, config.WideViewMinWidth)
}

func TestLoadConfig_UnknownKeyFails(t *testing.T) {
	path := writeConfigFile(t, "image = true\n")

	_, err := LoadConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image")
}

func TestApply_EmptyFileKeepsDefaults(t *testing.T) {
	config := Default()
	require.NoError(t, (&FileConfig{}).Apply(config))
	assert.Equal(t, Default(), config)
}

func TestApply_ClampsLikeTheFlags(t *testing.T) {
	cfg, err := LoadConfig(writeConfigFile(t, "pages = 99\nindent = 0\ncomment_width = -1\narticle_width = 0\n"))
	require.NoError(t, err)

	config := Default()
	require.NoError(t, cfg.Apply(config))
	assert.Equal(t, 5, config.PageMultiplier)
	assert.Equal(t, 1, config.Indent)
	assert.Equal(t, 1, config.CommentWidth)
	assert.Equal(t, 1, config.ArticleWidth)
}

func TestApply_WideView(t *testing.T) {
	tests := []struct {
		toml string
		want int
	}{
		{`wide_view = 120`, 120},
		{`wide_view = "120"`, 120},
		{`wide_view = "always"`, 0},
		{`wide_view = "never"`, math.MaxInt},
	}

	for _, tt := range tests {
		cfg, err := LoadConfig(writeConfigFile(t, tt.toml))
		require.NoError(t, err, tt.toml)

		config := Default()
		require.NoError(t, cfg.Apply(config), tt.toml)
		assert.Equal(t, tt.want, config.WideViewMinWidth, tt.toml)
	}
}

func TestApply_WideViewInvalid(t *testing.T) {
	for _, content := range []string{
		`wide_view = 0`,
		`wide_view = "sometimes"`,
		`wide_view = true`,
	} {
		cfg, err := LoadConfig(writeConfigFile(t, content))
		require.NoError(t, err, content)

		err = cfg.Apply(Default())
		require.Error(t, err, content)
		assert.Contains(t, err.Error(), "wide_view", content)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	require.NoError(t, WriteDefaultConfig(path))
	require.Error(t, WriteDefaultConfig(path), "refuses to overwrite an existing config")

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, &FileConfig{}, cfg, "every key in the default config should be commented out")
}

// Uncommenting every setting line in the template must yield a file where
// each FileConfig field is set, so the template can't drift from the schema.
func TestDefaultConfig_TemplateMatchesSchema(t *testing.T) {
	var lines []string

	for line := range strings.SplitSeq(defaultConfigBody(), "\n") {
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ") {
			line = strings.TrimPrefix(line, "#")
		}

		lines = append(lines, line)
	}

	cfg, err := LoadConfig(writeConfigFile(t, strings.Join(lines, "\n")))
	require.NoError(t, err)

	assert.NotNil(t, cfg.CommentWidth)
	assert.NotNil(t, cfg.ArticleWidth)
	assert.NotNil(t, cfg.Indent)
	assert.NotNil(t, cfg.History)
	assert.NotNil(t, cfg.NerdFonts)
	assert.NotNil(t, cfg.Graphics)
	assert.NotNil(t, cfg.ShowImagesOnOpen)
	assert.NotEmpty(t, cfg.Categories)
	assert.NotNil(t, cfg.Pages)
	assert.NotNil(t, cfg.WideView)

	require.NoError(t, cfg.Apply(Default()), "uncommented defaults should be valid values")
}
