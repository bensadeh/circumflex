package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
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
	"github.com/spf13/pflag"
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

const maxHelpWidth = 80

func helpWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return maxHelpWidth
	}

	return min(w, maxHelpWidth)
}

func registerTemplateFuncs() {
	w := helpWidth()

	cobra.AddTemplateFunc("header", func(s string) string {
		return ansi.Bold + s + ansi.Reset
	})
	cobra.AddTemplateFunc("stylizeFlags", stylizeFlags)
	cobra.AddTemplateFunc("flagUsages", func(flags *pflag.FlagSet) string {
		// Wrap at w-1 because stylizeFlags adds 1 extra space to each flag line.
		return flags.FlagUsagesWrapped(w - 1)
	})
	cobra.AddTemplateFunc("descCol", func(cmd *cobra.Command) int {
		usage := cmd.Flags().FlagUsagesWrapped(w - 1)
		for line := range strings.SplitSeq(usage, "\n") {
			m := flagDefRe.FindStringSubmatchIndex(line)
			if m != nil && m[1] < len(line) && line[m[1]] == ' ' {
				col := m[1]
				for col < len(line) && line[col] == ' ' {
					col++
				}

				return col + 1 // +1 for the extra space we add in stylizeFlags
			}
		}

		return 20
	})
	cobra.AddTemplateFunc("cmdName", func(use string, col int) string {
		name, args, _ := strings.Cut(use, " ")

		colored := ansi.Blue + name + ansi.Reset
		if args != "" {
			colored += " " + ansi.Yellow + args + ansi.Reset
		}

		padding := col - 3
		padded := fmt.Sprintf("%-*s", padding, use)

		return colored + padded[len(use):]
	})
	cobra.AddTemplateFunc("wrapDesc", func(desc string, col int) string {
		available := w - col
		if available <= 0 || len(desc) <= available {
			return desc
		}

		var b strings.Builder

		indent := strings.Repeat(" ", col)

		for len(desc) > available {
			cut := strings.LastIndex(desc[:available], " ")
			if cut <= 0 {
				cut = available
			}

			b.WriteString(desc[:cut])
			b.WriteString("\n" + indent)

			desc = desc[cut:]
			if len(desc) > 0 && desc[0] == ' ' {
				desc = desc[1:]
			}
		}

		b.WriteString(desc)

		return b.String()
	})
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
		"set the number of pages to fetch per category (min 1, max 5)")

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug-mode", "q", false,
		"enable debug mode (offline mode) by using mock data for the endpoints")
	rootCmd.Flag("debug-mode").Hidden = true

	rootCmd.PersistentFlags().BoolVar(&debugFallible, "debug-fallible", false,
		"enable debug mode with random failures for testing error handling")
	rootCmd.Flag("debug-fallible").Hidden = true
}

func getConfig() *settings.Config {
	config := settings.Default()

	t, err := theme.Load(settings.ThemePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not load theme config\n  %v\n", err)
		os.Exit(1)
	}

	config.Theme = t
	style.Init(t)

	config.CommentWidth = commentWidth
	config.ArticleWidth = articleWidth
	config.Indent = settings.ClampIndent(indent)
	config.DoNotMarkSubmissionsAsRead = disableHistory
	config.EnableNerdFonts = resolveNerdFonts(nerdFontFlag, nerdFontChanged)
	config.DebugMode = debugMode
	config.DebugFallible = debugFallible
	config.PageMultiplier = settings.ClampPageMultiplier(pageMultiplier)

	return config
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

var (
	// flagDefRe matches the flag definition part of a pflag usage line (without
	// trailing padding). A line is a flag definition if m[1] is followed by a space.
	flagDefRe = regexp.MustCompile(`^(\s+)(?:(-\w)(, ))?(--[\w-]+)( \w+)?`)
	defaultRe = regexp.MustCompile(`\(default ([^)]+)\)`)
)

func stylizeFlags(s string) string {
	var b strings.Builder

	for line := range strings.SplitSeq(s, "\n") {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}

		m := flagDefRe.FindStringSubmatchIndex(line)
		if m == nil || m[1] >= len(line) || line[m[1]] != ' ' {
			// Continuation line: shift right by 1 to match the extra space
			// added to flag definition lines.
			trimmed := strings.TrimLeft(line, " ")
			if indent := len(line) - len(trimmed); indent > 6 {
				b.WriteString(strings.Repeat(" ", indent+1) + trimmed)
			} else {
				b.WriteString(line)
			}

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

		// Write original padding + 1 extra space (for 2-space gap on longest flag),
		// then the description.
		b.WriteString(" " + line[m[1]:])
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
{{- if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{$col := descCol .}}{{if eq (len .Groups) 0}}

{{header "Commands:"}}{{range $cmds}}{{if .IsAvailableCommand}}
  {{cmdName .Use $col}} {{wrapDesc .Short $col}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{header .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{cmdName .Use $col}} {{wrapDesc .Short $col}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{header "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{cmdName .Use $col}} {{wrapDesc .Short $col}}{{end}}{{end}}{{end}}{{end}}{{end}}
{{- if .HasAvailableLocalFlags}}

{{header "Flags:"}}
{{flagUsages .LocalFlags | trimTrailingWhitespaces | stylizeFlags}}{{end}}
{{- if .HasAvailableInheritedFlags}}

{{header "Global Flags:"}}
{{flagUsages .InheritedFlags | trimTrailingWhitespaces | stylizeFlags}}{{end}}
{{- if .HasHelpSubCommands}}

{{header "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`
