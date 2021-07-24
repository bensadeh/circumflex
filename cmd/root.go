package cmd

import (
	"clx/clx"
	clx2 "clx/constants/clx"
	"clx/settings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	plainHeadlines bool
	plainComments  bool
	disableHistory bool
	altIndentBlock bool
)

var rootCmd = &cobra.Command{
	Use:   "clx",
	Short: "It's Hacker News in your terminal",
	Long:  "circumflex " + clx2.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if plainHeadlines {
			viper.Set(settings.HighlightHeadlinesKey, false)
		}

		if plainComments {
			viper.Set(settings.CommentHighlightingKey, false)
		}

		if disableHistory {
			viper.Set(settings.MarkAsReadKey, false)
		}

		if altIndentBlock {
			viper.Set(settings.UseAltIndentBlockKey, true)
		}

		clx.Run()
	},
}

func Execute() error {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVarP(&plainHeadlines, "plain-headlines", "l", false,
		"disable syntax highlighting for headlines")
	rootCmd.PersistentFlags().BoolVarP(&plainComments, "plain-comments", "c", false,
		"disable syntax highlighting for comments")
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().BoolVarP(&altIndentBlock, "use-alt-indent-block", "a", false,
		"use alternate indentation block")

	return rootCmd.Execute()
}
