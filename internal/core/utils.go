package core

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ResolveImportAndType extracts the import required, and type representation of the given java type.
func ResolveImportAndType(typ string) (string, string, error) {
	if !strings.Contains(typ, ".") {
		return "", typ, nil
	}

	parts := strings.Split(typ, ".")
	capitalIdx := slices.IndexFunc(parts, func(s string) bool {
		r, _ := utf8.DecodeRuneInString(s)
		return unicode.IsUpper(r)
	})

	if capitalIdx == -1 {
		// fatal error, this should never happen
		return "", "", fmt.Errorf("failed resolving type and import for %s", typ)
	}

	if capitalIdx == 0 {
		// special case - nested class in same package, no import required
		return "", strings.Join(parts, "."), nil
	}
	// build the import and the type name from the resolved outer class name
	return strings.Join(parts[:capitalIdx+1], "."), strings.Join(parts[capitalIdx:], "."), nil
}

// Annotate adds the given annotation to the given Java type. The annotation will always be added to the final
// element of the type. For types with
// E.g.
// - for a package-qualified type: "org.example.@Annotation Foo"
// - for an array type: "Foo @Annotation []"
// - for a nested type: "Foo.@Annotation Bar"
func Annotate(typ, annotation string) string {
	annotation = strings.TrimSpace(annotation)
	if strings.HasSuffix(typ, "[]") {
		return fmt.Sprintf("%s %s []", strings.TrimSuffix(typ, "[]"), annotation)
	}

	if !strings.Contains(typ, ".") {
		return fmt.Sprintf("%s %s", annotation, typ)
	}

	// TODO - take into account type parameters?

	parts := strings.Split(typ, ".")
	return fmt.Sprintf("%s.%s %s", strings.Join(parts[:len(parts)-1], "."), annotation, parts[len(parts)-1])
}
