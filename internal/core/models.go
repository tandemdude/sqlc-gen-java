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
	SqlType    string
	Type       string
	IsList     bool
	IsNullable bool
}

type QueryArg struct {
	Number   int
	Name     string
	JavaType JavaType
}

// TODO - enum types

var literalBindTypes = []string{"Integer", "Long", "Short", "String", "Boolean", "Float", "Double", "BigDecimal", "byte[]"}
var typeToMethodRename = map[string]string{
	"Integer": "Int",
	"byte[]":  "Bytes",
}
var typeToJavaSqlTypeConst = map[string]string{
	"Integer": "INTEGER",
	"Long":    "BIGINT",
	"Short":   "SMALLINT",
	"Boolean": "BOOLEAN",
	"Float":   "REAL",
	"Double":  "DOUBLE",
}

func (q QueryArg) BindStmt() string {
	typeOnly := q.JavaType.Type[strings.LastIndex(q.JavaType.Type, ".")+1:]

	if q.JavaType.IsList {
		if q.JavaType.IsNullable {
			return fmt.Sprintf("stmt.setArray(%d, %s == null ? null : conn.createArrayOf(\"%s\", %s.toArray()));", q.Number, q.Name, q.JavaType.SqlType, q.Name)
		}
		return fmt.Sprintf("stmt.setArray(%d, conn.createArrayOf(\"%s\", %s.toArray()));", q.Number, q.JavaType.SqlType, q.Name)
	}

	if slices.Contains(literalBindTypes, typeOnly) {
		javaSqlType, ok := typeToJavaSqlTypeConst[typeOnly]
		// annoying special cases
		if found, ok := typeToMethodRename[typeOnly]; ok {
			typeOnly = found
		}
		rawSet := fmt.Sprintf("stmt.set%s(%d, %s);", typeOnly, q.Number, q.Name)

		// if the arg is not nullable, or supports null directly though the method
		if !q.JavaType.IsNullable || !ok {
			return rawSet
		}

		return fmt.Sprintf("%s == null ? stmt.setNull(%d, Types.%s) : %s", q.Name, q.Number, javaSqlType, rawSet)
	}

	return fmt.Sprintf("stmt.setObject(%d, %s);", q.Number, q.Name)
}

type QueryReturn struct {
	Name          string
	JavaType      JavaType
	EmbeddedModel *string
}

func (q QueryReturn) ResultStmt(number int) string {
	typeOnly := q.JavaType.Type[strings.LastIndex(q.JavaType.Type, ".")+1:]

	if q.JavaType.IsList {
		// TODO - check for nullable array support
		return fmt.Sprintf("Arrays.asList((%s[]) results.getArray(%d).getArray())", typeOnly, number)
	}

	if slices.Contains(literalBindTypes, typeOnly) {
		_, ok := typeToJavaSqlTypeConst[typeOnly]
		// annoying special cases
		if found, ok := typeToMethodRename[typeOnly]; ok {
			typeOnly = found
		}

		if q.JavaType.IsNullable && ok {
			return fmt.Sprintf("get%s(results, %d)", typeOnly, number)
		}
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
type EmbeddedModels map[string][]QueryReturn
