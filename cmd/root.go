package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/theme"
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	commentWidth       int
	articleWidth       int
	disableHistory     bool
	debugMode          bool
	debugFallible      bool
	nerdFontFlag       string
	pageMultiplier     int
	selectedCategories string
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "\n" + style.Magenta("circumflex") + " is a command line tool for browsing Hacker News in your terminal",
		Version: version.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()

			cat, err := categories.New(selectedCategories)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			view.Run(config, cat)
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
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", settings.Default().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().IntVarP(&articleWidth, "article-width", "a", settings.Default().ArticleWidth,
		"set the article width in reader mode")
	rootCmd.PersistentFlags().StringVarP(&nerdFontFlag, "nerdfonts", "n", "",
		"enable or disable Nerd Fonts (true/false, auto-enabled for Ghostty, env: NERDFONTS)")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", "top,best,ask,show",
		"set the categories in the header (available: "+strings.Join(categories.AvailableNames(), ", ")+")")
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
	config.ArticleWidth = articleWidth
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = resolveNerdFonts(nerdFontFlag)
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

func readerWidth(maxWidth int) int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return maxWidth
	}

	return layout.ReaderContentWidth(w, maxWidth)
}
