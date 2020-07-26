package sqlparser_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	sqlparser "github.com/smockoro/go-sql-parser"
	"github.com/smockoro/go-sql-parser/query"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Name     string
	SQL      string
	Expected *query.Query
	Err      error
}

func TestNoType(t *testing.T) {
	tcs := []testCase{
		{
			Name:     "No Type error",
			SQL:      "SEGECT a, b FROM table",
			Expected: nil,
			Err:      fmt.Errorf("syntax error\n"),
		},
	}

	t.Parallel()
	for _, tc := range tcs {
		p := &sqlparser.Parser{}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			query, err := p.Parse(tc.SQL)
			if tc.Err == nil && !cmp.Equal(query, tc.Expected) {
				t.Errorf("expected %v, actual %v", tc.Expected, query)
			}
			if tc.Err != err {
				require.Equal(t, tc.Err, err, "syntax error\n")
			}
		})
	}

}

func TestSELECT(t *testing.T) {
	tcs := []testCase{
		{
			Name: "SELECT a, b FROM table",
			SQL:  "SELECT a, b FROM table",
			Expected: &query.Query{
				Type:      "SELECT",
				TableName: "table",
				Fields:    []string{"a", "b"},
			},
			Err: nil,
		},
		{
			Name: "SELECT * FROM table",
			SQL:  "SELECT * FROM table",
			Expected: &query.Query{
				Type:      "SELECT",
				TableName: "table",
				Fields:    []string{"*"},
			},
			Err: nil,
		},
		{
			Name: "SELECT * table",
			SQL:  "SELECT * table",
			Err:  fmt.Errorf("syntax error\n"),
		},
		{
			Name: "SELECT a b FROM table",
			SQL:  "SELECT a b FROM table",
			Err:  fmt.Errorf("syntax error\n"),
		},
		{
			Name: "SELECT * FROM table WHERE a >= 1",
			SQL:  "SELECT * FROM table WHERE a >= 1",
			Expected: &query.Query{
				Type:      "SELECT",
				TableName: "table",
				Fields:    []string{"*"},
				Condition: []string{"a", ">=", "1"},
			},
			Err: nil,
		},
		{
			Name: "SELECT * FROM table WHERE a >= 1 ORDER BY b, c",
			SQL:  "SELECT * FROM table WHERE a >= 1 ORDER BY b, c",
			Expected: &query.Query{
				Type:          "SELECT",
				TableName:     "table",
				Fields:        []string{"*"},
				Condition:     []string{"a", ">=", "1"},
				OrderByFields: []string{"b", "c"},
			},
			Err: nil,
		},
		{
			Name: "SELECT * FROM table WHERE a >= 1 GROUP BY d, e ORDER BY b, c",
			SQL:  "SELECT * FROM table WHERE a >= 1 GROUP BY d, e ORDER BY b, c",
			Expected: &query.Query{
				Type:          "SELECT",
				TableName:     "table",
				Fields:        []string{"*"},
				Condition:     []string{"a", ">=", "1"},
				OrderByFields: []string{"b", "c"},
				GroupByFields: []string{"d", "e"},
			},
			Err: nil,
		},
		{
			Name: "SELECT * FROM table WHERE a >= 1 GROUP BY d",
			SQL:  "SELECT * FROM table WHERE a >= 1 GROUP BY d",
			Expected: &query.Query{
				Type:          "SELECT",
				TableName:     "table",
				Fields:        []string{"*"},
				Condition:     []string{"a", ">=", "1"},
				GroupByFields: []string{"d"},
			},
			Err: nil,
		},
	}

	t.Parallel()
	for _, tc := range tcs {
		p := &sqlparser.Parser{}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			query, err := p.Parse(tc.SQL)
			if tc.Err == nil && !cmp.Equal(query, tc.Expected) {
				t.Errorf("expected %v, actual %v", tc.Expected, query)
			}
			if tc.Err != err {
				require.Equal(t, tc.Err, err, "syntax error\n")
			}
		})
	}
}

