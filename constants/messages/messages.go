package messages

import "clx/constants/clx"

const (
	Refreshed                   = "Refreshed"
	Cancelled                   = "Cancelled"
	OfflineMessage              = "Could not fetch submissions          [-]Press 'r' to retry or 'q' to quit"
	ConfigConfirmation          = "[::b]config.env[::-] will be created in [::r]~/.config/circumflex[::-], press Y to Confirm"
	ConfigCreatedAt             = "Config created at [::b]~/.config./circumflex/config.env"
	ConfigNotCreated            = "Could not create config file"
	CommentsNotFetched          = "Could not fetch comments"
	EnterCommentSectionToUpdate = "[Enter comment section to update submission]"
	DeleteFromFavorites         = "[red]Delete[-] from Favorites? Press [::b]Y[::-] to Confirm"
	AddToFavorites              = "[green]Add[-] to Favorites? Press [::b]Y[::-] to Confirm"
)

func GetCircumflexStatusMessage() string {
	return "[::d]github.com/bensadeh/circumflex • version " + clx.Version
}
