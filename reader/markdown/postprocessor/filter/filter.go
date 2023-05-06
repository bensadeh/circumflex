package filter

import (
	"strings"

	"clx/constants/unicode"
	ansi "clx/utils/strip-ansi"
)

type RuleSet struct {
	skipLineContains []string
	skipLineEquals   []string
	skipParContains  []string
	skipParEquals    []string
	endLineContains  []string
	endLineEquals    []string
}

func (rs *RuleSet) Filter(text string) string {
	paragraphs := strings.Split(text, "\n\n")
	output := ""

	output = filterByParagraph(paragraphs, output, rs)

	lines := strings.Split(output, "\n")
	output = ""

	output = filterByLine(lines, output, rs)

	output = strings.ReplaceAll(output, "\n\n\n\n", "\n\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}

func filterByLine(lines []string, output string, rs *RuleSet) string {
	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if equals(rs.skipLineEquals, line) ||
			contains(rs.skipLineContains, line) {
			continue
		}

		if isOnFirstOrLastLine {
			output += line + "\n"

			continue
		}

		if IsOnLineBeforeTargetEquals(rs.endLineEquals, lines, i) ||
			IsOnLineBeforeTargetContains(rs.endLineContains, lines, i) {
			output += "\n"

			break
		}

		output += line + "\n"
	}

	return output
}

func filterByParagraph(paragraphs []string, output string, rs *RuleSet) string {
	for i, paragraph := range paragraphs {
		isOnFirstOrLastParagraph := i == 0 || i == len(paragraphs)-1
		parNoLeadingWhitespace := strings.TrimLeft(paragraph, " ")

		if len(parNoLeadingWhitespace) == 1 {
			continue
		}

		if equals(rs.skipParEquals, paragraph) ||
			contains(rs.skipParContains, paragraph) {
			continue
		}

		if isOnFirstOrLastParagraph {
			output += paragraph + "\n\n"

			continue
		}

		output += paragraph + "\n\n"
	}

	return output
}

func (rs *RuleSet) SkipLineContains(text string) {
	rs.skipLineContains = append(rs.skipLineContains, text)
}

func (rs *RuleSet) SkipLineEquals(text string) {
	rs.skipLineEquals = append(rs.skipLineEquals, text)
}

func (rs *RuleSet) SkipParContains(text string) {
	rs.skipParContains = append(rs.skipParContains, text)
}

func (rs *RuleSet) SkipParEquals(text string) {
	rs.skipParEquals = append(rs.skipParEquals, text)
}

func (rs *RuleSet) EndBeforeLineContains(text string) {
	rs.endLineContains = append(rs.endLineContains, text)
}

func (rs *RuleSet) EndBeforeLineEquals(text string) {
	rs.endLineEquals = append(rs.endLineEquals, text)
}

func equals(targets []string, line string) bool {
	for _, target := range targets {
		line = ansi.Strip(line)
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, unicode.InvisibleCharacterForTopLevelComments)

		if line == target {
			return true
		}
	}

	return false
}

func contains(targets []string, line string) bool {
	for _, target := range targets {
		target = ansi.Strip(target)
		if strings.Contains(line, target) {
			return true
		}
	}

	return false
}

func IsOnLineBeforeTargetEquals(targets []string, lines []string, i int) bool {
	for _, target := range targets {
		nextLine := lines[i+1]
		nextLine = ansi.Strip(nextLine)
		nextLine = strings.TrimSpace(nextLine)
		nextLine = strings.TrimLeft(nextLine, unicode.InvisibleCharacterForTopLevelComments)

		if nextLine == target {
			return true
		}
	}

	return false
}

func IsOnLineBeforeTargetContains(targets []string, lines []string, i int) bool {
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
