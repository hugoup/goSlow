# goSlow: MySQL Slow Query Log TUI

![goSlow Banner](https://img.shields.io/badge/TUI-Go-blue?style=flat-square)

**goSlow** is a blazing-fast, modern terminal user interface (TUI) for exploring and analyzing MySQL slow query logs. Designed for database engineers, developers, and performance enthusiasts, goSlow makes it effortless to find, sort, and inspect problematic queries—right from your terminal.

---

## 🚀 Features

- **Instant Grouping:** Automatically groups similar slow queries for easy analysis.
- **Interactive Table:** Navigate, sort, and filter queries with keyboard shortcuts.
- **Preview Panel:** View full SQL text and details for any query group.
- **Syntax Highlighting:** Custom, fast highlighting for SQL (no heavy dependencies).
- **Sort Modal:** Quickly sort by count, average time, rows examined, and more.
- **Help Panel:** Built-in help for all key bindings and features.
- **Export:** Save any query to a `.sql` file with a single keystroke.
- **Modern UI:** Clean, responsive, and visually appealing TUI.

---

## 🖥️ Screenshots

![goSlow Screenshot](https://raw.githubusercontent.com/hugoup/goSlow/main/screenshots/goSlow.png)

---

## ⌨️ Key Bindings

| Key         | Action                                 |
|-------------|----------------------------------------|
| ↑/↓         | Move selection                         |
| Tab         | Switch focus (table/preview)           |
| Enter       | Preview selected query group           |
| s           | Save selected query to file            |
| l           | Open sort modal                        |
| h           | Toggle SQL highlighting                |
| z           | Zoom preview panel                     |
| q / Ctrl+C  | Quit                                   |

## 🛠️ Requirements
- MySQL-compatible database with `slow_log` table enabled
- Running on localhost:3306 for now
