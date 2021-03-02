package constructor

import (
	"clx/column"
	"clx/constants/margins"
	"clx/constants/settings"
	"clx/file"
	"clx/screen"
	"clx/utils/format"
	ansi "clx/utils/strip-ansi"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

const (
	newLine       = "\n"
	newParagraph  = "\n\n"
	textDimmed    = "\033[2m"
	textUnderline = "\033[4m"
	textNormal    = "\033[0m"
)

type options struct {
	options []*option
}

func (o *options) addOption(name, key, value, defaultValue, description string) {
	newOption := new(option)
	newOption.name = name
	newOption.key = key
	newOption.value = value
	newOption.defaultValue = defaultValue
	newOption.description = description

	o.options = append(o.options, newOption)
}

func (o options) printAll(textWidth int) string {
	usableScreenWidth := screen.GetTerminalWidth() - margins.LeftMargin - margins.RightMargin
	hasEnoughScreenSpace := usableScreenWidth > (textWidth*2 + margins.SpaceBetweenDescriptions)

	if hasEnoughScreenSpace {
		return printOptionsInTwoColumns(o, textWidth, margins.SpaceBetweenDescriptions)
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
		isAtLeastTwoOptionsLeft := i+2 <= len(o.options)

		if isAtLeastTwoOptionsLeft {
			leftDesc := o.options[i].print(textWidth)
			rightDesc := o.options[i+1].print(textWidth)
			output += column.PutInColumns(leftDesc, rightDesc, textWidth, space) + newLine
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
	name         string
	key          string
	value        string
	defaultValue string
	description  string
}

func (o option) print(textWidth int) string {
	currentValue := highlight(o.value, o.defaultValue)
	headline := makeHeadline(o.name, currentValue, textWidth) + newLine
	wrappedDescription, _ := text.Wrap(o.description, textWidth)

	return headline + wrappedDescription + newLine
}

func highlight(currentValue string, defaultValue string) string {
	overridden := ""

	if currentValue != defaultValue {
		overridden = "*"
	}

	return textUnderline + overridden + currentValue + overridden + textNormal
}

func makeHeadline(name string, key string, textWidth int) string {
	nameLength := text.Len(name)
	keyLength := text.Len(key)
	spaceBetweenNameAndKey := textWidth - nameLength - keyLength
	whiteSpace := strings.Repeat(" ", spaceBetweenNameAndKey)

	return underlined(name + whiteSpace + key)
}

func (o option) printConfig() string {
	cleanDesc := ansi.Strip(o.description)
	description, _ := text.WrapWithPad(cleanDesc, 80, "# ")

	return description + newLine + "# " + o.key + "=" + o.defaultValue
}

func GetSettingsText() string {
	message := ""
	pathToConfigFile := file.PathToConfigFile()
	commentWidth := getCommentWidth()

	if file.Exists(pathToConfigFile) {
		message += format.Dim("Using config file at " + settings.ConfigFilePath)
	} else {
		message += format.Dim("Press T to create config.env in " + settings.ConfigDirPath)
	}

	o := initializeOptions()
	s := cview.TranslateANSI(message + newParagraph + o.printAll(commentWidth))

	return s
}

func getCommentWidth() int {
	commentWidthFromSettings := viper.GetInt(settings.CommentWidthKey)

	if commentWidthFromSettings == 0 {
		return screen.GetTerminalWidth() - margins.LeftMargin
	}

	return commentWidthFromSettings
}

func GetConfigFileContents() string {
	o := initializeOptions()

	return o.getConfigFileTemplate()
}

func initializeOptions() *options {
	currentCommentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidthKey))
	currentIndentSize := strconv.Itoa(viper.GetInt(settings.IndentSizeKey))
	currentPreserveRightMargin := strconv.FormatBool(viper.GetBool(settings.PreserveRightMarginKey))
	currentHighlightHeadlines := strconv.Itoa(viper.GetInt(settings.HighlightHeadlinesKey))
	currentRelativeNumbering := strconv.FormatBool(viper.GetBool(settings.RelativeNumberingKey))
	currentHideYCJobs := strconv.FormatBool(viper.GetBool(settings.HideYCJobsKey))

	o := new(options)
	o.addOption(settings.HighlightHeadlinesName, settings.HighlightHeadlinesKey, currentHighlightHeadlines,
		strconv.Itoa(settings.HighlightHeadlinesDefault), settings.HighlightHeadlinesDescription)
	o.addOption(settings.CommentWidthName, settings.CommentWidthKey, currentCommentWidth,
		strconv.Itoa(settings.CommentWidthDefault), settings.CommentWidthDescription)
	o.addOption(settings.PreserveRightMarginName, settings.PreserveRightMarginKey, currentPreserveRightMargin,
		strconv.FormatBool(settings.PreserveRightMarginDefault), settings.PreserveRightMarginDescription)
	o.addOption(settings.IndentSizeName, settings.IndentSizeKey, currentIndentSize,
		strconv.Itoa(settings.IndentSizeDefault), settings.IndentSizeDescription)
	o.addOption(settings.RelativeNumberingName, settings.RelativeNumberingKey, currentRelativeNumbering,
		strconv.FormatBool(settings.RelativeNumberingDefault), settings.RelativeNumberingDescription)
	o.addOption(settings.HideYCJobsName, settings.HideYCJobsKey, currentHideYCJobs,
		strconv.FormatBool(settings.HideYCJobsDefault), settings.HideYCJobsDescription)

	return o
}

func dim(text string) string {
	return textDimmed + text + textNormal
}

func underlined(text string) string {
	return textUnderline + text + textNormal
}
