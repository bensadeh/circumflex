package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/provider"
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
	indent             int
	disableHistory     bool
	debugMode          bool
	debugFallible      bool
	nerdFontFlag       bool
	nerdFontChanged    bool
	pageMultiplier     int
	selectedCategories string
	wideView           string
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "clx",
		Short:        style.Magenta("circumflex") + " is a command line tool for browsing Hacker News in your terminal",
		Version:      version.Version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			nerdFontChanged = cmd.Flags().Changed("nerdfonts")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := getConfig()
			if err != nil {
				return err
			}

			style.SetTheme(config.Theme)

			cat, err := categories.New(selectedCategories)
			if err != nil {
				return err
			}

			return view.Run(config, cat)
		},
	}

	registerTemplateFuncs()
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(commentsCmd())
	rootCmd.AddCommand(articleCmd())
	rootCmd.AddCommand(urlCmd())
	rootCmd.AddCommand(defaultThemeCmd())

	configureFlags(rootCmd)

	rootCmd.InitDefaultHelpFlag()
	rootCmd.InitDefaultVersionFlag()

	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "show help"
	}

	if f := rootCmd.Flags().Lookup("version"); f != nil {
		f.Usage = "show version"
	}

	return rootCmd
}

func configureFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", settings.Default().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().IntVarP(&articleWidth, "article-width", "a", settings.Default().ArticleWidth,
		"set the article width in reader mode")
	rootCmd.PersistentFlags().IntVar(&indent, "indent", settings.Default().Indent,
		"set the comment section indent size")
	rootCmd.PersistentFlags().BoolVarP(&nerdFontFlag, "nerdfonts", "n", false,
		"enable or disable Nerd Fonts")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", categories.Default,
		"set the categories in the header\n(available: "+strings.Join(categories.AvailableNames(), ", ")+")")
	rootCmd.PersistentFlags().IntVar(&pageMultiplier, "pages", settings.Default().PageMultiplier,
		"set pages to fetch per category (1-5)")
	rootCmd.PersistentFlags().StringVarP(&wideView, "wide-view", "w", strconv.Itoa(settings.DefaultWideViewMinWidth),
		"show stories in a pane next to the front page: \"always\", \"never\"\nor the minimum terminal width in columns")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true

	rootCmd.PersistentFlags().BoolVar(&debugFallible, "debug-fallible", false,
		"enable debug mode with random failures for testing error handling")
	rootCmd.Flag("debug-fallible").Hidden = true
}

func getConfig() (*settings.Config, error) {
	config := settings.Default()

	t, err := theme.Load(settings.ThemePath())
	if err != nil {
		return nil, fmt.Errorf("could not load theme config: %w", err)
	}

	config.Theme = t

	config.CommentWidth = commentWidth
	config.ArticleWidth = articleWidth
	config.Indent = settings.ClampIndent(indent)
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = resolveNerdFonts(nerdFontFlag, nerdFontChanged)
	config.DebugMode = debugMode
	config.DebugFallible = debugFallible
	config.PageMultiplier = settings.ClampPageMultiplier(pageMultiplier)

	config.WideViewMinWidth, err = settings.ParseWideView(wideView)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func parseID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid story ID", arg)
	}

	return id, nil
}

func resolveNerdFonts(flag, changed bool) bool {
	if changed {
		return flag
	}

	return isGhostty()
}

func isGhostty() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty"
}

func newService() hn.Service {
	return provider.NewService(debugMode, debugFallible)
}

func readerWidth(maxWidth int) int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return maxWidth
	}

	return layout.ReaderContentWidth(w, maxWidth)
}
