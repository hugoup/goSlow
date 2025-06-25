package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"
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
	boldStyle      = lipgloss.NewStyle().Bold(true)

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

func (m *Model) applyFilters(tableWidth int) {
	m.filteredGroups = m.allGroups
	// Sort by selected column and order
	less := func(i, j int) bool { return false }
	switch m.sortColumn {
	case 0: // Count
		less = func(i, j int) bool {
			if m.filteredGroups[i].Count == m.filteredGroups[j].Count {
				return m.filteredGroups[i].AvgQueryTime > m.filteredGroups[j].AvgQueryTime
			}
			return m.filteredGroups[i].Count > m.filteredGroups[j].Count
		}
	case 1: // Avg Time
		less = func(i, j int) bool {
			return m.filteredGroups[i].AvgQueryTime > m.filteredGroups[j].AvgQueryTime
		}
	case 2: // Avg Examined
		less = func(i, j int) bool {
			return m.filteredGroups[i].AvgRowsExamined > m.filteredGroups[j].AvgRowsExamined
		}
	case 3: // Avg Sent
		less = func(i, j int) bool {
			return m.filteredGroups[i].AvgRowsSent > m.filteredGroups[j].AvgRowsSent
		}
	case 4: // Type
		less = func(i, j int) bool {
			return m.filteredGroups[i].QueryType < m.filteredGroups[j].QueryType
		}
	case 5: // DB
		less = func(i, j int) bool {
			var dbi, dbj string
			if len(m.filteredGroups[i].Examples) > 0 {
				dbi = m.filteredGroups[i].Examples[0].DB
			}
			if len(m.filteredGroups[j].Examples) > 0 {
				dbj = m.filteredGroups[j].Examples[0].DB
			}
			return dbi < dbj
		}
	case 6: // Table
		less = func(i, j int) bool {
			return m.filteredGroups[i].FromTable < m.filteredGroups[j].FromTable
		}
	}
	if m.sortOrder == 0 {
		sort.Slice(m.filteredGroups, less)
	} else {
		sort.Slice(m.filteredGroups, func(i, j int) bool { return less(j, i) })
	}
	var rows []table.Row
	for i, g := range m.filteredGroups {
		db := ""
		if len(g.Examples) > 0 {
			db = g.Examples[0].DB
		}
		tableName := g.FromTable
		// Dynamically calculate max width for shortQuery
		minOtherCols := 4 + 8 + 24 + 16 + 8 + 10 + 12 + 10 + 8 // sum of fixed col widths + padding (added 16 for Table col)
		maxShortQuery := tableWidth - minOtherCols
		if maxShortQuery > 50 {
			maxShortQuery = 50 // hard cap to avoid wrapping
		}
		if maxShortQuery < 10 {
			maxShortQuery = 10
		}
		shortQuery := g.NormalizedSQL
		if len(shortQuery) > maxShortQuery {
			shortQuery = shortQuery[:maxShortQuery-3] + "..."
		}
		row := table.Row{
			fmt.Sprintf("%d", i+1),
			g.QueryType,
			db,
			tableName,
			fmt.Sprintf("%d", g.Count),
			fmt.Sprintf("%.2fs", g.AvgQueryTime),
			fmt.Sprintf("%.0f", g.AvgRowsExamined),
			fmt.Sprintf("%.0f", g.AvgRowsSent),
			shortQuery,
		}
		rows = append(rows, row)
	}

	cols := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Type", Width: 8},
		{Title: "DB", Width: 24},
		{Title: "Table", Width: 26},
		{Title: "Count", Width: 8},
		{Title: "Avg Time", Width: 10},
		{Title: "Avg Examined", Width: 12},
		{Title: "Avg Sent", Width: 10},
		{Title: "Query", Width: 50},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height/2-2),
	)
	// Set custom style for selected row
	t.SetStyles(table.Styles{
		Selected: selectedRowStyle,
	})
	m.table = t
}

