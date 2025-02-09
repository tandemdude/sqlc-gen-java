package codegen

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
)

func resultRecordName(q core.Query) string {
	return strcase.ToCamel(q.MethodName) + "Row"
}

func createEmbeddedModel(sb *IndentStringBuilder, prefix, suffix string, identLevel, paramIdx int, r core.QueryReturn, embeddedModels core.EmbeddedModels) int {
	modelName := *r.EmbeddedModel
	model := embeddedModels[modelName]

	sb.WriteIndentedString(identLevel, prefix+modelName+"(\n")
	for i, ret := range model {
		sb.WriteIndentedString(identLevel+1, ret.ResultStmt(paramIdx))

		if i != len(model)-1 {
			sb.WriteString(",\n")
			paramIdx++
		}
	}
	sb.WriteString("\n")
	sb.WriteIndentedString(identLevel, suffix)

	return paramIdx
}

func createResultRecord(sb *IndentStringBuilder, indentLevel int, q core.Query, embeddedModels core.EmbeddedModels) {
	paramIdx := 1

	if len(q.Returns) == 1 {
		// set ret to the item directly instead of wrapping it in the result record
		if q.Returns[0].EmbeddedModel != nil {
			createEmbeddedModel(sb, "var ret = new ", ");\n", indentLevel, paramIdx, q.Returns[0], embeddedModels)
			return
		}

		sb.WriteIndentedString(indentLevel, "var ret = "+q.Returns[0].ResultStmt(1)+";\n")
		return
	}

	recordName := resultRecordName(q)
	sb.WriteIndentedString(indentLevel, "var ret = new "+recordName+"(\n")
	for i, ret := range q.Returns {
		// if this return is an embedded model we need to do a lil bit extra
		if ret.EmbeddedModel != nil {
			paramIdx = createEmbeddedModel(sb, "new ", ")", indentLevel+1, paramIdx, ret, embeddedModels)
		} else {
			sb.WriteIndentedString(indentLevel+1, ret.ResultStmt(paramIdx))
		}

		if i != len(q.Returns)-1 {
			sb.WriteString(",\n")
		}

		paramIdx++
	}
	sb.WriteString("\n")
	sb.WriteIndentedString(indentLevel, ");\n")
}

