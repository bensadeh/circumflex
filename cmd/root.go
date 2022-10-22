package cmd

import (
	"fmt"
	"os"

	"clx/app"
	"clx/bubble"
	"clx/cli"
	"clx/indent"
	"clx/less"
	"clx/settings"

	"github.com/charmbracelet/lipgloss"
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
	forceLightMode              bool
	forceDarkMode               bool
	autoExpandComments          bool
	noLessVerify                bool
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "\n" + aurora.Magenta("circumflex").Italic().String() + " is a command line tool for browsing Hacker News in your terminal",
		Version: app.Version,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			config.IndentationSymbol = indent.GetIndentSymbol(hideIndentSymbol)

			verifyLess(noLessVerify)

			lesskey := less.NewLesskey()
			config.LesskeyPath = lesskey.GetPath()
			defer lesskey.Remove()

			bubble.Run(config)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(viewCmd())
	rootCmd.AddCommand(readCmd())
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
	rootCmd.PersistentFlags().BoolVar(&forceLightMode, "force-light-mode", false,
		"force use light color scheme")
	rootCmd.PersistentFlags().BoolVar(&forceDarkMode, "force-dark-mode", false,
		"force use dark color scheme")
	rootCmd.PersistentFlags().BoolVarP(&autoExpandComments, "auto-expand", "a", false,
		"automatically expand all replies upon entering the comment section")
	rootCmd.PersistentFlags().BoolVar(&noLessVerify, "no-less-verify", false,
		"disable checking less version on startup")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true
}

func getConfig() *settings.Config {
	config := settings.Default()

	config.CommentWidth = commentWidth
	config.DisableHeadlineHighlighting = disableHeadlineHighlighting
	config.DisableCommentHighlighting = disableCommentHighlighting
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = enableNerdFont
	config.HideIndentSymbol = hideIndentSymbol
	config.AutoExpandComments = autoExpandComments
	config.DisableEmojis = disableEmojis
	config.DebugMode = debugMode
	config.NoLessVerify = noLessVerify

	if forceLightMode {
		lipgloss.SetHasDarkBackground(false)
	}

	if forceDarkMode {
		lipgloss.SetHasDarkBackground(true)
	}

	return config
}

func verifyLess(noLessVerify bool) {
	if noLessVerify {
		return
	}

	isValid, currentLessVersion := cli.VerifyLessVersion(app.MinimumLessVersion)

	if !isValid {
		flag := aurora.Bold("--no-less-verify").String()
		less := aurora.Magenta("less").String()
		clx := aurora.Magenta("clx").String()

		fmt.Printf("Your version of %s is outdated\n\n", less)
		fmt.Printf("Your version:     %d\n", currentLessVersion)
		fmt.Printf("Required version: %d\n\n", app.MinimumLessVersion)
		fmt.Printf("If you think this is an error, re-run %s with the %s flag to disable check", clx, flag)

		os.Exit(1)
	}
}
