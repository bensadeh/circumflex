package cmd

import (
	"clx/clx"
	clx2 "clx/constants/clx"
	"clx/settings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	plainHeadlines       bool
	plainComments        bool
	disableHistory       bool
	altIndentBlock       bool
	smileyEmojis         bool
	relativeNumbering    bool
	showYCJobs           bool
	preserveCommentWidth bool
	hideIndentSymbol     bool
	orangeHeader         bool
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx [add id|clear|config|view id]",
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
	rootCmd.PersistentFlags().BoolVarP(&preserveCommentWidth, "preserve-comment-width", "m", false,
		"do not shorten the comment width for replies")
	rootCmd.PersistentFlags().BoolVarP(&hideIndentSymbol, "hide-indent", "t", false,
		"hide the indentation bar to the left of the reply")
	rootCmd.PersistentFlags().BoolVarP(&orangeHeader, "orange-header", "n", false,
		"set the header on orange")

	rootCmd.PersistentFlags().IntP("comment-width", "c", settings.CommentWidthDefault,
		"set the comment width")
	_ = viper.BindPFlag(settings.CommentWidthKey, rootCmd.PersistentFlags().Lookup("comment-width"))
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

	if hideIndentSymbol {
		viper.Set(settings.HideIndentSymbolKey, true)
	}

	if orangeHeader {
		viper.Set(settings.OrangeHeaderKey, true)
	}
}
