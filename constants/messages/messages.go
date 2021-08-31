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
	LessScreenInfo              = "You are now in 'less' • Press q to return or h for help"
	LessCommentInfo             = "d/u to scroll half page • n/N to move between top-level comments"
	LessArticleInfo             = "d/u to scroll half page • n/N to move between headlines"
	FavoriteNotAdded            = "Could not add submission to favorites"
	FavoriteAdded               = "Submission added to favorites"
)

func GetCircumflexStatusMessage() string {
	return "[::d]press ?/i to return • github.com/bensadeh/circumflex • version " + clx.Version
}

func GetSeparator(width int) string {
	return strings.Repeat(separator, width)
}
