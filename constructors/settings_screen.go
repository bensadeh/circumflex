package constructor

import (
	"clx/config"
	"clx/constants/settings"
	"github.com/spf13/viper"
	"os"
	"path"
	"strconv"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

func getSettingsText() string {
	message := ""
	configPath := config.GetConfigPath()
	pathToConfigFile := path.Join(configPath, settings.ConfigFileNameFull)
	settingsScreenText := "Configure circumflex by editing [::b]config.env[::-]. " +
		"You can also export the same variables in your shell (for example in zshenv, bash_profile or config.fish).\n\n" +
		""

	if fileExists(pathToConfigFile) {
		message += "Config file found at " + pathToConfigFile
	} else {
		message += "Config file not found at " + pathToConfigFile + "" +
			""
	}

	return settingsScreenText + newParagraph + message + newParagraph + getCurrentSettings()
}

func fileExists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func getCurrentSettings() string {
	output := ""
	commentWidth := strconv.Itoa(viper.GetInt(settings.CommentWidth))
	indentSize := strconv.Itoa(viper.GetInt(settings.IndentSize))

	output += bold("Comment Section ") + newParagraph
	output += "Comment Width: " + commentWidth + newLine
	output += newLine
	output += "Indent Size:    " + indentSize + newLine

	return output
}

func bold(text string) string {
	return "[::b]" + text + "[::-]"
}
