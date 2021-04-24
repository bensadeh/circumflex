package vim

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Register struct {
	register string
}

func (r *Register) PutInRegister(number rune) {
	isEmptyAndNewNumberIsZero := len(r.register) == 0 && string(number) == "0"
	hasMoreThanThreeDigits := len(r.register) > 2

	switch {
	case isEmptyAndNewNumberIsZero:
		return

	case r.register == "g":
		r.Clear()

	case hasMoreThanThreeDigits:
		r.register = trimFirstRune(r.register)
		r.register += string(number)

	default:
		r.register += string(number)
	}
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)

	return s[i:]
}

func (r *Register) Print() string {
	return r.register + "  "
}

func (r *Register) LowerCaseG(currentItem int, storiesToShow int, currentPage int) int {
	switch {
	case r.register == "g":
		r.Clear()

		return 0

	case r.containsOnlyNumbers():
		r.register += "g"

		return currentItem

	case r.isNumberWithLowerCaseGAppended():
		r.register = strings.TrimSuffix(r.register, "g")
		itemToJumpTo := r.getItemToJumpTo(currentItem, storiesToShow, currentPage)

		r.Clear()

		return itemToJumpTo

	default:
		r.register += "g"

		return currentItem
	}
}

func (r *Register) UpperCaseG(currentItem int, storiesToShow int, currentPage int) int {
	lastElement := storiesToShow - 1

	switch {
	case r.register == "":
		return lastElement

	case r.containsOnlyNumbers():
		itemToJumpTo := r.getItemToJumpTo(currentItem, storiesToShow, currentPage)

		r.Clear()

		return itemToJumpTo

	default:
		r.Clear()

		return currentItem
	}
}

func (r *Register) containsOnlyNumbers() bool {
	return r.toInt() != 0
}

func (r *Register) Clear() {
	r.register = ""
}

func (r *Register) isNumberWithLowerCaseGAppended() bool {
	expression := regexp.MustCompile(`\d+g`)
	result := expression.FindString(r.register)

	return result != ""
}

func (r *Register) GetItemUp(currentItem int) int {
	register := r.toInt()
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

func (r *Register) GetItemDown(currentItem int, itemCount int) int {
	register := r.toInt()
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

func (r *Register) getItemToJumpTo(currentItem int, storiesToShow int, currentPage int) int {
	rankToJumpTo := r.toInt()
	isVimRegisterEmpty := rankToJumpTo == 0

	if isVimRegisterEmpty {
		return currentItem
	}

	itemToJumpTo := rankToJumpTo - (storiesToShow * currentPage)

	if itemToJumpTo <= 0 {
		return 0
	}

	if itemToJumpTo >= storiesToShow {
		return storiesToShow - 1
	}

	return itemToJumpTo - 1
}

func (r *Register) toInt() int {
	integer, _ := strconv.Atoi(r.register)

	return integer
}
