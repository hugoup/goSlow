package ui

import (
	"fmt"
	"os"
	"time"

	"slowlog-tui/types"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	leftStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	rightStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	activeBorder   = lipgloss.Color("#00afff") // cyan blue for active
	inactiveBorder = lipgloss.Color("#444444") // gray for inactive
	appStyle       = lipgloss.NewStyle().Margin(1, 1)

	selectedRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a")).Background(lipgloss.Color("#00afff")).Bold(true)
)

type focusArea int

const (
	focusTable focusArea = iota
	focusPreview
)

type HighlightMode int

const (
	HighlightOff HighlightMode = iota
	HighlightSimple
)

type Model struct {
	table          table.Model
	allGroups      []types.GroupedQuery
	filteredGroups []types.GroupedQuery
	viewport       viewport.Model
	focus          focusArea
	height         int
	lastCursor     int
	statusText     string         // for flash/status messages
	statusColor    lipgloss.Color // color for status message
	highlightMode  HighlightMode  // 0=off, 1=simple
	zoomed         bool           // fullscreen preview mode

	// Sorting modal state
	showSortModal   bool
	sortColumn      int
	sortModalCursor int
	sortColumns     []string
	sortOrder       int // 0=asc, 1=desc
	sortModalFocus  int // 0=columns, 1=order
}

func NewModel(groups []types.GroupedQuery) Model {
	m := Model{
		allGroups:      groups,
		focus:          focusTable,
		lastCursor:     -1,
		highlightMode:  HighlightSimple, // default to simple highlighter
		sortColumn:     0,
		sortColumns:    []string{"Count", "Avg Time", "Avg Examined", "Avg Sent", "Type", "DB", "Table"},
		sortOrder:      0,
		sortModalFocus: 0,
	}
	m.viewport = viewport.New(1, 20)
	return m
}

// Remove table logic from applyFilters, use tablepanel.go
func (m *Model) applyFilters(tableWidth int) {
	m.filteredGroups = m.allGroups
	SortGroups(m.filteredGroups, m.sortColumn, m.sortOrder)
	m.table = NewTablePanel(m.filteredGroups, tableWidth, m.height/2-2)
}

func (m *Model) updateViewport() {
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filteredGroups) {
		g := m.filteredGroups[cursor]
		// Use NewPreviewPanel for preview logic
		m.viewport = NewPreviewPanel(g, int(m.highlightMode), m.viewport.Width, m.viewport.Height)
		m.statusText = ""
		m.statusColor = ""
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focus == focusTable {
				m.focus = focusPreview
				m.table.Blur()
			} else {
				m.focus = focusTable
				m.table.Focus()
			}
		case "s":
			i := m.table.Cursor()
			if i >= 0 && i < len(m.filteredGroups) {
				f, _ := os.Create(fmt.Sprintf("query_%d.sql", i+1))
				defer f.Close()
				f.WriteString(m.filteredGroups[i].Examples[0].SQLText)
				m.statusText = "Query saved!"
				m.statusColor = lipgloss.Color("#00d700") // green for success
				return m, flashStatus()
			}
		case "enter":
			if m.focus == focusTable {
				m.lastCursor = m.table.Cursor()
				m.updateViewport()
				m.viewport.GotoTop() // reset scroll position to top
			}
		case "h":
			m.highlightMode = (m.highlightMode + 1) % 2
			m.updateViewport()
		case "z":
			m.zoomed = !m.zoomed
			if m.zoomed {
				m.focus = focusPreview
				m.table.Blur()
			} else {
				m.focus = focusTable
				m.table.Focus()
			}
		case "l":
			m.showSortModal = true
			return m, nil
		}
		if m.showSortModal {
			switch msg.String() {
			case "tab":
				m.sortModalFocus = 1 - m.sortModalFocus
			case "left", "h":
				m.sortModalFocus = 0
			case "right", "l":
				m.sortModalFocus = 1
			case "up", "k":
				if m.sortModalFocus == 0 && m.sortModalCursor > 0 {
					m.sortModalCursor--
					m.sortColumn = m.sortModalCursor // move selection with cursor
				}
				if m.sortModalFocus == 1 && m.sortOrder > 0 {
					m.sortOrder--
				}
			case "down", "j":
				if m.sortModalFocus == 0 && m.sortModalCursor < len(m.sortColumns)-1 {
					m.sortModalCursor++
					m.sortColumn = m.sortModalCursor // move selection with cursor
				}
				if m.sortModalFocus == 1 && m.sortOrder < 1 {
					m.sortOrder++
				}
			case "enter":
				m.showSortModal = false
				m.applyFilters(m.viewport.Width)
				return m, nil
			case "esc":
				m.showSortModal = false
				return m, nil
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		panelWidth := msg.Width - 2 // account for border width
		helpPanelHeight := 8        // height for help panel + border and margin
		remainingHeight := msg.Height - helpPanelHeight
		tablePanelHeight := remainingHeight / 2
		previewPanelHeight := remainingHeight - tablePanelHeight

		m.applyFilters(panelWidth)
		m.table.SetHeight(tablePanelHeight)
		m.viewport.Width = panelWidth
		m.viewport.Height = previewPanelHeight
		m.lastCursor = -1 // force viewport update
	case flashStatusMsg:
		// Clear status after a short delay
		return m, tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return clearStatusMsg{}
		})
	case clearStatusMsg:
		m.statusText = ""
		m.statusColor = ""
	}

	if m.focus == focusTable {
		m.table, cmd = m.table.Update(msg)
		// Only update viewport on enter, not on cursor move
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// Add View method to Model to satisfy tea.Model interface
func (m Model) View() string {
	if m.showSortModal {
		return RenderSortModalView(m)
	}
	if m.zoomed {
		return RenderZoomedPreviewView(m)
	}
	return RenderMainUIView(m)
}

func flashStatus() tea.Cmd {
	return func() tea.Msg {
		return flashStatusMsg{}
	}
}

type flashStatusMsg struct{}

type clearStatusMsg struct{}
