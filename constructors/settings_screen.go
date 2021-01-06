package constructor

import (
	"clx/constants/settings"
	"clx/file"
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
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
	output += "Current value: " + invert(o.value)

	return output
}

func getSettingsText() string {
	message := ""
	pathToConfigDirectory := file.PathToConfigDirectory()
	pathToConfigFile := file.PathToConfigFile()

	if file.Exists(pathToConfigFile) {
		message += "Using config file at " + pathToConfigFile
	} else {
		message += "Configure circumflex by editing [::b]config.env[::-] or by exporting environment variables. " +
			"Press T to create a [::b]config.env[::-] in " + pathToConfigDirectory
	}

	currentCommentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidthKey))
	currentIndentSize := strconv.Itoa(viper.GetInt(settings.IndentSizeKey))

	options := new(options)
	options.addOption(settings.CommentWidthName, settings.CommentWidthKey, currentCommentWidth, settings.CommentWidthDescription)
	options.addOption(settings.IndentSizeName, settings.IndentSizeKey, currentIndentSize, settings.IndentSizeDescription)

	return message + newParagraph + options.printAll(70)
}

func NewDialogueBox() *cview.Modal {

	modal := cview.NewModal()
	modal.SetText("Would you like to create a config file at " + file.PathToConfigFile() +
		"? \n  \n   \n Press Y to confirm or N to cancel.")
	modal.SetBackgroundColor(tcell.ColorDefault)
	modal.SetTextColor(tcell.ColorDefault)

	return modal
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
