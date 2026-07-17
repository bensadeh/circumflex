package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/settings"

	"github.com/spf13/cobra"
)

func defaultConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "default-config",
		Short:                 "write default config file",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := settings.ConfigPath()

			if err := settings.WriteDefaultConfig(path); err != nil {
				return err
			}

			fmt.Printf("Default config written to %s\n", path)

			return nil
		},
	}
}
