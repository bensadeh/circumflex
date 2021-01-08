package constructor

import (
	"clx/constants/settings"
	"clx/file"
	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/viper"
	"strconv"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
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
	output := ""
	for i := 0; i < len(o.options); i++ {
		output += o.options[i].print(textWidth) + newParagraph
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

	output += underline(o.name) + " " + dim(o.key) + newLine
	output += wrappedDescription + newParagraph
	output += "Current value: " + dim(o.value)

	return output
}

func (o option) printConfig() string {
	description, _ := text.WrapWithPad(o.description, 80, "# ")

	return description + newLine + "#" + o.key + "=" + o.value
}

func GetSettingsText() string {
	message := ""
	pathToConfigDirectory := file.PathToConfigDirectory()
	pathToConfigFile := file.PathToConfigFile()

	if file.Exists(pathToConfigFile) {
		message += dim("Using config file at " + pathToConfigFile)
	} else {
		message += dim("Press T to create config.env in " + pathToConfigDirectory)
	}

	options := initializeOptions()

	return message + newParagraph + options.printAll(viper.GetInt(settings.CommentWidthKey))
}

func GetConfigFileContents() string {
	options := initializeOptions()
	return options.getConfigFileTemplate()
}

func initializeOptions() *options {
	currentCommentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidthKey))
	currentIndentSize := strconv.Itoa(viper.GetInt(settings.IndentSizeKey))
	currentPreserveRightMargin := strconv.FormatBool(viper.GetBool(settings.PreserveRightMarginKey))

	options := new(options)
	options.addOption(settings.CommentWidthName, settings.CommentWidthKey, currentCommentWidth, settings.CommentWidthDescription)
	options.addOption(settings.IndentSizeName, settings.IndentSizeKey, currentIndentSize, settings.IndentSizeDescription)
	options.addOption(settings.PreserveRightMarginName, settings.PreserveRightMarginKey, currentPreserveRightMargin, settings.PreserveRightMarginDescription)
	return options
}

func underline(text string) string {
	return "[::u]" + text + "[::-]"
}

func dim(text string) string {
	return "[::d]" + text + "[::-]"
}

func invert(text string) string {
	return "[::r]" + text + "[::-]"
}