func TestDELETE(t *testing.T) {
	tcs := []testCase{
		{
			Name: "DELETE FROM table",
			SQL:  "DELETE FROM table",
			Expected: &query.Query{
				Type:      "DELETE",
				TableName: "table",
			},
			Err: nil,
		},
		{
			Name: "DELETE FROM table WHERE a >= 1",
			SQL:  "DELETE FROM table WHERE a >= 1",
			Expected: &query.Query{
				Type:      "DELETE",
				TableName: "table",
				Condition: []string{"a", ">=", "1"},
			},
			Err: nil,
		},
		{
			Name: "DELETE FROM table WHERE a >= 1 AND b + c <= 100",
			SQL:  "DELETE FROM table WHERE a >= 1 AND b + c <= 100",
			Expected: &query.Query{
				Type:      "DELETE",
				TableName: "table",
				Condition: []string{"a", ">=", "1", "AND",
					"b", "+", "c", "<=", "100"},
			},
			Err: nil,
		},
		{
			Name: "FROM not exists",
			SQL:  "DELETE table",
			Err:  fmt.Errorf("syntax error\n"),
		},
		{
			Name: "WHERE not exists",
			SQL:  "DELETE FROM table a >= 1",
			Err:  fmt.Errorf("syntax error\n"),
		},
	}

	t.Parallel()
	for _, tc := range tcs {
		p := &sqlparser.Parser{}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			query, err := p.Parse(tc.SQL)
			if tc.Err == nil && !cmp.Equal(query, tc.Expected) {
				t.Errorf("expected %v, actual %v", tc.Expected, p.Query)
			}
			if tc.Err != err {
				require.Equal(t, tc.Err, err, "syntax error\n")
			}
		})
	}
}

func TestInsert(t *testing.T) {
	tcs := []testCase{
		{
			Name: "INSERT INTO table VALUES (1, 2, 3)",
			SQL:  "INSERT INTO table VALUES (1, 2, 3)",
			Expected: &query.Query{
				Type:      "INSERT",
				TableName: "table",
				InsertValues: [][]string{
					[]string{"1", "2", "3"},
				},
			},
			Err: nil,
		},
		{
			Name: "INSERT INTO table VALUES (1, 2, 3),(4,5,6)",
			SQL:  "INSERT INTO table VALUES (1, 2, 3),(4,5,6)",
			Expected: &query.Query{
				Type:      "INSERT",
				TableName: "table",
				InsertValues: [][]string{
					[]string{"1", "2", "3"},
					[]string{"4", "5", "6"},
				},
			},
			Err: nil,
		},
		{
			Name: "INSERT INTO table (a, b, c) VALUES (1, 2, 3),(4,5,6)",
			SQL:  "INSERT INTO table (a, b, c) VALUES (1, 2, 3),(4,5,6)",
			Expected: &query.Query{
				Type:      "INSERT",
				TableName: "table",
				Fields:    []string{"a", "b", "c"},
				InsertValues: [][]string{
					[]string{"1", "2", "3"},
					[]string{"4", "5", "6"},
				},
			},
			Err: nil,
		},
	}

	t.Parallel()
	for _, tc := range tcs {
		p := &sqlparser.Parser{}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			query, err := p.Parse(tc.SQL)
			if tc.Err == nil && !cmp.Equal(query, tc.Expected) {
				t.Errorf("expected %v, actual %v", tc.Expected, p.Query)
			}
			if tc.Err != err {
				require.Equal(t, tc.Err, err, "syntax error\n")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tcs := []testCase{
		{
			Name: "UPDATE table SET a = 1",
			SQL:  "UPDATE table SET a = 1",
			Expected: &query.Query{
				Type:         "UPDATE",
				TableName:    "table",
				Fields:       []string{"a"},
				UpdateValues: []string{"1"},
			},
			Err: nil,
		},
		{
			Name: "UPDATE table SET a = 1, b = 2",
			SQL:  "UPDATE table SET a = 1, b = 2",
			Expected: &query.Query{
				Type:         "UPDATE",
				TableName:    "table",
				Fields:       []string{"a", "b"},
				UpdateValues: []string{"1", "2"},
			},
			Err: nil,
		},
	}

	t.Parallel()
	for _, tc := range tcs {
		p := &sqlparser.Parser{}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			query, err := p.Parse(tc.SQL)
			if tc.Err == nil && !cmp.Equal(query, tc.Expected) {
				t.Errorf("expected %v, actual %v", tc.Expected, p.Query)
			}
			if tc.Err != err {
				require.Equal(t, tc.Err, err, "syntax error\n")
			}
		})
	}
}
