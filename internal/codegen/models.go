package codegen

import (
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
)

func BuildModelsFile(config core.Config, models map[string][]core.QueryReturn) (string, []byte, error) {
	imports := make([]string, 0)

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
	body.WriteString("public class Models {")

	for modelName, modelFields := range models {
		body.WriteString("\n")
		body.WriteIndentedString(1, "public record "+strcase.ToCamel(modelName)+"(\n")
		for i, ret := range modelFields {
			jt := ret.JavaType.Type
			if strings.Contains(jt, ".") {
				parts := strings.Split(jt, ".")

				imports = append(imports, jt)
				jt = parts[len(parts)-1]
			}

			if ret.JavaType.IsList {
				imports = append(imports, "java.util.List", "java.util.Arrays")
				jt = "List<" + jt + ">"
			}

			annotation := nonNullAnnotation
			if ret.JavaType.Nullable {
				annotation = nullableAnnotation
			}

			body.WriteIndentedString(2, annotation+jt+" "+strcase.ToLowerCamel(ret.Name))
			if i != len(modelFields)-1 {
				body.WriteString(",\n")
			}
		}
		body.WriteString("\n")
		body.WriteIndentedString(1, ") {}\n")

	}
	body.WriteString("}")

	// sort alphabetically and remove duplicate imports
	slices.Sort(imports)
	imports = slices.Compact(imports)
	for _, imp := range imports {
		header.WriteString("import " + imp + ";\n")
	}

	return "Models.java", []byte(header.String() + body.String()), nil
}
