package cmd

import (
	"clx/bubble"
	"clx/categories"
	"clx/cli"
	"clx/hn"
	"clx/indent"
	"clx/less"
	"clx/settings"
	"clx/style"
	"clx/theme"
	"clx/version"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	disableHeadlineHighlighting bool
	commentWidth                int
	disableCommentHighlighting  bool
	disableHistory              bool
	disableEmojis               bool
	hideIndentSymbol            bool
	debugMode                   bool
	debugFallible               bool
	nerdFontFlag                string
	pageMultiplier              int
	selectedCategories          string
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "\n" + style.Magenta("circumflex") + " is a command line tool for browsing Hacker News in your terminal",
		Version: version.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			config.IndentationSymbol = indent.IndentSymbol(hideIndentSymbol)

			cat, err := categories.New(selectedCategories)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if config.EnableNerdFonts {
				cli.EnableNerdFontsInLess()
			}

			lessKey, err := less.NewLesskey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not create lesskey: %v\n", err)
				os.Exit(1)
			}

			config.LesskeyPath = lessKey.Path()
			defer lessKey.Remove()

			bubble.Run(config, cat)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(commentsCmd())
	rootCmd.AddCommand(articleCmd())
	rootCmd.AddCommand(urlCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(defaultThemeCmd())

	configureFlags(rootCmd)

	return rootCmd
}

func configureFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&disableHeadlineHighlighting, "plain-headlines", "p", false,
		"disable syntax highlighting for headlines")
	rootCmd.PersistentFlags().BoolVarP(&disableCommentHighlighting, "plain-comments", "o", false,
		"disable syntax highlighting for comments")
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().BoolVarP(&disableEmojis, "disable-emojis", "e", false,
		"disable conversion of smileys to emojis")
	rootCmd.PersistentFlags().BoolVarP(&hideIndentSymbol, "hide-indent", "t", false,
		"hide the indentation bar to the left of the reply")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", settings.Default().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().StringVarP(&nerdFontFlag, "nerdfonts", "n", "",
		"enable or disable Nerd Fonts (true/false, auto-enabled for Ghostty, env: NERDFONTS)")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", "top,best,ask,show",
		"set the categories in the header")
	rootCmd.PersistentFlags().IntVar(&pageMultiplier, "pages", settings.Default().PageMultiplier,
		"set the number of pages to fetch per category (1-5)")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true

	rootCmd.PersistentFlags().BoolVar(&debugFallible, "debug-fallible", false,
		"enable debug mode with random failures for testing error handling")
	rootCmd.Flag("debug-fallible").Hidden = true
}

func getConfig() *settings.Config {
	config := settings.Default()

	t, err := theme.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not load theme config\n  %v\n", err)
		os.Exit(1)
	}

	config.Theme = t
	style.Init(t)

	config.CommentWidth = commentWidth
	config.DisableHeadlineHighlighting = disableHeadlineHighlighting
	config.DisableCommentHighlighting = disableCommentHighlighting
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = resolveNerdFonts(nerdFontFlag)
	config.HideIndentSymbol = hideIndentSymbol
	config.DisableEmojis = disableEmojis
	config.DebugMode = debugMode
	config.DebugFallible = debugFallible
	config.PageMultiplier = settings.ClampPageMultiplier(pageMultiplier)

	return config
}

func resolveNerdFonts(flag string) bool {
	switch flag {
	case "true":
		return true
	case "false":
		return false
	default:
		_, nerdFontsEnvIsSet := os.LookupEnv("NERDFONTS")

		return nerdFontsEnvIsSet || isGhostty()
	}
}

func isGhostty() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty"
}

func newService() hn.Service {
	return hn.NewService(debugMode, debugFallible)
}
