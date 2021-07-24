package cmd

import (
	"clx/clx"
	clx2 "clx/constants/clx"
	"clx/settings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	PlainHeadlines bool
	PlainComments  bool
)

var rootCmd = &cobra.Command{
	Use:   "clx",
	Short: "It's Hacker News in your terminal",
	Long:  "circumflex " + clx2.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if PlainHeadlines {
			viper.Set(settings.HighlightHeadlinesKey, false)
		}

		if PlainComments {
			viper.Set(settings.HighlightHeadlinesKey, false)
		}

		clx.Run()
	},
}

func Execute() error {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVarP(&PlainHeadlines, "plain-headlines", "l", false,
		"disable syntax highlighting for headlines")
	rootCmd.PersistentFlags().BoolVarP(&PlainComments, "plain-comments", "c", false,
		"disable syntax highlighting for comments")

	return rootCmd.Execute()
}
