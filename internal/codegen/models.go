package codegen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
)

func BuildModelFile(config core.Config, name string, model []core.QueryReturn) (string, []byte, error) {
	imports := make([]string, 0)

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
	header.WriteString("package " + config.Package + ".models;\n")
	header.WriteString("\n")
	header.WriteString("import javax.annotation.processing.Generated;\n")
	header.WriteString("\n")

	body := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	body.WriteString("\n")
	body.WriteString("@Generated(\"io.github.tandemdude.sqlc-gen-java\")\n")
	body.WriteString("public record " + strcase.ToCamel(name) + "(\n")
	for i, ret := range model {
		imps, err := body.writeParameter(ret.JavaType, ret.Name, nonNullAnnotation, nullableAnnotation)
		if err != nil {
			return "", nil, err
		}
		if imps != nil {
			imports = append(imports, imps...)
		}

		if i != len(model)-1 {
			body.WriteString(",\n")
		}
	}
	body.WriteString("\n")
	body.WriteString(") {}\n")

	// sort alphabetically and remove duplicate imports
	slices.Sort(imports)
	imports = slices.Compact(imports)
	for _, imp := range imports {
		if imp == "" {
			continue
		}

		header.WriteString("import " + imp + ";\n")
	}

	return fmt.Sprintf("models/%s.java", strcase.ToCamel(name)), []byte(header.String() + body.String()), nil
}
