package core

import (
	"fmt"
	"slices"
	"strings"
)

type QueryCommand int

const (
	One QueryCommand = iota
	Many
	Exec
	ExecRows
	ExecResult
	CopyFrom
)

func QueryCommandFor(rawCommand string) (QueryCommand, error) {
	switch rawCommand {
	case ":one":
		return One, nil
	case ":many":
		return Many, nil
	case ":exec":
		return Exec, nil
	case ":execrows":
		return ExecRows, nil
	case ":execresult":
		return ExecResult, nil
	case ":copyfrom":
		return CopyFrom, nil
	default:
		return One, fmt.Errorf(`unknown query command "%s"`, rawCommand)
	}
}

type JavaType struct {
	SqlType  string
	Type     string
	IsList   bool
	Nullable bool
}

type QueryArg struct {
	Number   int
	Name     string
	JavaType JavaType
}

// TODO - enum types
var literalBindTypes = []string{"Long", "Short", "String", "Boolean", "Float", "Double", "BigDecimal"}

func (q QueryArg) BindStmt() string {
	typeOnly := q.JavaType.Type[strings.LastIndex(q.JavaType.Type, ".")+1:]

	if q.JavaType.IsList {
		if q.JavaType.Nullable {
			return fmt.Sprintf("stmt.setArray(%d, %s == null ? null : conn.createArrayOf(\"%s\", %s.toArray()));", q.Number, q.Name, q.JavaType.SqlType, q.Name)
		}
		return fmt.Sprintf("stmt.setArray(%d, conn.createArrayOf(\"%s\", %s.toArray()));", q.Number, q.JavaType.SqlType, q.Name)
	}

	if slices.Contains(literalBindTypes, typeOnly) {
		return fmt.Sprintf("stmt.set%s(%d, %s);", typeOnly, q.Number, q.Name)
	}

	return fmt.Sprintf("stmt.setObject(%d, %s);", q.Number, q.Name)
}

type QueryReturn struct {
	Name     string
	JavaType JavaType
}

func (q QueryReturn) ResultStmt(number int) string {
	typeOnly := q.JavaType.Type[strings.LastIndex(q.JavaType.Type, ".")+1:]

	if q.JavaType.IsList {
		return fmt.Sprintf("Arrays.asList((%s[]) results.getArray(%d).getArray())", typeOnly, number)
	}

	if slices.Contains(literalBindTypes, typeOnly) {
		return fmt.Sprintf("results.get%s(%d)", typeOnly, number)
	}

	return fmt.Sprintf("results.getObject(%d, %s.class)", number, typeOnly)
}

type Query struct {
	RawCommand   string
	Command      QueryCommand
	Text         string
	RawQueryName string
	MethodName   string
	Args         []QueryArg
	Returns      []QueryReturn
}

type Queries map[string][]Query
