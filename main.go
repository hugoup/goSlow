package main

import (
	"fmt"
	"os"

	"slowlog-tui/db"
	"slowlog-tui/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dsn := "root:test123@tcp(127.0.0.1:3306)/mysql"
	queries, err := db.FetchSlowQueries(dsn)

	if err != nil {
		fmt.Println("-> Error loading slow log:", err)
		os.Exit(1)
	}

	fmt.Printf("-> Loaded %d grouped slow queries\n", len(queries))
	if len(queries) == 0 {
		fmt.Println("->  No slow queries found in mysql.slow_log")
		return
	}

	model := ui.NewModel(queries)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("-> Error running TUI:", err)
		os.Exit(1)
	}
}
