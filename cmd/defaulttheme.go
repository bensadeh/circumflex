package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/theme"

	"github.com/spf13/cobra"
)

func defaultThemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "default-theme",
		Short:                 "write default theme config file",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := settings.ThemePath()

			if err := theme.WriteDefaultConfig(path); err != nil {
				return err
			}

			fmt.Printf("Default theme written to %s\n", path)

			return nil
		},
	}
}
