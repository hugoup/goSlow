package ui

import (
	"fmt"
	"sort"

	"slowlog-tui/types"

	"github.com/charmbracelet/bubbles/table"
)

// TablePanel handles the grouped queries table logic
// It is stateless; state is managed by the main Model
func NewTablePanel(filteredGroups []types.GroupedQuery, tableWidth, tableHeight int) table.Model {
	var rows []table.Row
	for i, g := range filteredGroups {
		db := ""
		if len(g.Examples) > 0 {
			db = g.Examples[0].DB
		}
		tableName := g.FromTable
		minOtherCols := 4 + 8 + 24 + 16 + 8 + 10 + 12 + 10 + 8
		maxShortQuery := tableWidth - minOtherCols
		if maxShortQuery > 50 {
			maxShortQuery = 50
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

	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)
	tbl.SetStyles(table.Styles{
		Selected: selectedRowStyle,
	})
	return tbl
}

// SortGroups sorts the filteredGroups slice in-place by the given column and order
func SortGroups(groups []types.GroupedQuery, sortColumn, sortOrder int) {
	less := func(i, j int) bool { return false }
	switch sortColumn {
	case 0: // Count
		less = func(i, j int) bool {
			if groups[i].Count == groups[j].Count {
				return groups[i].AvgQueryTime > groups[j].AvgQueryTime
			}
			return groups[i].Count > groups[j].Count
		}
	case 1: // Avg Time
		less = func(i, j int) bool {
			return groups[i].AvgQueryTime > groups[j].AvgQueryTime
		}
	case 2: // Avg Examined
		less = func(i, j int) bool {
			return groups[i].AvgRowsExamined > groups[j].AvgRowsExamined
		}
	case 3: // Avg Sent
		less = func(i, j int) bool {
			return groups[i].AvgRowsSent > groups[j].AvgRowsSent
		}
	case 4: // Type
		less = func(i, j int) bool {
			return groups[i].QueryType < groups[j].QueryType
		}
	case 5: // DB
		less = func(i, j int) bool {
			var dbi, dbj string
			if len(groups[i].Examples) > 0 {
				dbi = groups[i].Examples[0].DB
			}
			if len(groups[j].Examples) > 0 {
				dbj = groups[j].Examples[0].DB
			}
			return dbi < dbj
		}
	case 6: // Table
		less = func(i, j int) bool {
			return groups[i].FromTable < groups[j].FromTable
		}
	}
	if sortOrder == 0 {
		sort.Slice(groups, less)
	} else {
		sort.Slice(groups, func(i, j int) bool { return less(j, i) })
	}
}
