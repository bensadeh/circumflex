package constructor

import (
	"clx/config"
	"clx/constants/settings"
	text "github.com/MichaelMure/go-term-text"
	"github.com/spf13/viper"
	"os"
	"path"
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
	output := "Options length: " + strconv.Itoa(len(o.options)) + newParagraph
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
	configPath := config.GetConfigPath()
	pathToConfigFile := path.Join(configPath, settings.ConfigFileNameFull)
	settingsScreenText := "Configure circumflex by editing [::b]config.env[::-] or by exporting environment variables. "

	if fileExists(pathToConfigFile) {
		message += "Config file found at " + pathToConfigFile
	} else {
		message += "Press T to create a [::b]config.env[::-] in " + configPath
	}

	commentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidthKey))
	indentSize := strconv.Itoa(viper.GetInt(settings.IndentSizeKey))

	options := new(options)
	options.addOption(settings.CommentWidthName, settings.CommentWidthKey, commentWidth, settings.CommentWidthDescription)
	options.addOption(settings.IndentSizeName, settings.IndentSizeKey, indentSize, settings.IndentSizeDescription)

	return settingsScreenText + newParagraph + message + newParagraph + options.printAll(70)
}

func fileExists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
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