func (m *Model) updateViewport() {
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filteredGroups) {
		g := m.filteredGroups[cursor]
		header := fmt.Sprintf("%s | %d queries | Avg: %.2fs, %.0f rows examined, %.0f sent\n\n",
			boldStyle.Render(g.QueryType),
			g.Count,
			g.AvgQueryTime,
			g.AvgRowsExamined,
			g.AvgRowsSent,
		)
		var allQueries strings.Builder
		for i, q := range g.Examples {
			allQueries.WriteString(q.SQLText)
			if i < len(g.Examples)-1 {
				allQueries.WriteString("\n---\n")
			}
		}
		start := time.Now()
		var content string
		switch m.highlightMode {
		case HighlightSimple:
			content = HighlightSQL(allQueries.String())
		case HighlightOff:
			content = allQueries.String()
		}
		dur := time.Since(start)
		m.statusText = fmt.Sprintf("Render: %dms", dur.Milliseconds())
		m.statusColor = lipgloss.Color("#ffaf00") // orange for timing
		m.viewport.SetContent(header + content)
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

func flashStatus() tea.Cmd {
	return func() tea.Msg {
		return flashStatusMsg{}
	}
}

type flashStatusMsg struct{}

var helpOptions = []struct{ Key, Desc string }{
	{"↑/↓", "Scroll"},
	{"↵", "Show Queries"},
	{"Tab", "Switch panel"},
	{"l", "Sort"},
	{"s", "Save queries"},
	{"z", "Zoom"},
	{"h", "HL-mode"},
	{"q", "Quit"},
}

func (m Model) View() string {
	panelWidth := m.viewport.Width // use the actual viewport width for all panels

	if m.showSortModal {
		modalWidth := 60
		modalHeight := len(m.sortColumns) + 6
		var b strings.Builder
		b.WriteString("Sort by column and order:\n\n")
		b.WriteString("Column" + strings.Repeat(" ", 28) + "Order\n")
		b.WriteString(strings.Repeat("-", modalWidth-4) + "\n")
		maxRows := len(m.sortColumns)
		if maxRows < 2 {
			maxRows = 2
		}
		for i := 0; i < maxRows; i++ {
			// Column radio
			colRadio := "  "
			if m.sortModalCursor == i && m.sortModalFocus == 0 {
				colRadio = "▶ " // focused
			} else {
				colRadio = "  "
			}
			if i < len(m.sortColumns) {
				selected := "○"
				if m.sortColumn == i {
					selected = "●"
				}
				colRadio += selected + " " + m.sortColumns[i]
			} else {
				colRadio = ""
			}
			// Order radio
			orderRadio := "  "
			if m.sortOrder == i && m.sortModalFocus == 1 {
				orderRadio = "▶ "
			} else {
				orderRadio = "  "
			}
			if i < 2 {
				selected := "○"
				if m.sortOrder == i {
					selected = "●"
				}
				orderName := "Ascending"
				if i == 1 {
					orderName = "Descending"
				}
				orderRadio += selected + " " + orderName
			}
			b.WriteString(fmt.Sprintf("%-35s   %-20s\n", colRadio, orderRadio))
		}
		b.WriteString("\n[Tab] Switch  [↑/↓] Move  [Enter] Apply  [Esc] Cancel")
		modal := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(modalWidth).Height(modalHeight).Align(lipgloss.Center).Render(b.String())
		padTop := (m.height - modalHeight) / 2
		padLeft := (panelWidth - modalWidth) / 2
		if padTop < 0 {
			padTop = 0
		}
		if padLeft < 0 {
			padLeft = 0
		}
		modal = strings.Repeat("\n", padTop) + lipgloss.NewStyle().MarginLeft(padLeft).Render(modal)
		return modal
	}

	// Render main UI (table + preview + help) as usual
	tableContent := m.table.View()
	var tableBoxStyle, sqlBoxStyle lipgloss.Style
	if m.focus == focusTable {
		tableBoxStyle = leftStyle.BorderForeground(activeBorder)
		sqlBoxStyle = rightStyle.BorderForeground(inactiveBorder)
	} else {
		tableBoxStyle = leftStyle.BorderForeground(inactiveBorder)
		sqlBoxStyle = rightStyle.BorderForeground(activeBorder)
	}
	tableBox := tableBoxStyle.Width(panelWidth).Render(tableContent)
	sqlBox := sqlBoxStyle.Width(panelWidth).Render(m.viewport.View())

	highlightStatus := "[h] Highlight: "
	switch m.highlightMode {
	case HighlightSimple:
		highlightStatus += "ON"
	case HighlightOff:
		highlightStatus += "OFF"
	}

	helpParts := make([]string, len(helpOptions))
	for i, opt := range helpOptions {
		helpParts[i] = fmt.Sprintf("[%s] %s", opt.Key, opt.Desc)
	}
	helpText := strings.Join(helpParts, "  ") + "  " + highlightStatus
	status := m.statusText
	statusColor := m.statusColor
	if status == "" {
		status = ""
		statusColor = ""
	}
	space := panelWidth - lipgloss.Width(helpText) - lipgloss.Width(status) - 4 // 4 for border padding
	if space < 1 {
		space = 1
	}
	statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	helpLine := helpText + strings.Repeat(" ", space) + statusStyled
	helpBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(panelWidth).Height(1).Render(helpLine)

	mainUI := appStyle.Margin(0, 0).Render(
		tableBox + "\n" + sqlBox + "\n" + helpBox,
	)

	// Restore true zoom: if zoomed, show only the preview panel fullscreen
	if m.zoomed {
		zoomBoxStyle := rightStyle.BorderForeground(activeBorder)
		m.viewport.Height = m.height - 3 // use all available rows minus help/status line
		zoomBox := zoomBoxStyle.Width(panelWidth).Height(m.viewport.Height).Render(m.viewport.View())
		// Help/status line
		helpText := "[z] Unzoom  [h] Highlight  [q] Quit"
		status := m.statusText
		statusColor := m.statusColor
		if status == "" {
			status = ""
			statusColor = ""
		}
		space := panelWidth - lipgloss.Width(helpText) - lipgloss.Width(status) - 4
		if space < 1 {
			space = 1
		}
		statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
		helpLine := helpText + strings.Repeat(" ", space) + statusStyled
		helpBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(panelWidth).Height(1).Render(helpLine)
		mainUI = appStyle.Margin(0, 0).Render(zoomBox + "\n" + helpBox)
	}

	return mainUI
}

type clearStatusMsg struct{}
