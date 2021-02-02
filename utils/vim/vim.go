package vim

import "strconv"

func GetItemDown(vimNumberRegister string, currentItem int, itemCount int) int {
	register, _ := strconv.Atoi(vimNumberRegister)
	availableItemsDown := itemCount - currentItem
	isNumbersInRegister := vimNumberRegister != ""
	isAtTheBottomOfTheList := currentItem+1 == itemCount

	var selectedItem int

	switch {
	case isAtTheBottomOfTheList:
		selectedItem = currentItem
	case !isNumbersInRegister:
		selectedItem = currentItem + 1
	case register >= availableItemsDown:
		selectedItem = itemCount - 1
	case register < availableItemsDown:
		selectedItem += currentItem + register
	default:
		break
	}

	return selectedItem
}

func GetItemUp(vimNumberRegister string, currentItem int) int {
	register, _ := strconv.Atoi(vimNumberRegister)
	availableItemsUp := currentItem
	isNumbersInRegister := vimNumberRegister != ""
	isAtTheTopOfTheList := currentItem == 0

	var selectedItem int

	switch {
	case isAtTheTopOfTheList:
		selectedItem = currentItem
	case !isNumbersInRegister:
		selectedItem = currentItem - 1
	case register >= availableItemsUp:
		selectedItem = 0
	case register < availableItemsUp:
		selectedItem -= currentItem + register
	default:
		break
	}

	return selectedItem
}
