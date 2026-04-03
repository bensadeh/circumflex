package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
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
		Short:   style.Magenta("circumflex") + " is a command line tool for browsing Hacker News in your terminal",
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

	return rootCmd
}

func registerTemplateFuncs() {
	cobra.AddTemplateFunc("header", func(s string) string {
		return ansi.Bold + s + ansi.Reset
	})
	cobra.AddTemplateFunc("stylizeFlags", stylizeFlags)
	cobra.AddTemplateFunc("usePadding", func(cmds []*cobra.Command) int {
		longest := 0
		for _, c := range cmds {
			if n := len(c.Use); n > longest {
				longest = n
			}
		}

		return longest
	})
	cobra.AddTemplateFunc("cmdName", func(use string, padding int) string {
		name, args, _ := strings.Cut(use, " ")

		colored := ansi.Blue + name + ansi.Reset
		if args != "" {
			colored += " " + ansi.Yellow + args + ansi.Reset
		}

		padded := fmt.Sprintf("%-*s", padding, use)

		return colored + padded[len(use):]
	})
}

func configureFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&disableHistory, "disable-history", "d", false,
		"disable marking stories as read")
	rootCmd.PersistentFlags().IntVarP(&commentWidth, "comment-width", "c", settings.Default().CommentWidth,
		"set the comment width")
	rootCmd.PersistentFlags().IntVarP(&articleWidth, "article-width", "a", settings.Default().ArticleWidth,
		"set the article width in reader mode")
	rootCmd.PersistentFlags().StringVarP(&nerdFontFlag, "nerdfonts", "n", "",
		"enable or disable Nerd Fonts (true/false, auto-enabled for Ghostty, env: "+ansi.Green+"NERDFONTS"+ansi.Reset+")")
	rootCmd.PersistentFlags().StringVar(&selectedCategories, "categories", "top,best,ask,show",
		"set the categories in the header\n(available: "+strings.Join(categories.AvailableNames(), ", ")+")")
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

var (
	// flagLineRe matches the structured part of a pflag usage line:
	// optional short flag (-X, ), long flag (--name), and optional type (int, string, ...).
	// The type is distinguished from the description by being followed by 2+ spaces.
	flagLineRe = regexp.MustCompile(`^(\s+)(?:(-\w)(, ))?(--[\w-]+)( \w+)?\s{2}`)
	defaultRe  = regexp.MustCompile(`\(default ([^)]+)\)`)
)

func stylizeFlags(s string) string {
	var b strings.Builder

	for line := range strings.SplitSeq(s, "\n") {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}

		m := flagLineRe.FindStringSubmatchIndex(line)
		if m == nil {
			b.WriteString(line)

			continue
		}

		b.WriteString(line[m[2]:m[3]]) // leading whitespace

		if m[4] >= 0 { // short flag
			b.WriteString(ansi.Cyan + line[m[4]:m[5]] + ansi.Reset)
			b.WriteString(line[m[6]:m[7]]) // ", "
		}

		b.WriteString(ansi.Cyan + line[m[8]:m[9]] + ansi.Reset) // long flag

		if m[10] >= 0 { // type
			b.WriteString(ansi.Yellow + line[m[10]:m[11]] + ansi.Reset)
		}

		b.WriteString(line[m[1]:]) // rest of line (padding + description)
	}

	// Second pass: colorize default values across the full output.
	// This runs after the per-line pass because defaults may appear anywhere
	// in the description text, not just on the structured flag line.
	result := b.String()
	result = defaultRe.ReplaceAllString(result,
		ansi.Faint+"(default "+ansi.Red+"${1}"+ansi.Reset+ansi.Faint+")"+ansi.Reset)

	return result
}

const usageTemplate = `
{{- if gt (len .Aliases) 0}}
{{header "Aliases:"}}
  {{.NameAndAliases}}
{{end -}}
{{- if .HasExample}}
{{header "Examples:"}}
{{.Example}}
{{end -}}
{{- if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{$pad := usePadding $cmds}}{{if eq (len .Groups) 0}}

{{header "Commands:"}}{{range $cmds}}{{if .IsAvailableCommand}}
  {{cmdName .Use $pad}} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{header .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{cmdName .Use $pad}} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{header "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{cmdName .Use $pad}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}
{{- if .HasAvailableLocalFlags}}

{{header "Flags:"}}
{{.LocalFlags.FlagUsagesWrapped 80 | trimTrailingWhitespaces | stylizeFlags}}{{end}}
{{- if .HasAvailableInheritedFlags}}

{{header "Global Flags:"}}
{{.InheritedFlags.FlagUsagesWrapped 80 | trimTrailingWhitespaces | stylizeFlags}}{{end}}
{{- if .HasHelpSubCommands}}

{{header "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`
