package codegen

import (
	"errors"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/poet"
)

var connectionClass = poet.NewClassName("java.sql", "Connection")
var generatedClass = poet.NewClassName("javax.annotation.processing", "Generated")

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
			_, jt, _ = core.ResolveImportAndType(q.Returns[0].JavaType.Type.Name)
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

func BuildQueriesFile(engine string, config core.Config, queryFilename string, queries []core.Query, embeddedModels core.EmbeddedModels, nullableHelpers core.NullableHelpers) (string, []byte, error) {
	ctx := poet.NewContext(
		config.Package,
		poet.WithIndent(strings.Repeat(config.IndentChar, config.CharsPerIndentLevel)),
	)

	var nonNullAnnotation poet.Annotation
	if config.NonNullAnnotation != "" {
		lastIndex := strings.LastIndex(config.NonNullAnnotation, ".")
		pkg := config.NonNullAnnotation[:lastIndex]
		name := config.NonNullAnnotation[lastIndex+1:]

		nonNullAnnotation = poet.NewAnnotationBuilder(poet.NewClassName(pkg, name)).Build()
	}
	var nullableAnnotation poet.Annotation
	if config.NullableAnnotation != "" {
		lastIndex := strings.LastIndex(config.NullableAnnotation, ".")
		pkg := config.NullableAnnotation[:lastIndex]
		name := config.NullableAnnotation[lastIndex+1:]

		nullableAnnotation = poet.NewAnnotationBuilder(poet.NewClassName(pkg, name)).Build()
	}

	className := strcase.ToCamel(strings.TrimSuffix(queryFilename, ".sql"))
	className = strings.TrimSuffix(className, "Query")
	className = strings.TrimSuffix(className, "Queries")
	className += "Queries"

	classBuilder := poet.NewClassBuilder(className).
		WithAnnotation(
			poet.NewAnnotationBuilder(generatedClass).
				WithMember("value", "$S", "io.github.tandemdude.sqlc-gen-java").
				Build(),
		).
		WithModifiers(poet.ModifierPublic).
		WithFields(poet.ClassField{
			Name:      "conn",
			Type:      connectionClass,
			Modifiers: []poet.Modifier{poet.ModifierPrivate, poet.ModifierFinal},
		}).
		WithConstructor(
			poet.NewConstructorBuilder().
				WithParameters(poet.NewMethodParam("conn", connectionClass)).
				WithCode(
					poet.NewCodeBuilder().
						WithStatement("this.conn = conn").
						Build(),
				).
				Build(),
		)

	methods := writeNullableHelpers(ctx, nullableHelpers, nonNullAnnotation, nullableAnnotation)
	var methodBuilder *poet.MethodBuilder
	var method poet.Method

	if config.ExposeConnection {
		method = poet.NewMethodBuilder("getConn", connectionClass).
			WithCode(poet.NewCodeBuilder().
				WithStatement("return this.conn").
				Build(),
			).
			Build()

		methods = append(methods, method)
	}

	for _, q := range queries {
		queryStrBuilder := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
		queryStrBuilder.WriteString("\"\"\"\n")
		queryStrBuilder.WriteIndentedString(1, "-- name: "+q.RawQueryName+" "+q.RawCommand+"\n")
		for _, part := range strings.Split(q.Text, "\n") {
			if part == "" {
				continue
			}

			queryStrBuilder.WriteIndentedString(1, part+"\n")
		}
		queryStrBuilder.WriteIndentedString(1, "\"\"\";")

		classBuilder.WithFields(
			poet.NewClassFieldBuilder(q.MethodName, poet.String).
				WithModifiers(poet.ModifierPublic, poet.ModifierStatic, poet.ModifierFinal).
				WithInitializer(queryStrBuilder.String()).
				Build(),
		)

		// write the output record class
		var returnType poet.TypeName
		if len(q.Returns) > 1 {
			recordName := resultRecordName(q)
			recordBuilder := poet.NewRecordBuilder(recordName)

			// FIXME: Annotations
			// Look at common.go:writeParameter
			// , nonNullAnnotation, nullableAnnotation
			for _, ret := range q.Returns {
				recordBuilder.WithParameters(poet.NewMethodParam(ret.Name, ret.JavaType.Type))
			}

			classBuilder.WithMembers(recordBuilder.Build())

			returnType = poet.NewClassName("", recordName)
		} else if len(q.Returns) == 1 {
			// the query only outputs a single value, we don't need to wrap it in an xxRow record class
			ret := q.Returns[0]

			if ret.JavaType.IsList {
				returnType = poet.ListOf(ret.JavaType.Type)
			} else {
				returnType = ret.JavaType.Type
			}
		}

		// figure out what the return type of the method should be
		switch q.Command {
		case core.One:
			returnType = poet.OptionalOf(returnType)
		case core.Many:
			returnType = poet.ListOf(returnType)
		case core.Exec:
			returnType = poet.Void
		case core.ExecRows:
			returnType = poet.Int
		case core.ExecResult:
			returnType = poet.Long
		case core.CopyFrom:
			return "", []byte{}, errors.New("copyFrom is not currently supported")
		}

		methodBuilder = poet.NewMethodBuilder(q.MethodName, returnType).WithModifiers(poet.ModifierPublic).WithThrows(sqlExceptionClass)
		codeBuilder := poet.NewCodeBuilder()

		if q.Command == core.ExecResult {
			codeBuilder.WithStatement("var stmt = conn.prepareStatement($L, java.sql.Statement.RETURN_GENERATED_KEYS)", q.MethodName)
		} else {
			codeBuilder.WithStatement("var stmt = conn.prepareStatement($L)", q.MethodName)
		}

		for _, arg := range q.Args {
			// FIXME: Annotations
			// Look at common.go:writeParameter
			// , nonNullAnnotation, nullableAnnotation
			methodBuilder.WithParameters(poet.NewMethodParam(arg.Name, arg.JavaType.Type))
			// FIXME: Make BindStmt take in the code builder
			codeBuilder.WithRawCode(arg.BindStmt(engine))
		}

		// FIXME: finish
		//methodBody := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
		//completeMethodBody(methodBody, q, embeddedModels)

		method = methodBuilder.WithCode(codeBuilder.Build()).Build()
		methods = append(methods, method)
	}

	classBuilder.WithMethods(methods...)

	fileContents := poet.FormatFile(ctx, classBuilder.Build(), poet.WithFileComment(core.FileHeaderComment))
	return className + ".java", []byte(fileContents), nil
}
