package poet

import (
	"strings"
)

type Modifier int

const (
	ModifierPrivate Modifier = iota
	ModifierPackagePrivate
	ModifierProtected
	ModifierPublic
	ModifierAbstract
	ModifierStatic
	ModifierFinal
)

var accessModifiers = []Modifier{
	ModifierPrivate,
	ModifierPackagePrivate,
	ModifierProtected,
	ModifierPublic,
}

func formatModifier(modifier Modifier) string {
	switch modifier {
	case ModifierPrivate:
		return "private"
	case ModifierPackagePrivate:
		return ""
	case ModifierProtected:
		return "protected"
	case ModifierPublic:
		return "public"
	case ModifierAbstract:
		return "abstract"
	case ModifierStatic:
		return "static"
	case ModifierFinal:
		return "final"
	default:
		return ""
	}
}

func formatModifiers(modifiers []Modifier) string {
	// assume slice is already sorted and deduplicated
	var sb strings.Builder

	for i, mod := range modifiers {
		if formatted := formatModifier(mod); formatted != "" {
			sb.WriteString(formatted)
			if i != len(modifiers)-1 {
				sb.WriteString(" ")
			}
		}
	}

	return sb.String()
}
