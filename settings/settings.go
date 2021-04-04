package settings

import (
	"clx/column"
	"clx/constants/margins"
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
		hasAtLeastTwoOptionsLeft := i+2 <= len(o.options)

		if hasAtLeastTwoOptionsLeft {
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
	if currentValue != defaultValue {
		return textUnderline + "*" + currentValue + "*" + textNormal
	}

	return textUnderline + currentValue + textNormal
}

func makeHeadline(name string, key string, textWidth int) string {
	nameLength := text.Len(name)
	keyLength := text.Len(key)
	spaceBetweenNameAndKey := textWidth - nameLength - keyLength
	whiteSpace := strings.Repeat(" ", spaceBetweenNameAndKey)

	return dim(underlined(name + whiteSpace + key))
}

func (o option) printConfig() string {
	cleanDesc := ansi.Strip(o.description)
	description, _ := text.WrapWithPad(cleanDesc, 80, "# ")
	separator := newLine + "#" + newLine
	setting := "# " + o.key + "=" + o.defaultValue

	return description + separator + setting
}

func GetSettingsText() string {
	message := ""
	pathToConfigFile := file.PathToConfigFile()
	commentWidth := getCommentWidth()

	if file.Exists(pathToConfigFile) {
		message += format.Dim("Using config file at " + ConfigFilePath)
	} else {
		message += format.Dim("Press T to create config.env in " + ConfigDirPath)
	}

	o := initializeOptions()
	s := cview.TranslateANSI(message + newParagraph + o.printAll(commentWidth))

	return s
}

func getCommentWidth() int {
	commentWidthFromSettings := viper.GetInt(CommentWidthKey)

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
	currentCommentWidth := strconv.Itoa(viper.GetInt(CommentWidthKey))
	currentIndentSize := strconv.Itoa(viper.GetInt(IndentSizeKey))
	currentPreserveRightMargin := strconv.FormatBool(viper.GetBool(PreserveRightMarginKey))
	currentHighlightHeadlines := strconv.Itoa(viper.GetInt(HighlightHeadlinesKey))
	currentRelativeNumbering := strconv.FormatBool(viper.GetBool(RelativeNumberingKey))
	currentHideYCJobs := strconv.FormatBool(viper.GetBool(HideYCJobsKey))

	o := new(options)
	o.addOption(HighlightHeadlinesName, HighlightHeadlinesKey, currentHighlightHeadlines,
		strconv.Itoa(HighlightHeadlinesDefault), HighlightHeadlinesDescription)
	o.addOption(CommentWidthName, CommentWidthKey, currentCommentWidth,
		strconv.Itoa(CommentWidthDefault), CommentWidthDescription)
	o.addOption(PreserveRightMarginName, PreserveRightMarginKey, currentPreserveRightMargin,
		strconv.FormatBool(PreserveRightMarginDefault), PreserveRightMarginDescription)
	o.addOption(IndentSizeName, IndentSizeKey, currentIndentSize,
		strconv.Itoa(IndentSizeDefault), IndentSizeDescription)
	o.addOption(RelativeNumberingName, RelativeNumberingKey, currentRelativeNumbering,
		strconv.FormatBool(RelativeNumberingDefault), RelativeNumberingDescription)
	o.addOption(HideYCJobsName, HideYCJobsKey, currentHideYCJobs,
		strconv.FormatBool(HideYCJobsDefault), HideYCJobsDescription)

	return o
}

func dim(text string) string {
	return textDimmed + text + textNormal
}

func underlined(text string) string {
	return textUnderline + text + textNormal
}
