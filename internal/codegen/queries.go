package codegen

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"slices"
	"strings"
)

func resultRecordName(q core.Query) string {
	return strcase.ToCamel(q.MethodName) + "Row"
}

func createResultRecord(sb *IndentStringBuilder, indentLevel int, q core.Query) {
	recordName := resultRecordName(q)

	sb.WriteIndentedString(indentLevel, "var ret = new "+recordName+"(\n")
	for i, ret := range q.Returns {
		sb.WriteIndentedString(indentLevel+1, ret.ResultStmt(i+1))

		if i < len(q.Returns)-1 {
			sb.WriteString(",\n")
		}
	}
	sb.WriteString("\n")
	sb.WriteIndentedString(indentLevel, ");\n")
}

func completeMethodBody(sb *IndentStringBuilder, q core.Query) {
	sb.WriteString("\n")

	switch q.Command {
	case core.One, core.Many:
		sb.WriteIndentedString(2, "var results = stmt.executeQuery();\n")
	case core.Exec, core.ExecRows, core.ExecResult:
		sb.WriteIndentedString(2, "stmt.execute();\n")
	default:
		sb.WriteIndentedString(2, "// TODO\n")
	}

	switch q.Command {
	case core.One:
		sb.WriteIndentedString(2, "if (!results.next()) {\n")
		sb.WriteIndentedString(3, "return Optional.empty()\n")
		sb.WriteIndentedString(2, "}\n\n")
		createResultRecord(sb, 2, q)
		sb.WriteIndentedString(2, "if (results.next()) {\n")
		sb.WriteIndentedString(3, "throw new SQLException(\"expected one row in result set, but got many\");\n")
		sb.WriteIndentedString(2, "}\n\n")
		sb.WriteIndentedString(2, "return Optional.of(ret);\n")
	case core.Many:
		sb.WriteIndentedString(2, "var retList = new ArrayList<"+resultRecordName(q)+">();\n")
		sb.WriteIndentedString(2, "while (results.next()) {\n")
		createResultRecord(sb, 3, q)
		sb.WriteIndentedString(3, "retList.add(ret);\n")
		sb.WriteIndentedString(2, "}\n\n")
		sb.WriteIndentedString(2, "return retList;\n")
	case core.Exec:
		break
	case core.ExecRows:
		sb.WriteIndentedString(2, "return stmt.getUpdateCount();\n")
	case core.ExecResult:
		sb.WriteIndentedString(2, "var results = stmt.getGeneratedKeys();\n")
		sb.WriteIndentedString(2, "if (!results.next()) {\n")
		sb.WriteIndentedString(3, "throw new SQLException(\"no generated key returned\");\n")
		sb.WriteIndentedString(2, "}\n\n")
		sb.WriteIndentedString(2, "return results.getLong(1);\n")
	default:
		sb.WriteIndentedString(2, "// TODO\n")
	}
}

