package article

import (
	"strings"

	. "github.com/logrusorgru/aurora/v3"
)

func Parse(article string) string {
	article = highlightReferences(article)

	return article
}

func highlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[1]", "["+Red("1").String()+"]")
	input = strings.ReplaceAll(input, "[2]", "["+Yellow("2").String()+"]")
	input = strings.ReplaceAll(input, "[3]", "["+Green("3").String()+"]")
	input = strings.ReplaceAll(input, "[4]", "["+Blue("4").String()+"]")
	input = strings.ReplaceAll(input, "[5]", "["+Cyan("5").String()+"]")
	input = strings.ReplaceAll(input, "[6]", "["+Magenta("6").String()+"]")
	input = strings.ReplaceAll(input, "[7]", "["+White("7").String()+"]")
	input = strings.ReplaceAll(input, "[8]", "["+BrightRed("8").String()+"]")
	input = strings.ReplaceAll(input, "[9]", "["+BrightYellow("9").String()+"]")
	input = strings.ReplaceAll(input, "[10]", "["+BrightGreen("10").String()+"]")
	input = strings.ReplaceAll(input, "[11]", "["+BrightBlue("11").String()+"]")
	input = strings.ReplaceAll(input, "[12]", "["+BrightCyan("12").String()+"]")
	input = strings.ReplaceAll(input, "[13]", "["+White("13").String()+"]")
	input = strings.ReplaceAll(input, "[14]", "["+Red("14").String()+"]")
	input = strings.ReplaceAll(input, "[15]", "["+Yellow("15").String()+"]")
	input = strings.ReplaceAll(input, "[16]", "["+Green("16").String()+"]")

	return input
}
