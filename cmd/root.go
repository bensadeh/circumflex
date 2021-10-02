package cmd

import (
	"clx/clx"
	clx2 "clx/constants/clx"
	"clx/settings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	plainHeadlines      bool
	plainComments       bool
	disableHistory      bool
	altIndentBlock      bool
	smileyEmojis        bool
	relativeNumbering   bool
	showYCJobs          bool
	preserveRightMargin bool
	hideIndentSymbol    bool
	orangeHeader        bool
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Short:   "It's Hacker News in your terminal",
		Long:    "circumflex is a command line tool for browsing Hacker News in your terminal",
		Version: clx2.Version,
		Run: func(cmd *cobra.Command, args []string) {
			overrideConfig()

			clx.Run()
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	configureFlags(rootCmd)

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(viewCmd())

	return rootCmd
}

func configureFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&plainHeadlines, "plain-headlines", "p", false,
		"disable syntax highlighting for headlines")
	rootCmd.PersistentFlags().BoolVarP(&plainComments, "plain-comments", "o", false,
		"disable syntax highlighting for comments")
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().BoolVarP(&altIndentBlock, "use-alt-indent-block", "a", false,
		"use alternate indentation block")
	rootCmd.PersistentFlags().BoolVarP(&smileyEmojis, "smiley-emojis", "s", false,
		"convert smileys to emojis")
	rootCmd.PersistentFlags().BoolVarP(&relativeNumbering, "relative-numbering", "r", false,
		"use relative numbering for submissions")
	rootCmd.PersistentFlags().BoolVarP(&showYCJobs, "show-jobs", "j", false,
		"show submissions of the type 'X is hiring'")
	rootCmd.PersistentFlags().BoolVarP(&preserveRightMargin, "preserve-right-margin", "m", false,
		"preserve right margin at the cost of comment width")
	rootCmd.PersistentFlags().BoolVarP(&hideIndentSymbol, "hide-indent", "t", false,
		"hide the indentation symbol")
	rootCmd.PersistentFlags().BoolVarP(&orangeHeader, "orange-header", "n", false,
		"set the header to black on orange")

	rootCmd.PersistentFlags().IntP("comment-width", "c", settings.CommentWidthDefault,
		"set the comment width")
	_ = viper.BindPFlag(settings.CommentWidthKey, rootCmd.PersistentFlags().Lookup("comment-width"))

	rootCmd.PersistentFlags().IntP("indent-size", "i", settings.IndentSizeDefault,
		"set the indent size")
	_ = viper.BindPFlag(settings.IndentSizeKey, rootCmd.PersistentFlags().Lookup("indent-size"))
}

func overrideConfig() {
	if plainHeadlines {
		viper.Set(settings.HighlightHeadlinesKey, false)
	}

	if plainComments {
		viper.Set(settings.HighlightCommentsKey, false)
	}

	if disableHistory {
		viper.Set(settings.MarkAsReadKey, false)
	}

	if altIndentBlock {
		viper.Set(settings.UseAltIndentBlockKey, true)
	}

	if smileyEmojis {
		viper.Set(settings.EmojiSmileysKey, true)
	}

	if relativeNumbering {
		viper.Set(settings.RelativeNumberingKey, true)
	}

	if showYCJobs {
		viper.Set(settings.HideYCJobsKey, false)
	}

	if preserveRightMargin {
		viper.Set(settings.PreserveRightMarginKey, true)
	}

	if hideIndentSymbol {
		viper.Set(settings.HideIndentSymbolKey, true)
	}

	if orangeHeader {
		viper.Set(settings.OrangeHeaderKey, true)
	}
}
