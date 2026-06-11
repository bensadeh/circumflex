package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

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
		return splitDefaults(flags.FlagUsagesWrapped(w - 1))
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

var (
	// flagDefRe matches the flag definition part of a pflag usage line (without
	// trailing padding). A line is a flag definition if m[1] is followed by a space.
	flagDefRe = regexp.MustCompile(`^(\s+)(?:(-\w)(, ))?(--[\w-]+)( \w+)?`)
	defaultRe = regexp.MustCompile(`\(default ([^)]+)\)`)
)

// splitDefaults moves any "(default X)" segment to its own continuation line so
// every flag's default value appears below the description rather than inline.
func splitDefaults(s string) string {
	var col int

	for line := range strings.SplitSeq(s, "\n") {
		m := flagDefRe.FindStringSubmatchIndex(line)
		if m == nil || m[1] >= len(line) || line[m[1]] != ' ' {
			continue
		}

		c := m[1]
		for c < len(line) && line[c] == ' ' {
			c++
		}

		col = c

		break
	}

	if col == 0 {
		return s
	}

	indent := strings.Repeat(" ", col)

	var b strings.Builder

	first := true

	for line := range strings.SplitSeq(s, "\n") {
		if !first {
			b.WriteByte('\n')
		}

		first = false

		loc := defaultRe.FindStringIndex(line)
		if loc == nil {
			b.WriteString(line)

			continue
		}

		before := strings.TrimRight(line[:loc[0]], " ")
		if before == "" {
			b.WriteString(line)

			continue
		}

		b.WriteString(before)
		b.WriteByte('\n')
		b.WriteString(indent)
		b.WriteString(line[loc[0]:loc[1]])

		if loc[1] < len(line) {
			b.WriteString(line[loc[1]:])
		}
	}

	return b.String()
}

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
