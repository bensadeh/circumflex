package messages

import "clx/constants/clx"

const (
	OfflineMessage              = "Could not fetch submissions          [-]Press 'r' to retry or 'q' to quit"
	ConfigCreatedAt             = "Config created at [::b]~/.config./circumflex/config.env"
	ConfigNotCreated            = "Could not create config file"
	CommentsNotFetched          = "Could not fetch comments"
	EnterCommentSectionToUpdate = "[Enter comment section to update submission]"
)

func GetCircumflexStatusMessage() string {
	return "[::d]github.com/bensadeh/circumflex • version " + clx.Version
}
