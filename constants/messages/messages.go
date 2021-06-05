package messages

import (
	"clx/constants/clx"
	"strings"
)

const (
	separator = "-"

	Refreshed                   = "Refreshed"
	Cancelled                   = "[::d]Cancelled"
	OfflineMessage              = "Could not fetch stories • Press [::b]r[::-] to retry or [::b]q[::-] to quit"
	CommentsNotFetched          = "Could not fetch comments"
	ArticleNotFetched           = "Could not fetch article"
	EnterCommentSectionToUpdate = "[Enter comment section to update story]"
	DeleteFromFavorites         = "[red]Delete[-] from Favorites? Press [::b]y[::-] to Confirm"
	ItemDeleted                 = "Item deleted"
	AddToFavorites              = "[green]Add[-] to Favorites? Press [::b]y[::-] to Confirm"
	LessScreenInfo              = "You are now in 'less' • Press 'q' to return and 'h' for help"
	HowToExitF                  = "[::d]Leave ID blank to return to main screen[::-]"
	AddedStoryByID              = "Submission added"
	FavoriteNotAdded            = "Could not add submission to favorites"
	FavoriteAdded               = "Submission added to favorites"
)

func GetCircumflexStatusMessage() string {
	return "[::d]github.com/bensadeh/circumflex • version " + clx.Version
}

func GetSeparator(width int) string {
	return strings.Repeat(separator, width)
}
