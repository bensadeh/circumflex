package vim

import (
	"regexp"
	"strconv"
	"strings"
)

func GetItemDown(vimNumberRegister string, currentItem int, itemCount int) int {
	register, _ := strconv.Atoi(vimNumberRegister)
	availableItemsDown := itemCount - currentItem
	isVimRegisterEmpty := register == 0
	isAtTheBottomOfTheList := currentItem+1 == itemCount

	var selectedItem int

	switch {
	case isAtTheBottomOfTheList:
		selectedItem = currentItem
	case isVimRegisterEmpty:
		selectedItem = currentItem + 1
	case register >= availableItemsDown:
		selectedItem = itemCount - 1
	case register < availableItemsDown:
		selectedItem += currentItem + register
	}

	return selectedItem
}

func GetItemUp(vimNumberRegister string, currentItem int) int {
	register, _ := strconv.Atoi(vimNumberRegister)
	availableItemsUp := currentItem
	isVimRegisterEmpty := register == 0
	isAtTheTopOfTheList := currentItem == 0

	var selectedItem int

	switch {
	case isAtTheTopOfTheList:
		selectedItem = currentItem
	case isVimRegisterEmpty:
		selectedItem = currentItem - 1
	case register >= availableItemsUp:
		selectedItem = 0
	case register < availableItemsUp:
		selectedItem += currentItem - register
	}

	return selectedItem
}

func GetItemToJumpTo(vimNumberRegister string, currentItem int, submissionsToShow int, currentPage int) int {
	rankToJumpTo, _ := strconv.Atoi(vimNumberRegister)
	isVimRegisterEmpty := rankToJumpTo == 0

	if isVimRegisterEmpty {
		return currentItem
	}

	itemToJumpTo := rankToJumpTo - (submissionsToShow * currentPage)

	if itemToJumpTo <= 0 {
		return 0
	}

	if itemToJumpTo >= submissionsToShow {
		return submissionsToShow - 1
	}

	return itemToJumpTo - 1
}

func IsNumberWithGAppended(text string) bool {
	expression := regexp.MustCompile(`\d+g`)

	result := expression.FindString(text)

	return result != ""
}

func ContainsOnlyNumbers(text string) bool {
	numbers, _ := strconv.Atoi(text)

	return numbers != 0
}

func FormatRegisterOutput(register string) string {
	return register + strings.Repeat(" ", 5-len(register))
}
