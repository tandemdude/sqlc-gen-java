package codegen

import (
	"fmt"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"os"
	"strings"
)

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

func (b *IndentStringBuilder) writeSqlcHeader() {
	sqlcVersion := os.Getenv("SQLC_VERSION")

	b.WriteString("// Code generated by sqlc. DO NOT EDIT.\n")
	b.WriteString("// versions:\n")
	b.WriteString("//   sqlc " + sqlcVersion + "\n")
	b.WriteString("//   sqlc-gen-java " + core.PluginVersion + "\n")
}

func (b *IndentStringBuilder) writeQueriesBoilerplate(nonNullAnnotation, nullableAnnotation string) {
	methodTypes := [][]string{
		{"Integer", "Int"},
		{"Long", "Long"},
		{"Float", "Float"},
		{"Double", "Double"},
		{"Boolean", "Boolean"},
	}

	for _, methodType := range methodTypes {
		b.WriteIndentedString(1, fmt.Sprintf(
			"private static %s get%s(%s rs, int col) throws SQLException {\n",
			core.Annotate(methodType[0], nullableAnnotation),
			methodType[1],
			core.Annotate("ResultSet", nonNullAnnotation),
		))
		b.WriteIndentedString(2, fmt.Sprintf(
			"var colVal = rs.get%s(col); return rs.wasNull() ? null : colVal;\n",
			methodType[1],
		))
		b.WriteIndentedString(1, "}\n")
	}

	b.WriteIndentedString(1, fmt.Sprintf(
		"private static <T> %s getList(%s rs, int col, Class<T[]> as) throws SQLException {\n",
		core.Annotate("List<T>", nullableAnnotation),
		core.Annotate("ResultSet", nonNullAnnotation),
	))
	b.WriteIndentedString(2, "var colVal = rs.getArray(col); return rs.wasNull() ? null : Arrays.asList(as.cast(colVal.getArray()));\n")
	b.WriteIndentedString(1, "}\n")
}
