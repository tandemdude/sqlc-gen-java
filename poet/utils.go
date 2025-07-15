package poet

import (
	"slices"
)

func appendModifiers(initial []Modifier, new []Modifier) []Modifier {
	initial = append(initial, new...)
	slices.Sort(initial)
	return slices.Compact(initial)
}

func maybeSetPackagePrivate(modifiers []Modifier) []Modifier {
	if !slices.ContainsFunc(modifiers, func(m Modifier) bool {
		return slices.Contains(accessModifiers, m)
	}) {
		modifiers = append(modifiers, ModifierPackagePrivate)
		slices.Sort(modifiers)
	}
	return modifiers
}
