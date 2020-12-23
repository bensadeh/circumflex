package settings

import "strings"

type OptionsFormatter struct {
	Options               []string
	CurrentlySelectedItem int
	SpacingBetweenOptions string
}

func NewOptionsFormatter(options []string, selectedItem int, spacingSize int) *OptionsFormatter {
	of := new(OptionsFormatter)
	of.Options = options
	of.CurrentlySelectedItem = selectedItem
	of.SpacingBetweenOptions = strings.Repeat(" ", spacingSize)
	return of
}

func (of OptionsFormatter) PrintOptions() string {
	numberOfOptions := len(of.Options)
	options := ""

	for i := 0; i < numberOfOptions; i++ {
		if i == of.CurrentlySelectedItem {
			options += of.SpacingBetweenOptions + "[" + of.Options[i] + "]"
		} else {
			options += of.SpacingBetweenOptions + of.Options[i]
		}
	}

	return options
}

func PrintOptions(options []string, selectedItem int, spacingSize int) string {
	spacing := strings.Repeat(" ", spacingSize)
	output := ""

	for i := 0; i < len(options); i++ {
		if i == selectedItem {
			output += spacing + "[" + options[i] + "]"
		} else {
			output += spacing + options[i]
		}
	}

	return output
}
