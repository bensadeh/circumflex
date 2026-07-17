package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/provider"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/theme"
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view"

	"github.com/spf13/cobra"
)

var (
	commentWidth       int
	articleWidth       int
	indent             int
	disableHistory     bool
	debugMode          bool
	debugFallible      bool
	nerdFontFlag       bool
	enableImages       bool
	pageMultiplier     int
	selectedCategories string
	wideView           string

	// currentCmd is the command being executed, captured so getConfig can
	// tell explicitly-passed flags (which override config.toml) from ones
	// still at their default.
	currentCmd *cobra.Command
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "clx",
		Short:        style.Magenta("circumflex") + " is a command line tool for browsing Hacker News in your terminal",
		Version:      version.Version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			currentCmd = cmd
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := getConfig()
			if err != nil {
				return err
			}

			style.SetTheme(config.Theme)

			cat, err := categories.New(config.Categories)
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
	rootCmd.AddCommand(defaultConfigCmd())

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
	rootCmd.PersistentFlags().BoolVar(&enableImages, "reader-mode-images", envImagesEnabled(),
		"show article images in reader mode\n(opt-in; also set with CLX_READER_MODE_IMAGES=1)")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", settings.Default().Categories,
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

	fileConfig, err := settings.LoadConfig(settings.ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	if err := fileConfig.Apply(config); err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	t, err := theme.Load(settings.ThemePath())
	if err != nil {
		return nil, fmt.Errorf("could not load theme config: %w", err)
	}

	config.Theme = t

	if flagChanged("comment-width") {
		config.CommentWidth = commentWidth
	}

	if flagChanged("article-width") {
		config.ArticleWidth = articleWidth
	}

	if flagChanged("indent") {
		config.Indent = settings.ClampIndent(indent)
	}

	if flagChanged("disable-history") {
		config.DoNotMarkSubmissionsAsRead = disableHistory
	}

	if flagChanged("categories") {
		config.Categories = selectedCategories
	}

	if flagChanged("pages") {
		config.PageMultiplier = settings.ClampPageMultiplier(pageMultiplier)
	}

	switch {
	case flagChanged("nerdfonts"):
		config.EnableNerdFonts = nerdFontFlag
	case fileConfig.NerdFonts == nil:
		config.EnableNerdFonts = isGhostty()
	}

	switch {
	case flagChanged("reader-mode-images"):
		config.EnableImages = enableImages
	case fileConfig.Images == nil:
		config.EnableImages = envImagesEnabled()
	}

	if flagChanged("wide-view") {
		config.WideViewMinWidth, err = settings.ParseWideView(wideView)
		if err != nil {
			return nil, fmt.Errorf("--wide-view: %w", err)
		}
	}

	config.DebugMode = debugMode
	config.DebugFallible = debugFallible

	return config, nil
}

func flagChanged(name string) bool {
	return currentCmd != nil && currentCmd.Flags().Changed(name)
}

func parseID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid story ID", arg)
	}

	return id, nil
}

func isGhostty() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty"
}

// envImagesEnabled reads CLX_READER_MODE_IMAGES as the default for
// --reader-mode-images, so images can be turned on globally without passing the
// flag every time.
func envImagesEnabled() bool {
	enabled, err := strconv.ParseBool(os.Getenv("CLX_READER_MODE_IMAGES"))

	return err == nil && enabled
}

func newService() hn.Service {
	return provider.NewService(debugMode, debugFallible)
}
