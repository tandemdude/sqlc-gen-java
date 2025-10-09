package codegen

import (
	"fmt"
	"strings"

	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/poet"
)

var resultSetClass = poet.NewClassName("java.sql", "ResultSet")
var sqlExceptionClass = poet.NewClassName("java.sql", "SQLException")

type IndentStringBuilder struct {
	strings.Builder

	indentChar          string
	charsPerIndentLevel int
}

func NewIndentStringBuilder(indentChar string, charsPerIndentLevel int) *IndentStringBuilder {
	return &IndentStringBuilder{
		indentChar:          indentChar,
		charsPerIndentLevel: charsPerIndentLevel,
	}
}

func (b *IndentStringBuilder) WriteIndentedString(level int, s string) int {
	count, _ := b.WriteString(strings.Repeat(b.indentChar, level*b.charsPerIndentLevel) + s)
	return count
}

type nullableHelper struct {
	ShouldOutput bool
	ReturnType   poet.TypeName
	ArgType      string
}

func writeNullableHelpers(ctx *poet.Context, nullableHelpers core.NullableHelpers, nonNullAnnotation, nullableAnnotation poet.Annotation) []poet.Method {
	// FIXME: Annotations
	var methods []poet.Method

	methodTypes := []nullableHelper{
		{nullableHelpers.Int, poet.IntBoxed, "Int"},
		{nullableHelpers.Long, poet.LongBoxed, "Long"},
		{nullableHelpers.Float, poet.FloatBoxed, "Float"},
		{nullableHelpers.Double, poet.DoubleBoxed, "Double"},
		{nullableHelpers.Boolean, poet.BoolBoxed, "Boolean"},
	}

	for _, methodType := range methodTypes {
		if !methodType.ShouldOutput {
			continue
		}

		method := poet.NewMethodBuilder(fmt.Sprintf("get%s", methodType.ArgType), methodType.ReturnType).
			WithParameters(poet.NewMethodParam("rs", resultSetClass), poet.NewMethodParam("col", poet.Int)).
			WithThrows(sqlExceptionClass).
			WithCode(
				poet.NewCodeBuilder().
					WithStatement("var colVar = rs.get$L(col)", methodType.ArgType).
					WithStatement("return rs.wasNull() ? null : colVal").
					Build(),
			).
			Build()

		methods = append(methods, method)
	}

	if nullableHelpers.List {
		ctx.Import("java.util.Arrays")
		genericParamT := poet.NewGenericParam("T")

		method := poet.NewMethodBuilder("getList", poet.ListOf(genericParamT)).
			WithGenericParameters(genericParamT).
			WithParameters(poet.NewMethodParam("rs", resultSetClass), poet.NewMethodParam("col", poet.Int)).
			WithThrows(sqlExceptionClass).
			WithCode(
				poet.NewCodeBuilder().
					WithStatement("var colVal = rs.getArray(col)").
					WithStatement("return colVal == null ? null : Arrays.asList(as.cast(colVal.getArray()))").
					Build(),
			).
			Build()

		methods = append(methods, method)
	}

	return methods
}

//func (b *IndentStringBuilder) writeParameter(javaType core.JavaType, name, nonNullAnnotation, nullableAnnotation string) ([]string, error) {
//	imp, jt, err := core.ResolveImportAndType(javaType.Type)
//	if err != nil {
//		return nil, err
//	}
//	imports := []string{imp}
//
//	if javaType.IsList {
//		imports = append(imports, "java.util.List")
//		jt = "List<" + jt + ">"
//	}
//
//	annotation := nonNullAnnotation
//	if javaType.IsNullable {
//		annotation = nullableAnnotation
//	}
//
//	newType, unboxed := core.MaybeUnbox(javaType.Type, javaType.IsNullable)
//	if !unboxed {
//		newType = core.Annotate(jt, annotation)
//	}
//
//	b.WriteIndentedString(2, newType+" "+name)
//	return imports, nil
//}
