package poet

import (
	"slices"
	"strings"
)

func appendModifiers(initial []Modifier, new []Modifier) []Modifier {
	initial = append(initial, new...)
	slices.Sort(initial)
	return slices.Compact(initial)
}

func writeGenericParamList(ctx *Context, sb *strings.Builder, params []TypeName, includeTrailingSpace bool) {
	if len(params) == 0 {
		return
	}

	sb.WriteString("<")
	for i, param := range params {
		sb.WriteString(param.Format(ctx, ExcludeConstraints))
		if i < len(params)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(">")
	if includeTrailingSpace {
		sb.WriteString(" ")
	}
}
