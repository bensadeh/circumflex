package list

import (
	"clx/bfavorites"
	"clx/bheader"
	"clx/bubble/list/message"
	"clx/bubble/ranking"
	"clx/cli"
	"clx/comment"
	"clx/constants/category"
	"clx/constants/style"
	"clx/core"
	"clx/help"
	"clx/history"
	"clx/hn"
	hybrid "clx/hn/services/hybrid-bubble"
	"clx/hn/services/mock"
	"clx/item"
	"clx/screen"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	numberOfCategories = 5
)

// Item is an item that appears in the list.
//type Item interface{}

// ItemDelegate encapsulates the general functionality for all list items. The
// benefit to separating this logic from the item itself is that you can change
// the functionality of items without changing the actual items themselves.
//
// Note that if the delegate also implements help.KeyMap delegate-related
// help items will be added to the help view.
type ItemDelegate interface {
	// Render renders the item's view.
	Render(w io.Writer, m Model, index int, item *item.Item)

	// Height is the height of the list item.
	Height() int

	// Spacing is the size of the horizontal gap between list items in cells.
	Spacing() int

	// Update is the update loop for items. All messages in the list's update
	// loop will pass through here except when the user is setting a filter.
	// Use this method to perform item-level updates appropriate to this
	// delegate.
	Update(msg tea.Msg, m *Model) tea.Cmd
}

// Model contains the state of this component.
type Model struct {
	showTitle     bool
	showStatusBar bool
	disableInput  bool

	Title  string
	Styles Styles

	spinner                     spinner.Model
	showSpinner                 bool
	width                       int
	height                      int
	Paginator                   paginator.Model
	cursor                      int
	onStartup                   bool
	isVisible                   bool
	onAddToFavoritesPrompt      bool
	onRemoveFromFavoritesPrompt bool

	StatusMessageLifetime time.Duration

	statusMessage      string
	statusMessageTimer *time.Timer

	category int
	items    [][]*item.Item

	delegate  ItemDelegate
	history   history.History
	config    *core.Config
	service   hn.Service
	favorites *bfavorites.Favorites
}

func (m *Model) FetchFrontPageStories() tea.Cmd {
	return func() tea.Msg {
		firstPage := 0
		stories := m.service.FetchStories(firstPage, category.FrontPage)

		m.items[category.FrontPage] = stories
		return message.FetchingFinished{}
	}
}

func New(delegate ItemDelegate, config *core.Config, favorites *bfavorites.Favorites, width, height int) Model {
	styles := DefaultStyles()

	sp := spinner.New()
	sp.Spinner = getSpinner()
	sp.Style = styles.Spinner

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = styles.ActivePaginationDot.String()
	p.InactiveDot = styles.InactivePaginationDot.String()
	p.UseHLKeys = false
	p.UseJKKeys = false
	p.UseLeftRightKeys = false
	p.UsePgUpPgDownKeys = false
	p.UseUpDownKeys = false

	items := make([][]*item.Item, numberOfCategories)

	m := Model{
		showTitle:             true,
		showStatusBar:         true,
		Styles:                styles,
		Title:                 "List",
		StatusMessageLifetime: time.Second,

		width:        width,
		height:       height,
		delegate:     delegate,
		history:      getHistory(config.DebugMode, config.MarkAsRead),
		items:        items,
		Paginator:    p,
		spinner:      sp,
		onStartup:    true,
		isVisible:    true,
		disableInput: true,
		config:       config,
		service:      getService(config.DebugMode),
		favorites:    favorites,
	}

	m.updatePagination()

	return m
}

func getHistory(debugMode bool, markAsRead bool) history.History {
	if debugMode {
		return history.NewMockHistory()
	}

	if markAsRead {
		return history.NewPersistentHistory()
	}

	return history.NewNonPersistentHistory()
}

func getService(debugMode bool) hn.Service {
	if debugMode {
		return mock.MockService{}
	}

	return &hybrid.Service{}
}

// SetShowTitle shows or hides the title bar.
func (m *Model) SetShowTitle(v bool) {
	m.showTitle = v
	m.updatePagination()
}

