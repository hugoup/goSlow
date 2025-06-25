package db

import (
	"database/sql"
	"log"
	"regexp"
	"slowlog-tui/types"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// FormatSQLForDisplay inserts a newline before common SQL clauses for readability, but only if not already at line start
func FormatSQLForDisplay(sqlText string) string {
	clauses := []string{
		"ORDER BY", "GROUP BY", "HAVING", "LIMIT", "WHERE",
		// "LEFT JOIN", "RIGHT JOIN", "INNER JOIN","OUTER JOIN", "JOIN", "UNION", "EXCEPT", "INTERSECT", "RETURNING", "VALUES", "SET"
	}
	for _, clause := range clauses {
		// Regex: find clause with word boundary, case-insensitive
		pattern := `(?i)\b` + regexp.QuoteMeta(clause) + `\b`
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(sqlText, -1)
		if len(matches) == 0 {
			continue
		}
		// Insert newlines before clause if not at start of line or string
		offset := 0
		for _, match := range matches {
			start := match[0] + offset
			if start == 0 || sqlText[start-1] == '\n' {
				continue // already at start of string or line
			}
			sqlText = sqlText[:start] + "\n" + sqlText[start:]
			offset++ // account for inserted newline
		}
	}
	return sqlText
}

func FetchSlowQueries(dsn string) ([]types.GroupedQuery, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			start_time,
			user_host,
			db,
			query_time,
			rows_examined,
			rows_sent,
			lock_time,
			sql_text
		FROM mysql.slow_log
		WHERE sql_text NOT LIKE '%CREATE TABLE%' AND sql_text NOT LIKE '%ALTER TABLE%'
		ORDER BY query_time DESC
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allQueries []types.SlowQuery
	id := 1
	for rows.Next() {
		var q types.SlowQuery
		q.ID = id
		err := rows.Scan(&q.StartTime, &q.UserHost, &q.DB, &q.QueryTime, &q.RowsExamined, &q.RowsSent, &q.LockTime, &q.SQLText)
		if err != nil {
			log.Println(err)
			continue
		}
		q.QueryType = extractQueryType(q.SQLText)
		q.SQLText = FormatSQLForDisplay(q.SQLText) // <--- format for display
		allQueries = append(allQueries, q)
		id++
	}

	// Group queries by normalized SQL
	groups := make(map[string]*types.GroupedQuery)
	for _, q := range allQueries {
		norm := normalizeSQL(q.SQLText)
		fromTable := extractFromTable(norm)
		g, ok := groups[norm]
		if !ok {
			g = &types.GroupedQuery{
				NormalizedSQL: norm,
				QueryType:     q.QueryType,
				FromTable:     fromTable,
			}
			groups[norm] = g
		}
		g.Count++
		g.AvgQueryTime += parseTime(q.QueryTime)
		g.AvgRowsExamined += float64(q.RowsExamined)
		g.AvgRowsSent += float64(q.RowsSent)
		g.Examples = append(g.Examples, q)
	}
	var result []types.GroupedQuery
	for _, g := range groups {
		if g.Count > 0 {
			g.AvgQueryTime /= float64(g.Count)
			g.AvgRowsExamined /= float64(g.Count)
			g.AvgRowsSent /= float64(g.Count)
		}
		result = append(result, *g)
	}
	// Sort by count desc, then avg time desc (do this once here)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].AvgQueryTime > result[j].AvgQueryTime
		}
		return result[i].Count > result[j].Count
	})
	return result, nil
}

// extractQueryType returns the first SQL keyword (uppercased) from the SQL text, ignoring comments and blanks
func extractQueryType(sqlText string) string {
	sqlText = strings.TrimSpace(sqlText)
	if sqlText == "" {
		return ""
	}
	lines := strings.Split(sqlText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		kw := strings.ToUpper(parts[0])
		// List of common SQL query types
		sqlTypes := map[string]struct{}{
			"SELECT": {}, "INSERT": {}, "UPDATE": {}, "DELETE": {}, "ALTER": {}, "CREATE": {}, "DROP": {}, "RENAME": {}, "TRUNCATE": {}, "REPLACE": {}, "CALL": {}, "DO": {}, "HANDLER": {}, "LOAD": {}, "START": {}, "COMMIT": {}, "ROLLBACK": {}, "SAVEPOINT": {}, "RELEASE": {}, "LOCK": {}, "UNLOCK": {}, "SET": {}, "SHOW": {}, "DESCRIBE": {}, "EXPLAIN": {}, "USE": {},
		}
		if _, ok := sqlTypes[kw]; ok {
			return kw
		}
	}
	return "OTHER"
}

// normalizeSQL replaces numbers and quoted strings with ? to group similar queries
func normalizeSQL(sqlText string) string {
	r := strings.NewReplacer(
		"'", " ",
		"\"", " ",
	)
	s := r.Replace(sqlText)
	s = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// parseTime parses MySQL time string (e.g. 00:00:01) to seconds
func parseTime(t string) float64 {
	parts := strings.Split(t, ":")
	if len(parts) != 3 {
		return 0
	}
	h, m, s := parts[0], parts[1], parts[2]
	hh, _ := strconv.Atoi(h)
	mm, _ := strconv.Atoi(m)
	ss, _ := strconv.ParseFloat(s, 64)
	return float64(hh)*3600 + float64(mm)*60 + ss
}

// extractFromTable extracts the first table name after FROM in a normalized SQL string
func extractFromTable(normSQL string) string {
	upper := strings.ToUpper(normSQL)
	fromIdx := strings.Index(upper, " FROM ")
	if fromIdx == -1 {
		return ""
	}
	rest := normSQL[fromIdx+6:]
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