func BuildQueriesFile(config core.Config, queryFilename string, queries []core.Query) (string, []byte, error) {
	className := strcase.ToCamel(strings.TrimSuffix(queryFilename, ".sql"))
	className = strings.TrimSuffix(className, "Query")
	className = strings.TrimSuffix(className, "Queries")
	className += "Queries"

	imports := make([]string, 0)
	imports = append(imports, "java.sql.Connection", "java.sql.SQLException")

	var nonNullAnnotation string
	if config.NonNullAnnotation != "" {
		imports = append(imports, config.NonNullAnnotation)
		nonNullAnnotation = "@" + config.NonNullAnnotation[strings.LastIndex(config.NonNullAnnotation, ".")+1:] + " "
	}
	var nullableAnnotation string
	if config.NullableAnnotation != "" {
		imports = append(imports, config.NullableAnnotation)
		nullableAnnotation = "@" + config.NullableAnnotation[strings.LastIndex(config.NullableAnnotation, ".")+1:] + " "
	}

	header := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	writeSqlcHeader(header)
	header.WriteString("\n")
	header.WriteString("package " + config.Package + ";\n")
	header.WriteString("\n")

	body := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	body.WriteString("\n")
	// Add the class declaration and constructor
	body.WriteString("public class " + className + " {\n")
	body.WriteIndentedString(1, "private final Connection conn;\n\n")
	body.WriteIndentedString(1, "public "+className+"(Connection conn) {\n")
	body.WriteIndentedString(2, "this.conn = conn;\n")
	body.WriteIndentedString(1, "}\n")

	for _, q := range queries {
		body.WriteString("\n")

		// write the static attribute containing the query string
		body.WriteIndentedString(
			1,
			"private static final String "+q.MethodName+" = \"\"\"-- name: "+q.RawQueryName+" "+q.RawCommand+"\n",
		)
		// for each line in the query, ensure it is indented correctly
		for _, part := range strings.Split(q.Text, "\n") {
			if part == "" {
				continue
			}

			body.WriteIndentedString(2, part+"\n")
		}
		body.WriteIndentedString(2, "\"\"\";\n")

		// write the output record class - TODO figure out if the output is an entire table and if so use a shared model
		var returnType string
		if len(q.Returns) > 0 {
			returnType = resultRecordName(q)

			body.WriteString("\n")
			body.WriteIndentedString(1, "public record "+returnType+"(\n")
			for i, ret := range q.Returns {
				jt := ret.JavaType
				if strings.Contains(jt, ".") {
					parts := strings.Split(jt, ".")

					imports = append(imports, jt)
					jt = parts[len(parts)-1]
				}

				annotation := nonNullAnnotation
				if ret.Nullable {
					annotation = nullableAnnotation
				}

				body.WriteIndentedString(2, annotation+ret.JavaType+" "+ret.Name)
				if i != len(q.Returns)-1 {
					body.WriteString(",\n")
				}
			}
			body.WriteString("\n")
			body.WriteIndentedString(1, ") {}\n")
		}

		// figure out what the return type of the method should be
		switch q.Command {
		case core.One:
			imports = append(imports, "java.util.Optional")
			returnType = "Optional<" + returnType + ">"
		case core.Many:
			imports = append(imports, "java.util.List", "java.util.ArrayList")
			returnType = "List<" + returnType + ">"
		case core.Exec:
			returnType = "void"
		case core.ExecRows:
			returnType = "int"
		case core.ExecResult:
			returnType = "long"
		case core.CopyFrom:
			return "", []byte{}, errors.New("copyFrom is not currently supported")
		}

		methodBody := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
		methodBody.WriteIndentedString(2, "var stmt = conn.prepareStatement("+q.MethodName+");\n")

		// write the method signature
		body.WriteString("\n")
		body.WriteIndentedString(1, fmt.Sprintf("public %s %s(", returnType, q.MethodName))
		if len(q.Args) > 0 {
			body.WriteString("\n")

			for i, arg := range q.Args {
				jt := arg.JavaType
				if strings.Contains(jt, ".") {
					parts := strings.Split(jt, ".")

					imports = append(imports, jt)
					jt = parts[len(parts)-1]
				}

				annotation := nonNullAnnotation
				if arg.Nullable {
					annotation = nullableAnnotation
				}

				body.WriteIndentedString(2, annotation+arg.JavaType+" "+arg.Name)
				if i != len(q.Args)-1 {
					body.WriteString(",\n")
				}

				methodBody.WriteIndentedString(2, arg.BindStmt()+"\n")
			}
			body.WriteString("\n")
			body.WriteIndentedString(1, ") {\n")
		} else {
			body.WriteString(") {\n")
		}

		completeMethodBody(methodBody, q)
		body.WriteString(methodBody.String())
		body.WriteIndentedString(1, "}\n")
	}
	body.WriteString("}\n")

	// sort alphabetically and remove duplicate imports
	slices.Sort(imports)
	imports = slices.Compact(imports)
	for _, imp := range imports {
		header.WriteString("import " + imp + ";\n")
	}

	return className + ".java", []byte(header.String() + body.String()), nil
}