func (m *Model) SetIsVisible(v bool) {
	m.isVisible = v
}

// SetShowStatusBar shows or hides the view that displays metadata about the
// list, such as item counts.
func (m *Model) SetShowStatusBar(v bool) {
	m.showStatusBar = v
	m.updatePagination()
}

// ShowStatusBar returns whether or not the status bar is set to be rendered.
func (m Model) ShowStatusBar() bool {
	return m.showStatusBar
}

// Set the items available in the list. This returns a command.
func (m *Model) SetItems(i []*item.Item) tea.Cmd {
	var cmd tea.Cmd
	m.items[m.category] = i

	m.updatePagination()
	return cmd
}

// Select selects the given index of the list and goes to its respective page.
func (m *Model) Select(index int) {
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage
}

// VisibleItems returns the total items available to be shown.
func (m Model) VisibleItems() []*item.Item {
	return m.items[m.category]
}

// SelectedItems returns the current selected item in the list.
func (m Model) SelectedItem() *item.Item {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		//return nil
		return &item.Item{}
	}

	return items[i]
}

// Index returns the index of the currently selected item as it appears in the
// entire slice of items.
func (m Model) Index() int {
	return m.Paginator.Page*m.Paginator.PerPage + m.cursor
}

// Cursor returns the index of the cursor on the current page.
func (m Model) Cursor() int {
	return m.cursor
}

// CursorUp moves the cursor up. This can also move the state to the previous
// page.
func (m *Model) CursorUp() {
	m.cursor--

	// If we're at the top, stop
	if m.cursor < 0 {
		m.cursor = 0
		return
	}

	return
}

// CursorDown moves the cursor down. This can also advance the state to the
// next page.
func (m *Model) CursorDown() {
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))

	m.cursor++

	// If we're at the end, stop
	if m.cursor < itemsOnPage {
		return
	}

	m.cursor = itemsOnPage - 1
}

func (m *Model) getNextCategory() int {
	isAtLastCategory := m.category == m.getNumberOfCategories()-1
	if isAtLastCategory {
		return category.FrontPage
	}

	return m.category + 1
}
func (m *Model) getNumberOfCategories() int {
	if m.favorites.HasItems() {
		return numberOfCategories
	}

	return numberOfCategories - 1
}

func (m *Model) getPrevCategory() int {
	isAtFirstCategory := m.category == category.FrontPage
	if isAtFirstCategory && m.favorites.HasItems() {
		return category.Favorites
	}

	if isAtFirstCategory {
		return category.Show
	}

	return m.category - 1
}

func (m *Model) ToggleSpinner() tea.Cmd {
	if !m.showSpinner {
		return m.StartSpinner()
	}
	m.StopSpinner()
	return nil
}

func (m *Model) StartSpinner() tea.Cmd {
	// Hack: I can't get the spinner to reset properly. As a workaround, we
	// instantiate a new spinner each time we want to show it.
	m.spinner = spinner.New()
	m.spinner.Spinner = getSpinner()
	m.spinner.Style = DefaultStyles().Spinner

	m.showSpinner = true
	return m.spinner.Tick
}

func (m *Model) StopSpinner() {
	m.showSpinner = false
	m.spinner.Finish()
}

func (m *Model) NewStatusMessage(s string) tea.Cmd {
	m.statusMessage = s
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}

	m.statusMessageTimer = time.NewTimer(m.StatusMessageLifetime)

	// Wait for timeout
	return func() tea.Msg {
		<-m.statusMessageTimer.C
		return message.StatusMessageTimeout{}
	}
}

func (m *Model) NewStatusMessageWithDuration(s string, d time.Duration) tea.Cmd {
	m.statusMessage = s
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}

	m.statusMessageTimer = time.NewTimer(d)

	// Wait for timeout
	return func() tea.Msg {
		<-m.statusMessageTimer.C
		return message.StatusMessageTimeout{}
	}
}

func (m *Model) SetPermanentStatusMessage(s string) {
	m.statusMessage = s
}

