package reader

import (
	"clx/ansi"
	"clx/constants"
	"strings"
)

type ruleSet struct {
	skipLineContainsRules []string
	skipLineEqualsRules   []string
	skipParContainsRules  []string
	endLineContainsRules  []string
	endLineEqualsRules    []string
}

func (rs *ruleSet) filter(text string) string {
	paragraphs := strings.Split(text, "\n\n")
	output := ""

	output = filterByParagraph(paragraphs, rs)

	lines := strings.Split(output, "\n")

	output = filterByLine(lines, rs)

	output = strings.ReplaceAll(output, "\n\n\n\n", "\n\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}

func filterByLine(lines []string, rs *ruleSet) string {
	var sb strings.Builder

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if equals(rs.skipLineEqualsRules, line) ||
			contains(rs.skipLineContainsRules, line) {
			continue
		}

		if isOnFirstOrLastLine {
			sb.WriteString(line)
			sb.WriteByte('\n')

			continue
		}

		if isOnLineBeforeTargetEquals(rs.endLineEqualsRules, lines, i) ||
			isOnLineBeforeTargetContains(rs.endLineContainsRules, lines, i) {
			sb.WriteByte('\n')

			break
		}

		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	return sb.String()
}

func filterByParagraph(paragraphs []string, rs *ruleSet) string {
	var sb strings.Builder

	for i, paragraph := range paragraphs {
		isOnFirstOrLastParagraph := i == 0 || i == len(paragraphs)-1
		parNoLeadingWhitespace := strings.TrimLeft(paragraph, " ")

		if len(parNoLeadingWhitespace) == 1 {
			continue
		}

		if contains(rs.skipParContainsRules, paragraph) {
			continue
		}

		if isOnFirstOrLastParagraph {
			sb.WriteString(paragraph)
			sb.WriteString("\n\n")

			continue
		}

		sb.WriteString(paragraph)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (rs *ruleSet) skipLineEquals(text string) {
	rs.skipLineEqualsRules = append(rs.skipLineEqualsRules, text)
}

func (rs *ruleSet) skipParContains(text string) {
	rs.skipParContainsRules = append(rs.skipParContainsRules, text)
}

func (rs *ruleSet) endBeforeLineContains(text string) {
	rs.endLineContainsRules = append(rs.endLineContainsRules, text)
}

func (rs *ruleSet) endBeforeLineEquals(text string) {
	rs.endLineEqualsRules = append(rs.endLineEqualsRules, text)
}

func equals(targets []string, line string) bool {
	for _, target := range targets {
		line = ansi.Strip(line)
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, constants.InvisibleCharacterForTopLevelComments)

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

func isOnLineBeforeTargetEquals(targets []string, lines []string, i int) bool {
	for _, target := range targets {
		nextLine := lines[i+1]
		nextLine = ansi.Strip(nextLine)
		nextLine = strings.TrimSpace(nextLine)
		nextLine = strings.TrimLeft(nextLine, constants.InvisibleCharacterForTopLevelComments)

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
