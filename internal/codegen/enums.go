package codegen

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
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

func BuildEnumFile(engine string, conf core.Config, qualName string, enum core.Enum, defaultSchema string) (string, []byte, error) {
	className := EnumClassName(qualName, defaultSchema)

	sb := IndentStringBuilder{indentChar: conf.IndentChar, charsPerIndentLevel: conf.CharsPerIndentLevel}
	sb.writeSqlcHeader()
	sb.WriteString("\n")
	sb.WriteString("package " + conf.Package + ".enums;\n")
	sb.WriteString("\n")
	sb.WriteString("import javax.annotation.processing.Generated;\n")
	sb.WriteString("\n")
	sb.WriteString("@Generated(\"io.github.tandemdude.sqlc-gen-java\")\n")
	sb.WriteString("public enum " + className + " {\n")

	if engine == "mysql" {
		sb.WriteIndentedString(1, "BLANK(\"\"),\n")
	}

	// write other values
	for i, value := range enum.Values {
		name := enumValueName(value)
		sb.WriteIndentedString(1, fmt.Sprintf("%s(\"%s\")", name, value))

		if i < len(enum.Values)-1 {
			sb.WriteString(",\n")
		}
	}
	sb.WriteString(";\n\n")
	sb.WriteIndentedString(1, "private final String value;\n\n")
	sb.WriteIndentedString(1, className+"(final String value) {\n")
	sb.WriteIndentedString(2, "this.value = value;\n")
	sb.WriteIndentedString(1, "}\n\n")
	sb.WriteIndentedString(1, "public String getValue() {\n")
	sb.WriteIndentedString(2, "return this.value;")
	sb.WriteIndentedString(1, "}\n\n")
	sb.WriteIndentedString(1, "public static fromValue(final String value) {\n")
	sb.WriteIndentedString(2, "for (var v : "+className+".values()) {\n")
	sb.WriteIndentedString(3, "if (v.value.equals(value)) return v;\n")
	sb.WriteIndentedString(2, "}\n")
	sb.WriteIndentedString(2, "throw new IllegalArgumentException(\"No enum constant with value \" + value);\n")
	sb.WriteIndentedString(1, "}\n")
	sb.WriteString("}\n")

	return fmt.Sprintf("enums/%s.java", className), []byte(sb.String()), nil
}