// SetSize sets the width and height of this component.
func (m *Model) SetSize(width, height int) {
	m.setSize(width, height)
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

// Update pagination according to the amount of items for the current state.
func (m *Model) updatePagination() {
	index := m.Index()
	availHeight := m.height

	if m.showTitle {
		availHeight -= lipgloss.Height(m.titleView())
	}
	if m.showStatusBar {
		availHeight -= lipgloss.Height(m.statusView())
	}

	m.Paginator.PerPage = max(1, availHeight/(m.delegate.Height()+m.delegate.Spacing()))

	if pages := len(m.VisibleItems()); pages < 1 {
		m.Paginator.SetTotalPages(1)
	} else {
		m.Paginator.SetTotalPages(pages)
	}

	// Restore index
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage

	// Make sure the page stays in bounds
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	}
}

func (m *Model) hideStatusMessage() {
	m.statusMessage = ""
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.OnStartup() {
		m.SetSize(screen.GetTerminalWidth(), screen.GetTerminalHeight())

		m.service.Init(m.Paginator.PerPage)

		var cmds []tea.Cmd

		spinnerCmd := m.StartSpinner()
		cmds = append(cmds, spinnerCmd)

		m.SetOnStartup(false)

		m.items[category.Favorites] = m.favorites.GetItems()

		fetchCmd := m.FetchFrontPageStories()
		cmds = append(cmds, fetchCmd)

		return m, tea.Batch(cmds...)
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		newSpinnerModel, cmd := m.spinner.Update(msg)
		m.spinner = newSpinnerModel
		if m.showSpinner {
			cmds = append(cmds, cmd)
		}

	case message.FetchingFinished:
		m.StopSpinner()
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(screen.GetTerminalWidth()-h, screen.GetTerminalHeight()-v)
		m.disableInput = false

		return m, nil

	case message.StatusMessageTimeout:
		m.hideStatusMessage()

	case message.AddToFavorites:
		m.favorites.Add(msg.Item)
		m.items[category.Favorites] = m.favorites.GetItems()

		m.favorites.Write()

		m.updatePagination()

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)

	case message.EnteringCommentSection:
		return m, m.fetchCommentSectionAndPipeToLess(msg.Id, msg.CommentCount)

	case message.EnterHelpScreen:
		return m, m.showHelpScreen()

	case message.EditorFinishedMsg:
		m.SetIsVisible(true)

	case message.ChangeCategory:
		cmd := func() tea.Msg {
			stories := m.service.FetchStories(0, msg.Category)

			m.items[msg.Category] = stories

			return message.CategoryFetchingFinished{Category: msg.Category, Cursor: msg.Cursor}
		}

		return m, cmd

	case message.CategoryFetchingFinished:
		m.Paginator.Page = 0
		m.SetDisabledInput(false)
		m.StopSpinner()
		m.category = msg.Category

		itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
		m.cursor = min(msg.Cursor, itemsOnPage-1)
		m.updatePagination()
	}

	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}

func (m *Model) categoryHasStories(cat int) bool {
	return len(m.items[cat]) != 0
}

func (m *Model) changeToCategory(cat int) {
	m.category = cat
	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[m.category])-1)
	m.updatePagination()
}

