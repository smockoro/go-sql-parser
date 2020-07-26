package sqlparser

import (
	"fmt"
	"strings"

	"github.com/smockoro/go-sql-parser/query"
)

type Parser struct {
	SQL   string
	i     int
	Query *query.Query
	step  step
}

func (p *Parser) peek() string {
	peeked, _ := p.peekWithLength()
	return peeked
}

func (p *Parser) pop() {
	//fmt.Printf("pop, now p.i: %d\n", p.i)
	_, l := p.peekWithLength()
	p.i += l
	p.popWhitespace()
}

var reservedWords = []string{
	"(", ")", ">=", "<=", "!=", ",", "=", ">", "<",
	"SELECT", "INSERT", "INTO", "VALUES", "UPDATE",
	"DELETE", "WHERE", "FROM", "SET", "GROUP BY",
	"ORDER BY", "HAVING",
}

func (p *Parser) popWhitespace() {
	for ; p.i < len(p.SQL) && p.SQL[p.i] == ' '; p.i++ {
	}
}

func (p *Parser) peekWithLength() (string, int) {
	if p.i >= len(p.SQL) {
		return "", 0
	}
	for _, rWord := range reservedWords {
		token := p.SQL[p.i:min(len(p.SQL), p.i+len(rWord))]
		upToken := strings.ToUpper(token)
		if upToken == rWord {
			return upToken, len(upToken)
		}
	}
	return p.peekIdentifierWithLength()
}

func (p *Parser) peekIdentifierWithLength() (string, int) {
	for j := 1; p.i+j < len(p.SQL); j++ {
		if p.SQL[p.i+j] == ' ' || p.SQL[p.i+j] == ',' || p.SQL[p.i+j] == ')' {
			return p.SQL[p.i : p.i+j], j
		}
	}
	return p.SQL[p.i:], len(p.SQL[p.i:])
}

