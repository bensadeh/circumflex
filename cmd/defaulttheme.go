package cmd

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/theme"

	"github.com/spf13/cobra"
)

func defaultThemeCmd() *cobra.Command {
	path := settings.ThemePath()

	return &cobra.Command{
		Use:                   "default-theme",
		Short:                 "Write default theme config to " + settings.Tilde(path),
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := theme.WriteDefaultConfig(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Default theme written to %s\n", path)
		},
	}
}