func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	numItems := len(m.VisibleItems())

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.onAddToFavoritesPrompt && msg.String() == "y":
			m.onAddToFavoritesPrompt = false
			m.disableInput = false

			addToFavorites := func() tea.Msg {
				return message.AddToFavorites{Item: m.SelectedItem()}
			}

			cmds = append(cmds, addToFavorites)
			cmds = append(cmds, m.NewStatusMessageWithDuration("Item added", time.Second*2))

			return tea.Batch(cmds...)

		case m.onRemoveFromFavoritesPrompt && msg.String() == "y":
			m.onRemoveFromFavoritesPrompt = false
			m.disableInput = false

			//
			m.favorites.Remove(m.Index())
			m.items[category.Favorites] = m.favorites.GetItems()

			m.favorites.Write()

			//
			isOnLastItem := m.Index() == len(m.items[category.Favorites])
			hasOnlyOneItem := len(m.items[category.Favorites]) == 0

			itemRemovedMessage := "Item removed"

			if hasOnlyOneItem {
				m.cursor = 0

				cmds = append(cmds, func() tea.Msg {
					return message.ChangeCategory{Category: category.FrontPage, Cursor: m.cursor}
				})
				cmds = append(cmds, m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2))

				return tea.Batch(cmds...)
			}

			if isOnLastItem {
				m.cursor = m.cursor - 1
			}

			m.updatePagination()

			return m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2)

		case m.onAddToFavoritesPrompt || m.onRemoveFromFavoritesPrompt:
			m.onAddToFavoritesPrompt = false
			m.onRemoveFromFavoritesPrompt = false
			m.disableInput = false

			faint := "\033[2m"

			return m.NewStatusMessageWithDuration(faint+"Cancelled", time.Second*2)

		case m.disableInput && msg.String() != "r":
			return nil

		case msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c":
			return tea.Quit

		case msg.String() == "up" || msg.String() == "k":
			m.CursorUp()

			return nil

		case msg.String() == "down" || msg.String() == "j":
			m.CursorDown()

			return nil

		case msg.String() == "left" || msg.String() == "h":
			m.Paginator.PrevPage()

			return nil

		case msg.String() == "right" || msg.String() == "l":
			m.Paginator.NextPage()

			return nil

		case msg.String() == "tab":
			nextCat := m.getNextCategory()

			if m.categoryHasStories(nextCat) {
				m.changeToCategory(nextCat)

				return nil
			}

			m.SetDisabledInput(true)
			startSpinnerCmd := m.StartSpinner()

			changeCatCmd := func() tea.Msg {
				return message.ChangeCategory{Category: nextCat, Cursor: m.cursor}
			}

			cmds = append(cmds, startSpinnerCmd)
			cmds = append(cmds, changeCatCmd)

			return tea.Batch(cmds...)

		case msg.String() == "shift+tab":
			prevCat := m.getPrevCategory()

			if m.categoryHasStories(prevCat) {
				m.changeToCategory(prevCat)

				return nil
			}

			m.SetDisabledInput(true)
			startSpinnerCmd := m.StartSpinner()

			changeCatCmd := func() tea.Msg {
				return message.ChangeCategory{Category: prevCat, Cursor: m.cursor}
			}

			cmds = append(cmds, startSpinnerCmd)
			cmds = append(cmds, changeCatCmd)

			return tea.Batch(cmds...)

		case msg.String() == "g":
			m.Paginator.Page = 0
			m.cursor = 0

			return nil

		case msg.String() == "G":
			m.Paginator.Page = m.Paginator.TotalPages - 1
			m.cursor = m.Paginator.ItemsOnPage(numItems) - 1

			return nil

		case msg.String() == "e":
			cmd := m.NewStatusMessageWithDuration("Test", 2*time.Second)

			return cmd

		case msg.String() == "r":
			cmd := m.NewStatusMessageWithDuration("Input enabled", 1*time.Second)
			m.disableInput = false
			m.onAddToFavoritesPrompt = false

			return cmd

		case msg.String() == "f":
			m.SetPermanentStatusMessage("Add to favorites?")
			m.onAddToFavoritesPrompt = true
			m.disableInput = true

			return nil

		case msg.String() == "x" && m.category == category.Favorites:
			m.SetPermanentStatusMessage("Remove from favorites?")
			m.onRemoveFromFavoritesPrompt = true
			m.disableInput = true

			return nil

		case msg.String() == "enter":
			m.SetIsVisible(false)

			cmd := func() tea.Msg {
				return message.EnteringCommentSection{
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

			return cmd

		case msg.String() == "i":
			m.SetIsVisible(false)

			cmd := func() tea.Msg {
				return message.EnterHelpScreen{}
			}

			return cmd

		case msg.String() == "u":
			cmd := m.StartSpinner()

			return cmd

		case msg.String() == "i":
			cmd := m.NewStatusMessageWithDuration("is disabled: "+strconv.FormatBool(m.IsInputDisabled()), 1*time.Second)

			return cmd
		}
	}

	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// Keep the index in bounds when paginating
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.cursor > itemsOnPage-1 {
		m.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

func (m *Model) fetchCommentSectionAndPipeToLess(id int, commentCount int) tea.Cmd {
	lastVisited := m.history.GetLastVisited(id)

	m.history.AddToHistoryAndWriteToDisk(id, commentCount)

	comments := m.service.FetchStory(id)

	commentTree := comment.ToString(comments, m.config, m.width, lastVisited)

	command := cli.WrapLess(commentTree)

	return tea.ExecProcess(command, func(err error) tea.Msg {
		return message.EditorFinishedMsg{Err: err}
	})
}

func (m *Model) showHelpScreen() tea.Cmd {
	helpScreen := help.GetHelpScreen()

	command := cli.WrapLess(helpScreen)

	return tea.ExecProcess(command, func(err error) tea.Msg {
		return message.EditorFinishedMsg{Err: err}
	})
}

// View renders the component.
func (m Model) View() string {
	var (
		sections    []string
		availHeight = m.height
	)

	if !m.isVisible {
		return ""
	}

	if m.showTitle {
		v := m.titleView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	if m.showStatusBar {
		v := m.statusView()
		availHeight -= lipgloss.Height(v)
	}

	content := lipgloss.NewStyle().Height(availHeight).Render(m.populatedView())
	rankings := ranking.GetRankings(false, m.Paginator.PerPage, len(m.items[m.category]), m.cursor,
		m.Paginator.Page, m.Paginator.TotalPages)

	rankingsAndContent := lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
	sections = append(sections, rankingsAndContent)

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		sections = append(sections, v)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) titleView() string {
	return bheader.GetHeader(m.category, m.favorites.HasItems(), m.width) + "\n"
}

func (m Model) statusAndPaginationView() string {
	centerContent := ""

	if m.showSpinner {
		centerContent = m.spinnerView()
	} else {
		centerContent = m.statusMessage
	}

	left := lipgloss.NewStyle().Inline(true).
		Background(style.GetStatusBarBg()).
		Width(5).MaxWidth(5).Render("")

	center := lipgloss.NewStyle().Inline(true).
		Background(style.GetStatusBarBg()).
		Width(m.width - 5 - 5).Align(lipgloss.Center).Render(centerContent)

	right := lipgloss.NewStyle().Inline(true).
		Background(style.GetPaginatorBg()).
		Width(5).Align(lipgloss.Center).Render(m.Paginator.View())

	return m.Styles.StatusBar.Render(left) + m.Styles.StatusBar.Render(center) + m.Styles.StatusBar.Render(right)
}

func (m Model) statusView() string {
	var status string

	visibleItems := len(m.VisibleItems())

	plural := ""
	if visibleItems != 1 {
		plural = "s"
	}

	if len(m.items) == 0 {
		status = m.Styles.StatusEmpty.Render("")
	} else {
		status += fmt.Sprintf("%d item%s", visibleItems, plural)
	}

	return m.Styles.StatusBar.Render(status)
}

func (m Model) OnStartup() bool {
	return m.onStartup
}

func (m *Model) IsInputDisabled() bool {
	return m.disableInput
}

func (m *Model) SetDisabledInput(value bool) {
	m.disableInput = value
}

func (m *Model) SetOnStartup(value bool) {
	m.onStartup = value
}

func (m Model) populatedView() string {
	items := m.VisibleItems()

	var b strings.Builder

	// Empty states
	if len(items) == 0 {
		return m.Styles.NoItems.Render("")
	}

	if len(items) > 0 {
		start, end := m.Paginator.GetSliceBounds(len(items))
		docs := items[start:end]

		for i, item := range docs {
			m.delegate.Render(&b, m, i+start, item)
			if i != len(docs)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.Paginator.ItemsOnPage(len(items))
	if itemsOnPage < m.Paginator.PerPage {
		n := (m.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(items) == 0 {
			n -= m.delegate.Height() - 1
		}
		fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func (m Model) spinnerView() string {
	return m.spinner.View()
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
