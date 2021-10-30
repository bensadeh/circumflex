package cmd

import (
	"clx/clx"
	clx2 "clx/constants/clx"
	"clx/core"
	"clx/indent"

	"github.com/spf13/cobra"
)

var (
	plainHeadlines       bool
	commentWidth         int
	plainComments        bool
	disableHistory       bool
	disableEmojis        bool
	useRelativeNumbering bool
	showYCJobs           bool
	hideIndentSymbol     bool
	headerType           int
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "circumflex is a command line tool for browsing Hacker News in your terminal",
		Long:    "circumflex is a command line tool for browsing Hacker News in your terminal",
		Version: clx2.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			config.IndentationSymbol = indent.GetIndentSymbol(hideIndentSymbol)

			clx.Run(config)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	configureFlags(rootCmd)

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
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
	rootCmd.PersistentFlags().BoolVarP(&disableEmojis, "disable-emojis", "s", false,
		"disable conversion of smileys to emojis")
	rootCmd.PersistentFlags().BoolVarP(&useRelativeNumbering, "relative-numbering", "r", false,
		"use relative numbering for submissions")
	rootCmd.PersistentFlags().BoolVarP(&showYCJobs, "show-jobs", "j", false,
		"show submissions of the type 'X is hiring'")
	rootCmd.PersistentFlags().BoolVarP(&hideIndentSymbol, "hide-indent", "t", false,
		"hide the indentation bar to the left of the reply")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", core.GetConfigWithDefaults().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().IntVarP(&headerType, "header-type", "e", 0,
		"set the header type on the main screen")
}

func getConfig() *core.Config {
	config := core.GetConfigWithDefaults()

	config.CommentWidth = commentWidth
	config.HeaderType = headerType

	if plainHeadlines {
		config.HighlightHeadlines = false
	}

	if plainComments {
		config.HighlightComments = false
	}

	if disableHistory {
		config.MarkAsRead = false
	}

	if disableEmojis {
		config.EmojiSmileys = false
	}

	if useRelativeNumbering {
		config.RelativeNumbering = true
	}

	if showYCJobs {
		config.HideYCJobs = false
	}

	if hideIndentSymbol {
		config.HideIndentSymbol = true
	}

	return config
}
