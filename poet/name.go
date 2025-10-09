package poet

import (
	"strings"
)

type TypeName struct {
	Package string
	Name    string

	IsBuiltin bool
	IsArray   bool

	// FIXME: We need builders for this
	Annotations []Annotation

	IsParameterized bool
	Parameters      []TypeName

	IsGeneric bool
	Extends   []TypeName
}

func NewClassName(pkg, name string, annotations ...Annotation) TypeName {
	return TypeName{Package: pkg, Name: name, Annotations: annotations}
}

func (t TypeName) Array() TypeName {
	t.IsArray = true
	return t
}

func NewParameterizedClassName(pkg, name string, parameters ...TypeName) TypeName {
	return TypeName{
		Package:         pkg,
		Name:            name,
		IsParameterized: true,
		Parameters:      parameters,
	}
}

func NewGenericParam(name string, extends ...TypeName) TypeName {
	return TypeName{
		Package:   "",
		Name:      name,
		IsGeneric: true,
		Extends:   extends,
	}
}

type FormatOption int

var (
	ExcludeConstraints FormatOption = 1 << 0
	ExcludeParameters  FormatOption = 1 << 1
	ExcludeArrayBraces FormatOption = 1 << 2
)

func (opt FormatOption) has(other FormatOption) bool {
	return opt&other != 0
}

func (t TypeName) Format(ctx *Context, options ...FormatOption) string {
	var opts FormatOption
	for _, opt := range options {
		opts |= opt
	}

	var bld strings.Builder

	if t.Package != "" {
		// check if we need to use the fully qualified type name due to a collision
		existing, ok := ctx.Types[t.Name]
		if (ok && !(existing.Package == t.Package && existing.Name == t.Name)) || t.Name == ctx.CurrentTypeName {
			bld.WriteString(t.Package)
			bld.WriteString(".")
		} else {
			if !t.IsBuiltin {
				ctx.Import(t.Package + "." + t.Name)
				ctx.Types[t.Name] = t
			}
		}
	}

	for _, annotation := range t.Annotations {
		bld.WriteString(annotation.Format(ctx))
		bld.WriteString(" ")
	}

	bld.WriteString(t.Name)
	if t.IsGeneric && len(t.Extends) > 0 && !opts.has(ExcludeConstraints) {
		bld.WriteString(" extends ")
		for i, extend := range t.Extends {
			// TODO - can a generic parameter have its own generic constraint?
			bld.WriteString(extend.Format(ctx, opts))
			if i < len(t.Extends)-1 {
				bld.WriteString(" & ")
			}
		}
	} else if t.IsParameterized && !opts.has(ExcludeParameters) {
		bld.WriteString("<")
		for i, param := range t.Parameters {
			bld.WriteString(param.Format(ctx, opts))
			if i < len(t.Parameters)-1 {
				bld.WriteString(", ")
			}
		}
		bld.WriteString(">")
	}

	if !opts.has(ExcludeArrayBraces) && t.IsArray && !(t.IsGeneric && !opts.has(ExcludeConstraints)) {
		bld.WriteString("[]")
	}

	return bld.String()
}

var (
	Bool        = TypeName{Name: "boolean", IsBuiltin: true}
	BoolBoxed   = TypeName{Package: "java.lang", Name: "Boolean", IsBuiltin: true}
	Byte        = TypeName{Name: "byte", IsBuiltin: true}
	ByteBoxed   = TypeName{Package: "java.lang", Name: "Byte", IsBuiltin: true}
	Char        = TypeName{Name: "char", IsBuiltin: true}
	CharBoxed   = TypeName{Package: "java.lang", Name: "Character", IsBuiltin: true}
	Double      = TypeName{Name: "double", IsBuiltin: true}
	DoubleBoxed = TypeName{Package: "java.lang", Name: "Double", IsBuiltin: true}
	Float       = TypeName{Name: "float", IsBuiltin: true}
	FloatBoxed  = TypeName{Package: "java.lang", Name: "Float", IsBuiltin: true}
	Int         = TypeName{Name: "int", IsBuiltin: true}
	IntBoxed    = TypeName{Package: "java.lang", Name: "Integer", IsBuiltin: true}
	Long        = TypeName{Name: "long", IsBuiltin: true}
	LongBoxed   = TypeName{Package: "java.lang", Name: "Long", IsBuiltin: true}
	Object      = TypeName{Package: "java.lang", Name: "Object", IsBuiltin: true}
	Short       = TypeName{Name: "short", IsBuiltin: true}
	ShortBoxed  = TypeName{Package: "java.lang", Name: "Short", IsBuiltin: true}
	String      = TypeName{Package: "java.lang", Name: "String", IsBuiltin: true}
	Void        = TypeName{Name: "void", IsBuiltin: true}
	VoidBoxed   = TypeName{Package: "java.lang", Name: "Void", IsBuiltin: true}
	Wildcard    = TypeName{Name: "?", IsBuiltin: true}
)

func newSingleParameterType(pkg, name string) func(TypeName) TypeName {
	return func(t TypeName) TypeName {
		return NewParameterizedClassName(pkg, name, t)
	}
}

func newTwoParameterType(pkg, name string) func(TypeName, TypeName) TypeName {
	return func(t1 TypeName, t2 TypeName) TypeName {
		return NewParameterizedClassName(pkg, name, t1, t2)
	}
}

var (
	ListOf     = newSingleParameterType("java.util", "List")
	MapOf      = newTwoParameterType("java.util", "Map")
	SetOf      = newSingleParameterType("java.util", "Set")
	OptionalOf = newSingleParameterType("java.util", "Optional")
)
