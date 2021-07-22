package cmd

import (
	"clx/file"
	"clx/settings"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createConfigCmd)
}

var createConfigCmd = &cobra.Command{
	Use:   "create_config",
	Short: "Create an example config file",
	Long: "Create an example config file in ~/.config/circumflex/config.env.\n" +
		"If a config file already exists, it will be overwritten.",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := file.WriteToFileInConfigDir(file.PathToConfigFile(), settings.GetConfigFileContents())
		if err != nil {
			fmt.Println(err)

			os.Exit(1)
		}

		fmt.Println("Example config file written to ~/.config/circumflex/config.env")

		os.Exit(0)
	},
}
