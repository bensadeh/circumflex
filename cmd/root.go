package cmd

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/categories"

	"github.com/bensadeh/circumflex/app"
	"github.com/bensadeh/circumflex/bubble"
	"github.com/bensadeh/circumflex/cli"
	"github.com/bensadeh/circumflex/indent"
	"github.com/bensadeh/circumflex/less"
	"github.com/bensadeh/circumflex/settings"

	"github.com/logrusorgru/aurora/v3"
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
	enableNerdFont              bool
	autoExpandComments          bool
	noLessVerify                bool
	selectedCategories          string
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "github.com/bensadeh/circumflex",
		Short:   "\n" + aurora.Magenta("circumflex").String() + " is a command line tool for browsing Hacker News in your terminal",
		Version: app.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			config.IndentationSymbol = indent.GetIndentSymbol(hideIndentSymbol)

			cat := categories.New(selectedCategories)

			verifyLess(noLessVerify)

			if config.EnableNerdFonts {
				cli.EnableNerdFontsInLess()
			}

			lessKey := less.NewLesskey()
			config.LesskeyPath = lessKey.GetPath()
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
	rootCmd.PersistentFlags().BoolVarP(&enableNerdFont, "nerdfonts", "n", false,
		"enable Nerd Fonts")
	rootCmd.PersistentFlags().BoolVarP(&autoExpandComments, "auto-expand", "a", false,
		"automatically expand all replies upon entering the comment section")
	rootCmd.PersistentFlags().BoolVar(&noLessVerify, "no-less-verify", false,
		"disable checking less version on startup")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", "top,best,ask,show",
		"set the categories in the header")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true
}

func getConfig() *settings.Config {
	config := settings.Default()

	_, nerdFontsEnvIsSet := os.LookupEnv("NERDFONTS")

	config.CommentWidth = commentWidth
	config.DisableHeadlineHighlighting = disableHeadlineHighlighting
	config.DisableCommentHighlighting = disableCommentHighlighting
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = nerdFontsEnvIsSet || enableNerdFont
	config.HideIndentSymbol = hideIndentSymbol
	config.AutoExpandComments = autoExpandComments
	config.DisableEmojis = disableEmojis
	config.DebugMode = debugMode
	config.NoLessVerify = noLessVerify

	return config
}

func verifyLess(noLessVerify bool) {
	if noLessVerify {
		return
	}

	isValid, currentLessVersion := cli.VerifyLessVersion(app.MinimumLessVersion)

	if !isValid && currentLessVersion == "" {
		flag := aurora.Bold("--no-less-verify").String()
		lessCmd := aurora.Magenta("less").String()
		clxCmd := aurora.Magenta("github.com/bensadeh/circumflex").String()
		lessVersion := aurora.Yellow("?").String()

		fmt.Printf("Could not verify version of %s\n\n", lessCmd)
		fmt.Printf("Required: %d\n", app.MinimumLessVersion)
		fmt.Printf("Current:  %s\n\n", lessVersion)
		fmt.Printf("Re-run %s with the %s flag to disable this check\n", clxCmd, flag)

		os.Exit(1)
	}

	if !isValid {
		flag := aurora.Bold("--no-less-verify").String()
		lessCmd := aurora.Magenta("less").String()
		clxCmd := aurora.Magenta("github.com/bensadeh/circumflex").String()
		lessVersion := aurora.Yellow(currentLessVersion).String()

		fmt.Printf("Your version of %s is outdated\n\n", lessCmd)
		fmt.Printf("Required: %d\n", app.MinimumLessVersion)
		fmt.Printf("Current:  %s\n\n", lessVersion)
		fmt.Printf("Re-run %s with the %s flag to disable this check\n", clxCmd, flag)

		os.Exit(1)
	}
}
