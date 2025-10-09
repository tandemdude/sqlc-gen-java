package codegen

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/poet"
)

func BuildModelFile(config core.Config, name string, model []core.QueryReturn) (string, []byte, error) {
	ctx := poet.NewContext(
		config.Package+".models",
		poet.WithIndent(strings.Repeat(config.IndentChar, config.CharsPerIndentLevel)),
	)

	//var nonNullAnnotation poet.Annotation
	//if config.NonNullAnnotation != "" {
	//	lastIndex := strings.LastIndex(config.NonNullAnnotation, ".")
	//	pkg := config.NonNullAnnotation[:lastIndex]
	//	name := config.NonNullAnnotation[lastIndex+1:]
	//
	//	nonNullAnnotation = poet.NewAnnotationBuilder(poet.NewClassName(pkg, name)).Build()
	//}
	//var nullableAnnotation poet.Annotation
	//if config.NullableAnnotation != "" {
	//	lastIndex := strings.LastIndex(config.NullableAnnotation, ".")
	//	pkg := config.NullableAnnotation[:lastIndex]
	//	name := config.NullableAnnotation[lastIndex+1:]
	//
	//	nullableAnnotation = poet.NewAnnotationBuilder(poet.NewClassName(pkg, name)).Build()
	//}

	recordName := strcase.ToCamel(name)

	recordBuilder := poet.NewRecordBuilder(recordName).
		WithAnnotation(
			poet.NewAnnotationBuilder(generatedClass).
				WithMember("value", "$S", "io.github.tandemdude.sqlc-gen-java").
				Build(),
		).
		WithModifiers(poet.ModifierPublic)

	for _, ret := range model {
		// FIXME: Annotations
		// Look at common.go:writeParameter
		// , nonNullAnnotation, nullableAnnotation
		recordBuilder.WithParameters(poet.NewMethodParam(ret.Name, ret.JavaType.Type))
	}

	fileContents := poet.FormatFile(ctx, recordBuilder.Build(), poet.WithFileComment(core.FileHeaderComment))
	return fmt.Sprintf("models/%s.java", recordName), []byte(fileContents), nil
}