func min(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

type step int

const (
	stepType step = iota
	stepSelectField
	stepSelectFrom
	stepSelectComma
	stepSelectFromTable
	stepInsertTable
	stepInsertFieldsOpeningParens
	stepInsertFields
	stepInsertFieldsCommaOrClosingParens
	stepInsertValuesOpeningParens
	stepInsertValuesRWord
	stepInsertValues
	stepInsertValuesCommaOrClosingParens
	stepInsertValuesCommaBeforeOpeningParens
	stepUpdateTable
	stepUpdateSet
	stepUpdateField
	stepUpdateEquals
	stepUpdateValue
	stepUpdateComma
	stepDeleteFromTable
	stepWhere
	stepWhereField
	stepWhereOperator
	stepWhereValue
	stepOrderBy
	stepGroupBy
	stepOrderByField
	stepGroupByField
	stepOrderByComma
	stepGroupByComma
)

func (p *Parser) Parse(sql string) (*query.Query, error) {
	p.SQL = sql
	p.Query = &query.Query{}
	p.step = stepType
	for p.i < len(p.SQL) {
		//fmt.Printf("*********Loop %d : Step %v *******\n", count, p.step)
		switch p.step {
		case stepType:
			switch strings.ToUpper(p.peek()) {
			case "SELECT":
				p.Query.Type = "SELECT"
				p.step = stepSelectField
				p.pop()
			case "UPDATE":
				p.Query.Type = "UPDATE"
				p.step = stepUpdateTable
				p.pop()
			case "INSERT":
				p.Query.Type = "INSERT"
				p.pop()
				switch p.peek() {
				case "INTO":
					p.pop()
					p.step = stepInsertTable
				default:
					return nil, fmt.Errorf("syntax error\n")
				}
			case "DELETE":
				p.Query.Type = "DELETE"
				p.pop()
				if from := p.peek(); from != "FROM" {
					return nil, fmt.Errorf("syntax error\n")
				}
				p.pop()
				p.step = stepDeleteFromTable
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepSelectField:
			filedName := p.peek()
			p.Query.Fields = append(p.Query.Fields, filedName)
			p.step = stepSelectComma
			p.pop()
		case stepSelectComma:
			comma := p.peek()
			if comma != "," {
				p.step = stepSelectFrom
			} else {
				p.step = stepSelectField
				p.pop()
			}
		case stepSelectFrom:
			if from := p.peek(); from != "FROM" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.step = stepSelectFromTable
			p.pop()
		case stepSelectFromTable:
			tableName := p.peek()
			p.Query.TableName = tableName
			p.step = stepWhere
			p.pop()
		case stepDeleteFromTable:
			p.Query.TableName = p.peek()
			p.pop()
			p.step = stepWhere
		case stepWhere:
			if where := p.peek(); where != "WHERE" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.step = stepWhereField
			p.pop()
		case stepWhereField:
			opeField := p.peek()
			p.Query.Condition = append(p.Query.Condition, opeField)
			p.step = stepWhereOperator
			p.pop()
		case stepWhereOperator:
			operator := p.peek()
			p.Query.Condition = append(p.Query.Condition, operator)
			p.step = stepWhereValue
			p.pop()
		case stepWhereValue:
			opeValue := p.peek()
			p.Query.Condition = append(p.Query.Condition, opeValue)
			p.pop()
			switch p.peek() {
			case "ORDER BY":
				p.step = stepOrderBy
			case "GROUP BY":
				p.step = stepGroupBy
			default:
				p.step = stepWhereOperator
			}
		case stepOrderBy:
			p.pop()
			p.step = stepOrderByField
		case stepGroupBy:
			p.pop()
			p.step = stepGroupByField
		case stepOrderByField:
			field := p.peek()
			p.Query.OrderByFields = append(p.Query.OrderByFields, field)
			p.pop()
			switch p.peek() {
			case ",":
				p.step = stepOrderByComma
			default:
				p.step = -1
			}
		case stepGroupByField:
			field := p.peek()
			p.Query.GroupByFields = append(p.Query.GroupByFields, field)
			p.pop()
			switch p.peek() {
			case "ORDER BY":
				p.step = stepOrderBy
			case ",":
				p.step = stepGroupByComma
			default:
				p.step = -1
			}
		case stepOrderByComma:
			p.pop()
			p.step = stepOrderByField
		case stepGroupByComma:
			p.pop()
			p.step = stepGroupByField
		case stepInsertTable:
			p.Query.TableName = p.peek()
			p.pop()
			p.step = stepInsertFieldsOpeningParens
		case stepInsertFieldsOpeningParens:
			openingParens := p.peek()
			switch openingParens {
			case "(":
				p.pop()
				p.step = stepInsertFields
			case "VALUES":
				p.step = stepInsertValuesRWord
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepInsertFields:
			field := p.peek()
			p.Query.Fields = append(p.Query.Fields, field)
			p.pop()
			p.step = stepInsertFieldsCommaOrClosingParens
		case stepInsertFieldsCommaOrClosingParens:
			switch p.peek() {
			case ",":
				p.pop()
				p.step = stepInsertFields
			case ")":
				p.pop()
				p.step = stepInsertValuesRWord
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepInsertValuesRWord:
			if p.peek() != "VALUES" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.pop()
			p.step = stepInsertValuesOpeningParens
		case stepInsertValuesOpeningParens:
			openingParens := p.peek()
			switch openingParens {
			case "(":
				p.Query.InsertValues = append(p.Query.InsertValues, []string{})
				p.pop()
				p.step = stepInsertValues
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepInsertValues:
			value := p.peek()
			p.Query.InsertValues[len(p.Query.InsertValues)-1] = append(p.Query.InsertValues[len(p.Query.InsertValues)-1], value)
			p.pop()
			p.step = stepInsertValuesCommaOrClosingParens
		case stepInsertValuesCommaOrClosingParens:
			switch p.peek() {
			case ",":
				p.pop()
				p.step = stepInsertValues
			case ")":
				p.pop()
				p.step = stepInsertValuesCommaBeforeOpeningParens
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepInsertValuesCommaBeforeOpeningParens:
			switch p.peek() {
			case ",":
				p.pop()
				p.step = stepInsertValuesOpeningParens
			default:
				return nil, fmt.Errorf("syntax error\n")
			}
		case stepUpdateTable:
			tableName := p.peek()
			if tableName == "SET" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.Query.TableName = tableName
			p.pop()
			p.step = stepUpdateSet
		case stepUpdateSet:
			set := p.peek()
			if set != "SET" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.pop()
			p.step = stepUpdateField
		case stepUpdateField:
			p.Query.Fields = append(p.Query.Fields, p.peek())
			p.pop()
			p.step = stepUpdateEquals
		case stepUpdateEquals:
			if p.peek() != "=" {
				return nil, fmt.Errorf("syntax error\n")
			}
			p.pop()
			p.step = stepUpdateValue
		case stepUpdateValue:
			p.Query.UpdateValues = append(p.Query.UpdateValues, p.peek())
			p.pop()
			p.step = stepUpdateComma
		case stepUpdateComma:
			switch p.peek() {
			case ",":
				p.pop()
				p.step = stepUpdateField
			case "WHERE":
				p.step = stepWhere
			}
		}
	}
	return p.Query, nil
}
