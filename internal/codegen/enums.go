package codegen

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/poet"
)

var javaInvalidIdentChars = regexp.MustCompile("[^$\\w]")

func EnumClassName(qualifiedName, defaultSchema string) string {
	return strcase.ToCamel(strings.TrimPrefix(qualifiedName, defaultSchema+"."))
}

func enumValueName(value string) string {
	rep := strings.NewReplacer("-", "_", ":", "_", "/", "_", ".", "_")
	name := rep.Replace(value)
	name = strings.ToUpper(name)
	name = javaInvalidIdentChars.ReplaceAllString(name, "")

	r, _ := utf8.DecodeRuneInString(name)
	if unicode.IsDigit(r) {
		name = "_" + name
	}
	return name
}

func BuildEnumFile(engine string, config core.Config, qualName string, enum core.Enum, defaultSchema string) (string, []byte, error) {
	ctx := poet.NewContext(
		config.Package+".models",
		poet.WithIndent(strings.Repeat(config.IndentChar, config.CharsPerIndentLevel)),
	)

	enumName := EnumClassName(qualName, defaultSchema)

	enumBuilder := poet.NewEnumBuilder(enumName).
		WithAnnotation(
			poet.NewAnnotationBuilder(generatedClass).
				WithMember("value", "$S", "io.github.tandemdude.sqlc-gen-java").
				Build(),
		).
		WithModifiers(poet.ModifierPublic)

	if engine == "mysql" {
		enumBuilder.WithValue("BLANK", "")
	}

	// write other values
	for _, value := range enum.Values {
		enumBuilder.WithValue(enumValueName(value), value)
	}

	enumType := poet.NewClassName("", enumName)

	enumBuilder.WithMethods(
		poet.NewMethodBuilder("getValue", poet.String).
			WithCode(
				poet.NewCodeBuilder().
					WithStatement("return this.value").
					Build(),
			).
			Build(),

		poet.NewMethodBuilder("fromValue", enumType).
			// FIXME: Make value final String (?)
			WithParameters(poet.NewMethodParam("value", poet.String)).
			WithCode(
				poet.NewCodeBuilder().
					WithControlFlow("for (var v : $T.values())", func(cb *poet.CodeBuilder) {
						cb.WithRawCode("if (v.value.equals(value)) return v;")
					}, enumType).
					WithStatement(`throw new IllegalArgumentException("No enum constant with value " + value)`).
					Build(),
			).
			Build(),
	)

	fileContents := poet.FormatFile(ctx, enumBuilder.Build(), poet.WithFileComment(core.FileHeaderComment))
	return fmt.Sprintf("enums/%s.java", enumName), []byte(fileContents), nil
}
