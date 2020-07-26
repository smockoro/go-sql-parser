package query

type Query struct {
	Type            string
	TableName       string
	Fields          []string
	Condition       []string
	GroupByFields   []string
	HavingCondition []string
	OrderByFields   []string
	InsertValues    [][]string
	UpdateValues    []string
}
