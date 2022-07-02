package cmd

import (
	"clx/app"
	"clx/bubble"
	"clx/indent"
	"clx/settings"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	plainHeadlines   bool
	commentWidth     int
	plainComments    bool
	disableHistory   bool
	disableEmojis    bool
	hideIndentSymbol bool
	debugMode        bool
	enableNerdFont   bool
	forceLightMode   bool
	forceDarkMode    bool
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "circumflex is a command line tool for browsing Hacker News in your terminal",
		Long:    "circumflex is a command line tool for browsing Hacker News in your terminal",
		Version: app.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			config.IndentationSymbol = indent.GetIndentSymbol(hideIndentSymbol)

			bubble.Run(config)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(viewCmd())
	rootCmd.AddCommand(versionCmd())

	configureFlags(rootCmd)

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
	rootCmd.PersistentFlags().BoolVarP(&hideIndentSymbol, "hide-indent", "t", false,
		"hide the indentation bar to the left of the reply")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", settings.New().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().BoolVarP(&enableNerdFont, "nerdfonts", "n", false,
		"enable Nerd Fonts")
	rootCmd.PersistentFlags().BoolVar(&forceLightMode, "force-light-mode", false,
		"Force use light color scheme")
	rootCmd.PersistentFlags().BoolVar(&forceDarkMode, "force-dark-mode", false,
		"Force use dark color scheme")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true
}

func getConfig() *settings.Config {
	config := settings.New()

	config.CommentWidth = commentWidth

	if plainHeadlines {
		config.PlainHeadlines = true
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

	if enableNerdFont {
		config.EnableNerdFonts = true
	}

	if hideIndentSymbol {
		config.HideIndentSymbol = true
	}

	if forceLightMode {
		lipgloss.SetHasDarkBackground(false)
	}

	if forceDarkMode {
		lipgloss.SetHasDarkBackground(true)
	}

	if debugMode {
		config.DebugMode = true
	}

	return config
}
