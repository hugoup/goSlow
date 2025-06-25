package types

type SlowQuery struct {
	ID           int
	StartTime    string
	UserHost     string
	DB           string
	QueryTime    string
	RowsExamined int
	SQLText      string
	RowsSent     int
	LockTime     string
	QueryType    string // NEW: type of query (SELECT, UPDATE, etc)
}

type GroupedQuery struct {
	NormalizedSQL   string
	QueryType       string
	FromTable       string // extracted main table name
	Count           int
	AvgQueryTime    float64
	AvgRowsExamined float64
	AvgRowsSent     float64
	Examples        []SlowQuery // all queries in this group
}
