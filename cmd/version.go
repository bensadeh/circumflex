package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/app"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of circumflex",
		Long:  "Print the version number of circumflex",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(app.Version)
		},
	}
}
