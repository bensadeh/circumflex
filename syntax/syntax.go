package syntax

import (
	"clx/colors"
	"regexp"
)

func HighlightYCStartups(comment string) string {
	expression := regexp.MustCompile(`\((YC [SW]\d{2})\)`)

	return expression.ReplaceAllString(comment, colors.OrangeBackground+colors.NearBlack+" "+`$1`+" "+colors.Normal)
}
