package postprocessor

import (
	ansi "clx/utils/strip-ansi"
	"strings"
)

type ruleSet struct {
	skipContains []string
	skipEquals   []string
	endContains  []string
	endEquals    []string
}

func (rs *ruleSet) process(text string) string {
	lines := strings.Split(text, "\n")
	output := ""

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if isOnFirstOrLastLine {
			output += line + "\n"

			continue
		}

		if lineEquals(rs.skipEquals, line) ||
			lineContains(rs.skipContains, line) {
			continue
		}

		if isOnLineBeforeTargetEquals(rs.endEquals, lines, i) ||
			isOnLineBeforeTargetContains(rs.endContains, lines, i) {
			output += "\n"

			break
		}

		output += line + "\n"
	}

	output = strings.ReplaceAll(output, "\n\n\n\n", "\n\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}

func (rs *ruleSet) SkipContains(text string) {
	rs.skipContains = append(rs.skipContains, text)
}

func (rs *ruleSet) SkipEquals(text string) {
	rs.skipEquals = append(rs.skipEquals, text)
}

func (rs *ruleSet) EndContains(text string) {
	rs.endContains = append(rs.endContains, text)
}

func (rs *ruleSet) EndEquals(text string) {
	rs.endEquals = append(rs.endEquals, text)
}

func lineEquals(targets []string, line string) bool {
	for _, target := range targets {
		line = ansi.Strip(line)
		line = strings.TrimLeft(line, " ")

		if line == target {
			return true
		}
	}

	return false
}

func lineContains(targets []string, line string) bool {
	for _, target := range targets {
		if strings.Contains(line, target) {
			return true
		}
	}

	return false
}

func isOnLineBeforeTargetEquals(targets []string, lines []string, i int) bool {
	for _, target := range targets {
		nextLine := lines[i+1]
		nextLine = ansi.Strip(nextLine)
		nextLine = strings.TrimLeft(nextLine, " ")

		if nextLine == target {
			return true
		}
	}

	return false
}

func isOnLineBeforeTargetContains(targets []string, lines []string, i int) bool {
	for _, target := range targets {
		nextLine := lines[i+1]
		nextLine = ansi.Strip(nextLine)
		nextLine = strings.TrimLeft(nextLine, " ")

		if strings.Contains(nextLine, target) {
			return true
		}
	}

	return false
}
