package cmd

import (
	"clx/file"
	"clx/settings"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "create_example_config",
	Short: "Create a template config file",
	Long:  `Create a template config file in ~/.config/circumflex/config.env`,
	Run: func(cmd *cobra.Command, args []string) {
		err := file.WriteToFile(file.PathToConfigFile(), settings.GetConfigFileContents())
		if err != nil {
			fmt.Println(err)
		}

		os.Exit(0)
	},
}
