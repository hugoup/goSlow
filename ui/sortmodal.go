package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SortModalState struct {
	SortColumns     []string
	SortColumn      int
	SortOrder       int
	SortModalCursor int
	SortModalFocus  int
	Height          int
	Width           int
	PanelWidth      int
}

func RenderSortModal(state SortModalState) string {
	modalWidth := state.Width
	modalHeight := len(state.SortColumns) + 6
	var b strings.Builder
	b.WriteString("Sort by column and order:\n\n")
	b.WriteString("Column" + strings.Repeat(" ", 28) + "Order\n")
	b.WriteString(strings.Repeat("-", modalWidth-4) + "\n")
	maxRows := len(state.SortColumns)
	if maxRows < 2 {
		maxRows = 2
	}
	for i := 0; i < maxRows; i++ {
		// Column radio
		colRadio := "  "
		if state.SortModalCursor == i && state.SortModalFocus == 0 {
			colRadio = "▶ " // focused
		} else {
			colRadio = "  "
		}
		if i < len(state.SortColumns) {
			selected := "○"
			if state.SortColumn == i {
				selected = "●"
			}
			colRadio += selected + " " + state.SortColumns[i]
		} else {
			colRadio = ""
		}
		// Order radio
		orderRadio := "  "
		if state.SortOrder == i && state.SortModalFocus == 1 {
			orderRadio = "▶ "
		} else {
			orderRadio = "  "
		}
		if i < 2 {
			selected := "○"
			if state.SortOrder == i {
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
	padTop := (state.Height - modalHeight) / 2
	padLeft := (state.PanelWidth - modalWidth) / 2
	if padTop < 0 {
		padTop = 0
	}
	if padLeft < 0 {
		padLeft = 0
	}
	modal = strings.Repeat("\n", padTop) + lipgloss.NewStyle().MarginLeft(padLeft).Render(modal)
	return modal
}
