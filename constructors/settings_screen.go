package constructor

import (
	"clx/column"
	"clx/constants/margins"
	"clx/constants/settings"
	"clx/file"
	"clx/screen"
	ansi "clx/utils/strip-ansi"
	"strconv"

	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
	bold         = "\033[1m"
	dimmed       = "\033[2m"
	normal       = "\033[0m"
)

type options struct {
	options []*option
}

func (o *options) addOption(name string, key string, value string, description string) {
	newOption := new(option)
	newOption.name = name
	newOption.key = key
	newOption.value = value
	newOption.description = description

	o.options = append(o.options, newOption)
}

func (o options) printAll(textWidth int) string {
	spaceBetweenDescriptions := 7
	rightMargin := 1
	usableScreenWidth := screen.GetTerminalWidth() - margins.LeftMargin - rightMargin

	if usableScreenWidth > (textWidth*2 + spaceBetweenDescriptions) {
		return printOptionsInTwoColumns(o, textWidth, spaceBetweenDescriptions)
	}

	return printOptionsInOneColumn(o, textWidth)
}

func printOptionsInOneColumn(o options, textWidth int) string {
	output := ""
	for i := 0; i < len(o.options); i++ {
		output += o.options[i].print(textWidth) + newParagraph
	}

	return output
}

func printOptionsInTwoColumns(o options, textWidth int, space int) string {
	output := ""

	for i := 0; i < len(o.options); i += 2 {
		if i+2 <= len(o.options) {
			leftDesc, leftValue := o.options[i].printNoWrap(textWidth)
			rightDesc, rightValue := o.options[i+1].printNoWrap(textWidth)
			output += column.PutInColumns(leftDesc, rightDesc, textWidth, space) + newLine
			output += column.PutInColumns(leftValue, rightValue, textWidth, space) + newLine
		} else {
			output += o.options[i].print(textWidth) + newParagraph
		}
	}

	return output
}

func (o options) getConfigFileTemplate() string {
	output := ""
	for i := 0; i < len(o.options); i++ {
		output += o.options[i].printConfig() + newParagraph
	}

	return output
}

type option struct {
	name        string
	key         string
	value       string
	description string
}

func (o option) print(textWidth int) string {
	wrappedDescription, _ := text.Wrap(o.description, textWidth)
	output := ""

	output += makeHeadline(o.name, o.key, textWidth) + newLine
	output += wrappedDescription + newParagraph
	output += "Current value: " + dim(o.value)

	return output
}

func (o option) printNoWrap(textWidth int) (string, string) {
	description := ""

	description += makeHeadline(o.name, o.key, textWidth) + newLine
	description += o.description

	return description, "Current value: " + dim(o.value)
}

func makeHeadline(name string, key string, textWidth int) string {
	nameLength := text.Len(name)
	keyLength := text.Len(key)
	spaceBetweenNameAndKey := textWidth - nameLength - keyLength

	whiteSpace := ""
	for i := 0; i < spaceBetweenNameAndKey; i++ {
		whiteSpace += " "
	}

	return b(name) + whiteSpace + dim(key)
}

func (o option) printConfig() string {
	cleanDesc := ansi.Strip(o.description)
	description, _ := text.WrapWithPad(cleanDesc, 80, "# ")

	return description + newLine + "#" + o.key + "=" + o.value
}

func GetSettingsText() string {
	message := ""
	pathToConfigDirectory := file.PathToConfigDirectory()
	pathToConfigFile := file.PathToConfigFile()
	commentWidth := getCommentWidth()

	if file.Exists(pathToConfigFile) {
		message += dim("Using config file at " + pathToConfigFile)
	} else {
		message += dim("Press T to create config.env in " + pathToConfigDirectory)
	}

	options := initializeOptions()

	return cview.TranslateANSI(message + newParagraph + options.printAll(commentWidth))
}

func getCommentWidth() int {
	commentWidthFromSettings := viper.GetInt(settings.CommentWidthKey)

	if commentWidthFromSettings == 0 {
		return screen.GetTerminalWidth() - margins.LeftMargin
	}

	return commentWidthFromSettings
}

func GetConfigFileContents() string {
	options := initializeOptions()

	return options.getConfigFileTemplate()
}

func initializeOptions() *options {
	currentCommentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidthKey))
	currentIndentSize := strconv.Itoa(viper.GetInt(settings.IndentSizeKey))
	currentPreserveRightMargin := strconv.FormatBool(viper.GetBool(settings.PreserveRightMarginKey))
	currentHighlightHeadlines := strconv.Itoa(viper.GetInt(settings.HighlightHeadlinesKey))
	currentRelativeNumbering := strconv.FormatBool(viper.GetBool(settings.RelativeNumberingKey))

	options := new(options)
	options.addOption(settings.HighlightHeadlinesName, settings.HighlightHeadlinesKey,
		currentHighlightHeadlines, settings.HighlightHeadlinesDescription)
	options.addOption(settings.CommentWidthName, settings.CommentWidthKey,
		currentCommentWidth, settings.CommentWidthDescription)
	options.addOption(settings.PreserveRightMarginName, settings.PreserveRightMarginKey,
		currentPreserveRightMargin, settings.PreserveRightMarginDescription)
	options.addOption(settings.IndentSizeName, settings.IndentSizeKey,
		currentIndentSize, settings.IndentSizeDescription)
	options.addOption(settings.RelativeNumberingName, settings.RelativeNumberingKey,
		currentRelativeNumbering, settings.RelativeNumberingDescription)

	return options
}

func b(text string) string {
	return bold + text + normal
}

func dim(text string) string {
	return dimmed + text + normal
}
