package settings

const (
	ConfigFileNameAbbreviated = "config"
	ConfigFileNameFull        = "config.env"

	CommentWidthName          = "Comment Width"
	CommentWidthKey           = "CLX_COMMENT_WIDTH"
	CommentWidthDefault       = 70
	CommentWidthDescription   = "Sets the maximum number of characters on each line for comments, " +
		"replies, root submission comments and descriptions in settings. Set to 0 to use the whole screen."
	IndentSizeName        = "Indent Size"
	IndentSizeKey         = "CLX_INDENT_SIZE"
	IndentSizeDefault     = 4
	IndentSizeDescription = "The number of whitespaces prepended to each reply, not included the " +
		"color bar to the left of each reply."
	PreserveRightMarginName        = "Preserve Right Margin"
	PreserveRightMarginKey         = "CLX_PRESERVE_RIGHT_MARGIN"
	PreserveRightMarginDefault     = false
	PreserveRightMarginDescription = "Shortens replies so that the total length, including indentation, is the same as " +
		"the comment width. "
)
