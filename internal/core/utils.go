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
