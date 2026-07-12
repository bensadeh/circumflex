package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

const (
	maxHelpWidth = 80
	// fallbackCol aligns command descriptions when a command has no flags to
	// derive the column from.
	fallbackCol = 20
	flagGap     = 2
)

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
	cobra.AddTemplateFunc("flagSection", func(flags *pflag.FlagSet) string {
		return flagSection(flags, w)
	})
	cobra.AddTemplateFunc("descCol", func(cmd *cobra.Command) int {
		if col := flagDefColumn(cmd.Flags()); col > 0 {
			return col
		}

		return fallbackCol
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
		return wrapIndented(desc, col, w)
	})
}

// flagDef is one flag's rendered definition: the colorized form plus the
// width of its plain text, which the colorized form no longer reveals.
type flagDef struct {
	colored  string
	width    int
	usage    string
	defValue string // display form of the default, "" when suppressed
}

// flagDefs builds each visible flag's definition from the flag data —
// shorthand, name, and value type come from pflag as fields rather than
// being re-parsed out of its rendered usage text.
func flagDefs(flags *pflag.FlagSet) []flagDef {
	var defs []flagDef

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		varname, usage := pflag.UnquoteUsage(f)

		plain, colored := "  ", "  "

		if f.Shorthand != "" {
			plain += "-" + f.Shorthand + ", "
			colored += ansi.Cyan + "-" + f.Shorthand + ansi.Reset + ", "
		} else {
			plain += "    "
			colored += "    "
		}

		plain += "--" + f.Name
		colored += ansi.Cyan + "--" + f.Name + ansi.Reset

		if varname != "" {
			plain += " " + varname
			colored += " " + ansi.Yellow + varname + ansi.Reset
		}

		defs = append(defs, flagDef{
			colored:  colored,
			width:    len(plain),
			usage:    usage,
			defValue: displayDefault(f),
		})
	})

	return defs
}

// flagDefColumn is the description column: the widest flag definition plus a
// gap. Command descriptions align to the same column.
func flagDefColumn(flags *pflag.FlagSet) int {
	col := 0
	for _, d := range flagDefs(flags) {
		col = max(col, d.width)
	}

	if col == 0 {
		return 0
	}

	return col + flagGap
}

// flagSection renders a flag set as aligned, colorized usage lines: the
// definition, the wrapped description beside it, and the default value on
// its own line beneath.
func flagSection(flags *pflag.FlagSet, width int) string {
	defs := flagDefs(flags)
	col := flagDefColumn(flags)

	var b strings.Builder

	for i, d := range defs {
		if i > 0 {
			b.WriteByte('\n')
		}

		b.WriteString(d.colored)
		b.WriteString(strings.Repeat(" ", col-d.width))
		b.WriteString(wrapIndented(d.usage, col, width))

		if d.defValue != "" {
			b.WriteString("\n" + strings.Repeat(" ", col))
			b.WriteString(ansi.Faint + "(default " + ansi.Red + d.defValue + ansi.Reset + ansi.Faint + ")" + ansi.Reset)
		}
	}

	return b.String()
}

// displayDefault is the default value as the help shows it: quoted for
// strings, empty for zero values — which pflag's own usage also suppresses.
func displayDefault(f *pflag.Flag) string {
	switch f.DefValue {
	case "", "false", "0", "[]", "0s", "map[]", "<nil>":
		return ""
	}

	if f.Value.Type() == "string" {
		return fmt.Sprintf("%q", f.DefValue)
	}

	return f.DefValue
}

// wrapIndented word-wraps text to fit width, indenting continuation lines to
// col. Embedded newlines are deliberate breaks and keep the same indent.
func wrapIndented(desc string, col, w int) string {
	indent := strings.Repeat(" ", col)

	segments := strings.Split(desc, "\n")
	for i, seg := range segments {
		segments[i] = wrapSegment(seg, indent, w-col)
	}

	return strings.Join(segments, "\n"+indent)
}

func wrapSegment(desc, indent string, available int) string {
	if available <= 0 || len(desc) <= available {
		return desc
	}

	var b strings.Builder

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
{{flagSection .LocalFlags}}{{end}}
{{- if .HasAvailableInheritedFlags}}

{{header "Global Flags:"}}
{{flagSection .InheritedFlags}}{{end}}
{{- if .HasHelpSubCommands}}

{{header "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`
