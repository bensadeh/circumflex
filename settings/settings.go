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
	name         string
	key          string
	defaultValue string
	description  string
}

func (o *options) addOption(name, key, defaultValue, description string) {
	newOption := new(option)
	newOption.name = name
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

	o.addOption(HighlightHeadlinesName, HighlightHeadlinesKey,
		strconv.Itoa(HighlightHeadlinesDefault), HighlightHeadlinesDescription)
	o.addOption(CommentWidthName, CommentWidthKey,
		strconv.Itoa(CommentWidthDefault), CommentWidthDescription)
	o.addOption(PreserveRightMarginName, PreserveRightMarginKey,
		strconv.FormatBool(PreserveRightMarginDefault), PreserveRightMarginDescription)
	o.addOption(IndentSizeName, IndentSizeKey,
		strconv.Itoa(IndentSizeDefault), IndentSizeDescription)
	o.addOption(RelativeNumberingName, RelativeNumberingKey,
		strconv.FormatBool(RelativeNumberingDefault), RelativeNumberingDescription)
	o.addOption(HideYCJobsName, HideYCJobsKey,
		strconv.FormatBool(HideYCJobsDefault), HideYCJobsDescription)
	o.addOption(UseAlternateIndentBlockName, UseAlternateIndentBlockKey,
		strconv.FormatBool(UseAlternateIndentBlockDefault), UseAlternateIndentBlockDescription)
	o.addOption(CommentHighlightingName, CommentHighlightingKey,
		strconv.FormatBool(CommentHighlightingDefault), CommentHighlightingDescription)
	o.addOption(EmojiSmileysName, EmojiSmileysKey,
		strconv.FormatBool(EmojiSmileysDefault), EmojiSmileysDescription)
	o.addOption(MarkAsReadName, MarkAsReadKey,
		strconv.FormatBool(MarkAsReadDefault), MarkAsReadDescription)

	return o
}
