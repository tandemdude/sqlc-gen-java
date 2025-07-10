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
	IsEnum     bool
}

type QueryArg struct {
	Number   int
	Name     string
	JavaType JavaType
}

// TODO - enum types

var (
	literalBindTypes   = []string{"Integer", "Long", "Short", "String", "Boolean", "Float", "Double", "BigDecimal", "byte[]"}
	typeToMethodRename = map[string]string{
		"Integer": "Int",
		"byte[]":  "Bytes",
	}
)

var typeToJavaSqlTypeConst = map[string]string{
	"Integer": "INTEGER",
	"Long":    "BIGINT",
	"Short":   "SMALLINT",
	"Boolean": "BOOLEAN",
	"Float":   "REAL",
	"Double":  "DOUBLE",
}

func (q QueryArg) BindStmt(engine string) string {
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

		return fmt.Sprintf(`
		if (%s != null) {
		    %s
		} else {
		    stmt.setNull(%d, java.sql.Types.%s) 
		}
		`, q.Name, rawSet, q.Number, javaSqlType)
	}

	if q.JavaType.IsEnum {
		// postgres doesn't like it if you setString an enum directly unfortunately
		if engine == "postgresql" {
			if q.JavaType.IsNullable {
				return fmt.Sprintf("stmt.setObject(%d, %s == null ? null : %s.getValue(), java.sql.Types.OTHER);", q.Number, q.Name, q.Name)
			}
			return fmt.Sprintf("stmt.setObject(%d, %s.getValue(), java.sql.Types.OTHER);", q.Number, q.Name)
		}

		if q.JavaType.IsNullable {
			return fmt.Sprintf("stmt.setString(%d, %s == null ? null : %s.getValue());", q.Number, q.Name, q.Name)
		}
		return fmt.Sprintf("stmt.setString(%d, %s.getValue());", q.Number, q.Name)
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
		if q.JavaType.IsNullable {
			return fmt.Sprintf("getList(results, %d, %s[].class)", number, typeOnly)
		}
		return fmt.Sprintf("Arrays.asList(%s[].class.cast(results.getArray(%d).getArray()))", typeOnly, number)
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

	if q.JavaType.IsEnum {
		if q.JavaType.IsNullable {
			return fmt.Sprintf("Optional.ofNullable(results.getString(%d)).map(%s::fromValue).orElse(null)", number, typeOnly)
		}
		return fmt.Sprintf("%s.fromValue(results.getString(%d))", typeOnly, number)
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

type NullableHelpers struct {
	Int     bool
	Long    bool
	Float   bool
	Double  bool
	Boolean bool
	List    bool
}

type Enum struct {
	Schema string
	Name   string
	Values []string
}

type (
	Queries        map[string][]Query
	EmbeddedModels map[string][]QueryReturn
	// Enums is a map of "schema_name.enum_name" to enum value.
	Enums map[string]Enum
)
