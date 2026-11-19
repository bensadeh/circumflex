package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCenterComments_FlagOverridesConfig(t *testing.T) {
	xdgConfigHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	configDir := filepath.Join(xdgConfigHome, "circumflex")
	require.NoError(t, os.MkdirAll(configDir, 0o700))
	configPath := filepath.Join(configDir, "config.toml")

	for _, tt := range []struct {
		name       string
		config     bool
		flag       bool
		wantCenter bool
	}{
		{name: "false flag overrides true config", config: true, flag: false, wantCenter: false},
		{name: "true flag overrides false config", config: false, flag: true, wantCenter: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(configPath,
				[]byte(fmt.Sprintf("center_comments = %t\n", tt.config)), 0o600))

			root := Root()
			require.NoError(t, root.ParseFlags([]string{fmt.Sprintf("--center-comments=%t", tt.flag)}))
			currentCmd = root

			t.Cleanup(func() { currentCmd = nil })

			config, err := getConfig()
			require.NoError(t, err)
			require.Equal(t, tt.wantCenter, config.CenterComments)
		})
	}
}
