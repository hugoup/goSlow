package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))              // blue, no bold
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))              // yellow
	eqStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))              // red
	numStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))              // magenta
	commentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true) // gray
)

var (
	keywords = []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE", "CREATE", "TABLE", "PRIMARY", "KEY", "NOT", "NULL", "DEFAULT", "ON", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "GROUP", "BY", "ORDER", "LIMIT", "AS", "DISTINCT", "UNION", "ALL", "EXISTS", "IN", "IS", "LIKE", "BETWEEN", "CASE", "WHEN", "THEN", "ELSE", "END", "DESC", "ASC",
	}
	keywordRegex = regexp.MustCompile(`(?i)\b(` + strings.Join(keywords, "|") + `)\b`)
	stringRegex  = regexp.MustCompile(`'[^']*'|"[^"]*"`)
	eqRegex      = regexp.MustCompile(`=`)
	numRegex     = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	commentRegex = regexp.MustCompile(`--.*?$|/\*.*?\*/`)
)

// HighlightSQL applies minimal coloring to SQL code for TUI display.
func HighlightSQL(sql string) string {
	// Highlight comments first
	sql = commentRegex.ReplaceAllStringFunc(sql, func(m string) string {
		return commentStyle.Render(m)
	})
	// Highlight strings
	sql = stringRegex.ReplaceAllStringFunc(sql, func(m string) string {
		return stringStyle.Render(m)
	})
	// Highlight numbers
	sql = numRegex.ReplaceAllStringFunc(sql, func(m string) string {
		return numStyle.Render(m)
	})
	// Highlight keywords
	sql = keywordRegex.ReplaceAllStringFunc(sql, func(m string) string {
		return keywordStyle.Render(strings.ToUpper(m))
	})
	// Highlight equal signs
	sql = eqRegex.ReplaceAllStringFunc(sql, func(m string) string {
		return eqStyle.Render(m)
	})
	return sql
}
