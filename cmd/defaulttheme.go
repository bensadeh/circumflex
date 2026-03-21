package cmd

import (
	"clx/theme"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func defaultThemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "default-theme",
		Short:                 "Write default theme config to ~/.config/circumflex/theme.toml",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			path, err := theme.WriteDefaultConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Default theme written to %s\n", path)
		},
	}
}
