package primitives

import (
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
)

const (
	helpPage    = "help"
	offlinePage = "offline"
)

type MainView struct {
	Pages       *cview.Pages
	Grid        *cview.Grid
	Footer      *cview.TextView
	Header      *cview.TextView
	LeftMargin  *cview.TextView
	RightMargin *cview.TextView
}

func NewMainView(screenWidth int, viewableStoriesOnSinglePage int) *MainView {
	footerText := getFooterText(0, screenWidth)
	headlineText := getHeadline(screenWidth)

	main := new(MainView)
	main.Pages = cview.NewPages()
	main.Grid = cview.NewGrid()
	main.LeftMargin = newTextViewPrimitive("")
	main.LeftMargin.SetTextAlign(cview.AlignRight)
	main.SetLeftMarginRanks(0, viewableStoriesOnSinglePage)
	main.RightMargin = newTextViewPrimitive("")
	main.Header = newTextViewPrimitive(headlineText)
	main.Footer = newTextViewPrimitive(footerText)

	main.Grid.SetBorder(false)
	main.Grid.SetRows(2, 0, 1)
	main.Grid.SetColumns(7, 0, 3)
	main.Grid.SetBackgroundColor(tcell.ColorDefault)
	main.Grid.AddItem(main.Header, 0, 0, 1, 3, 0, 0, false)
	main.Grid.AddItem(main.Footer, 2, 0, 1, 3, 0, 0, false)
	main.Grid.AddItem(main.LeftMargin, 1, 0, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.Pages, 1, 1, 1, 1, 0, 0, true)
	main.Grid.AddItem(main.RightMargin, 1, 2, 1, 1, 0, 0, false)

	main.Pages.AddPage(helpPage, GetHelpScreen(), true, false)
	main.Pages.AddPage(offlinePage, GetOfflineScreen(), true, false)

	return main
}

func newTextViewPrimitive(text string) *cview.TextView {
	tv := cview.NewTextView()
	tv.SetTextAlign(cview.AlignLeft)
	tv.SetText(text)
	tv.SetBorder(false)
	tv.SetBackgroundColor(tcell.ColorDefault)
	tv.SetTextColor(tcell.ColorDefault)
	tv.SetDynamicColors(true)
	return tv
}

func getHeadline(screenWidth int) string {
	base := "[black:orange:]   [Y[] [::b]Hacker News"
	offset := -21
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	return base + whitespace
}

func (m MainView) SetHeaderTextToHN(screenWidth int) {
	m.Header.SetText(getHeadline(screenWidth))
}

func (m MainView) SetHeaderTextToKeymaps(screenWidth int) {
	base := "[white:rebeccapurple:]   [^] [::b]Keymaps"
	offset := -27
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	m.Header.SetText(base + whitespace)
}

func (m MainView) SetFooterText(currentPage int, screenWidth int) {
	footerText := getFooterText(currentPage, screenWidth)
	m.Footer.SetText(footerText)
}

func (m MainView) HideFooterText() {
	m.Footer.SetText("")
}

func (m MainView) SetLeftMarginRanks(currentPage int, viewableStoriesOnSinglePage int) {
	marginText := ""
	indentationFromRight := " "
	startingRank := viewableStoriesOnSinglePage*currentPage + 1
	for i := startingRank; i < startingRank+viewableStoriesOnSinglePage; i++ {
		marginText += strconv.Itoa(i) + "." + indentationFromRight + "\n\n"
	}
	m.LeftMargin.SetText(marginText)
}

func (m MainView) HideLeftMarginRanks() {
	m.LeftMargin.SetText("")
}

func getFooterText(currentPage int, screenWidth int) string {
	orangeDot := "[orange]" + "•" + "[-:-]"
	footerText := ""

	switch currentPage {
	case 0:
		footerText = "" + orangeDot + "◦◦◦"
	case 1:
		footerText = "◦" + orangeDot + "◦◦"
	case 2:
		footerText = "◦◦" + orangeDot + "◦"
	case 3:
		footerText = "◦◦◦" + orangeDot + ""
	default:
		footerText = ""
	}
	return padWithWhitespaceFromTheLeft(footerText, screenWidth)
}

func padWithWhitespaceFromTheLeft(s string, screenWidth int) string {
	offset := +10
	whitespace := ""
	for i := 0; i < screenWidth-text.Len(s)+offset; i++ {
		whitespace += " "
	}
	return whitespace + s
}
