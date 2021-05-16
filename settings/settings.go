package settings

import (
	ansi "clx/utils/strip-ansi"
	"strconv"

	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/viper"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

type options struct {
	options []*option
}

type option struct {
	name         string
	key          string
	value        string
	defaultValue string
	description  string
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

func (o options) getConfigFileTemplate() string {
	output := ""
	for i := 0; i < len(o.options); i++ {
		output += o.options[i].printConfig() + newParagraph
	}

	return output
}

func (o option) printConfig() string {
	cleanDesc := ansi.Strip(o.description)
	description, _ := text.WrapWithPad(cleanDesc, 80, "# ")
	separator := newLine + "#" + newLine
	setting := "# " + o.key + "=" + o.defaultValue

	return description + separator + setting
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