func completeMethodBody(sb *IndentStringBuilder, q core.Query, embeddedModels core.EmbeddedModels) {
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
		sb.WriteIndentedString(3, "return Optional.empty();\n")
		sb.WriteIndentedString(2, "}\n\n")
		createResultRecord(sb, 2, q, embeddedModels)
		sb.WriteIndentedString(2, "if (results.next()) {\n")
		sb.WriteIndentedString(3, "throw new SQLException(\"expected one row in result set, but got many\");\n")
		sb.WriteIndentedString(2, "}\n\n")
		sb.WriteIndentedString(2, "return Optional.of(ret);\n")
	case core.Many:
		jt := resultRecordName(q)
		if len(q.Returns) == 1 {
			_, jt, _ = core.ResolveImportAndType(q.Returns[0].JavaType.Type)
			if q.Returns[0].EmbeddedModel != nil {
				jt = *q.Returns[0].EmbeddedModel
			}
		}

		sb.WriteIndentedString(2, "var retList = new ArrayList<"+jt+">();\n")
		sb.WriteIndentedString(2, "while (results.next()) {\n")
		createResultRecord(sb, 3, q, embeddedModels)
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

func BuildQueriesFile(config core.Config, queryFilename string, queries []core.Query, embeddedModels core.EmbeddedModels, nullableHelpers core.NullableHelpers) (string, []byte, error) {
	className := strcase.ToCamel(strings.TrimSuffix(queryFilename, ".sql"))
	className = strings.TrimSuffix(className, "Query")
	className = strings.TrimSuffix(className, "Queries")
	className += "Queries"

	imports := make([]string, 0)
	imports = append(imports, "java.sql.SQLException", "java.sql.ResultSet", "java.sql.Types", "java.util.Arrays")

	var nonNullAnnotation string
	if config.NonNullAnnotation != "" {
		imports = append(imports, config.NonNullAnnotation)
		nonNullAnnotation = "@" + config.NonNullAnnotation[strings.LastIndex(config.NonNullAnnotation, ".")+1:]
	}
	var nullableAnnotation string
	if config.NullableAnnotation != "" {
		imports = append(imports, config.NullableAnnotation)
		nullableAnnotation = "@" + config.NullableAnnotation[strings.LastIndex(config.NullableAnnotation, ".")+1:]
	}

	header := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	header.writeSqlcHeader()
	header.WriteString("\n")
	header.WriteString("package " + config.Package + ";\n")
	header.WriteString("\n")

	body := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	body.WriteString("\n")
	// Add the class declaration and constructor
	body.WriteString("public class " + className + " {\n")
	body.WriteIndentedString(1, "private final java.sql.Connection conn;\n\n")
	body.WriteIndentedString(1, "public "+className+"(java.sql.Connection conn) {\n")
	body.WriteIndentedString(2, "this.conn = conn;\n")
	body.WriteIndentedString(1, "}\n")

	if config.ExposeConnection {
		body.WriteString("\n")
		body.WriteIndentedString(1, "public java.sql.Connection getConn() {return this.conn;}\n")
	}

	// boilerplate methods to allow for getting null primitive values
	body.WriteString("\n")
	body.writeNullableHelpers(nullableHelpers, nonNullAnnotation, nullableAnnotation)

	for _, q := range queries {
		body.WriteString("\n")

		// write the static attribute containing the query string
		body.WriteIndentedString(1, "private static final String "+q.MethodName+" = \"\"\"\n")
		body.WriteIndentedString(2, "-- name: "+q.RawQueryName+" "+q.RawCommand+"\n")
		// for each line in the query, ensure it is indented correctly
		for _, part := range strings.Split(q.Text, "\n") {
			if part == "" {
				continue
			}

			body.WriteIndentedString(2, part+"\n")
		}
		body.WriteIndentedString(2, "\"\"\";\n")

		// write the output record class
		var returnType string
		if len(q.Returns) > 1 {
			returnType = resultRecordName(q)

			body.WriteString("\n")
			body.WriteIndentedString(1, "public record "+returnType+"(\n")
			for i, ret := range q.Returns {
				imps, err := body.writeParameter(ret.JavaType, ret.Name, nonNullAnnotation, nullableAnnotation)
				if err != nil {
					return "", nil, err
				}
				if imps != nil {
					imports = append(imports, imps...)
				}

				if i != len(q.Returns)-1 {
					body.WriteString(",\n")
				}
			}
			body.WriteString("\n")
			body.WriteIndentedString(1, ") {}\n")
		} else if len(q.Returns) == 1 {
			// the query only outputs a single value, we don't need to wrap it in an xxRow record class
			ret := q.Returns[0]

			imp, jt, err := core.ResolveImportAndType(ret.JavaType.Type)
			if err != nil {
				return "", nil, err
			}
			imports = append(imports, imp)

			if ret.JavaType.IsList {
				imports = append(imports, "java.util.List")
				jt = "List<" + jt + ">"
			}

			returnType = jt
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
				imps, err := body.writeParameter(arg.JavaType, arg.Name, nonNullAnnotation, nullableAnnotation)
				if err != nil {
					return "", nil, err
				}
				if imps != nil {
					imports = append(imports, imps...)
				}

				if i != len(q.Args)-1 {
					body.WriteString(",\n")
				}

				methodBody.WriteIndentedString(2, arg.BindStmt()+"\n")
			}
			body.WriteString("\n")
			body.WriteIndentedString(1, ") throws SQLException {\n")
		} else {
			body.WriteString(") throws SQLException {\n")
		}

		completeMethodBody(methodBody, q, embeddedModels)
		body.WriteString(methodBody.String())
		body.WriteIndentedString(1, "}\n")
	}
	body.WriteString("}\n")

	// sort alphabetically and remove duplicate imports
	slices.Sort(imports)
	imports = slices.Compact(imports)
	for _, imp := range imports {
		if imp == "" {
			continue
		}

		header.WriteString("import " + imp + ";\n")
	}

	return className + ".java", []byte(header.String() + body.String()), nil
}
