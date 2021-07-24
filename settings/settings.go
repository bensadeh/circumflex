package settings

import (
	ansi "clx/utils/strip-ansi"
	"strconv"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

type options struct {
	options []*option
}

type option struct {
	key          string
	defaultValue string
	description  string
}

func (o *options) addOption(key, defaultValue, description string) {
	newOption := new(option)
	newOption.key = key
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
	o := new(options)

	o.addOption(HighlightHeadlinesKey, strconv.Itoa(HighlightHeadlinesDefault), HighlightHeadlinesDescription)
	o.addOption(CommentWidthKey, strconv.Itoa(CommentWidthDefault), CommentWidthDescription)
	o.addOption(PreserveRightMarginKey, strconv.FormatBool(PreserveRightMarginDefault), PreserveRightMarginDescription)
	o.addOption(IndentSizeKey, strconv.Itoa(IndentSizeDefault), IndentSizeDescription)
	o.addOption(RelativeNumberingKey, strconv.FormatBool(RelativeNumberingDefault), RelativeNumberingDescription)
	o.addOption(HideYCJobsKey, strconv.FormatBool(HideYCJobsDefault), HideYCJobsDescription)
	o.addOption(UseAltIndentBlockKey, strconv.FormatBool(UseAltIndentBlockDefault), UseAltIndentBlockDescription)
	o.addOption(CommentHighlightingKey, strconv.FormatBool(CommentHighlightingDefault), CommentHighlightingDescription)
	o.addOption(EmojiSmileysKey, strconv.FormatBool(EmojiSmileysDefault), EmojiSmileysDescription)
	o.addOption(MarkAsReadKey, strconv.FormatBool(MarkAsReadDefault), MarkAsReadDescription)

	return o
}
