package messages

import (
	"clx/constants/clx"
	"strings"
)

const (
	separator = "-"

	Refreshed          = "Refreshed"
	Cancelled          = "Cancelled"
	OfflineMessage     = "Could not fetch stories          [-]Press 'r' to retry or 'q' to quit"
	ConfigConfirmation = "[::b]config.env[::-] will be created in [::r]~/.config/circumflex[::-], " +
		"press Y to Confirm"
	ConfigCreatedAt             = "Config created at [::b]~/.config./circumflex/config.env"
	ConfigNotCreated            = "Could not create config file"
	CommentsNotFetched          = "Could not fetch comments"
	EnterCommentSectionToUpdate = "[Enter comment section to update story]"
	DeleteFromFavorites         = "[red]Delete[-] from Favorites? Press [::b]Y[::-] to Confirm"
	AddToFavorites              = "[green]Add[-] to Favorites? Press [::b]Y[::-] to Confirm"
	LessScreenInfo              = "You are now in 'less' • Press 'q' to return and 'h' for help"
	HowToExitF                  = "[::d]Leave ID blank to return to main screen"
)

func GetCircumflexStatusMessage() string {
	return "[::d]github.com/bensadeh/circumflex • version " + clx.Version
}

func GetSeparator(width int) string {
	return strings.Repeat(separator, width)
}
