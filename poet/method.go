package poet

import (
	"strings"
)

type MethodParameter struct {
	Name string
	Type TypeName
}

func NewMethodParam(name string, typ TypeName) MethodParameter {
	return MethodParameter{Name: name, Type: typ}
}

func (m MethodParameter) Format(ctx *Context) string {
	return m.Type.Format(ctx, ExcludeConstraints) + " " + m.Name
}

type Method struct {
	Name string

	Modifiers         []Modifier
	GenericParameters []TypeName
	Parameters        []MethodParameter
	ReturnType        TypeName
	Throws            []TypeName

	Body Code
}

func (m Method) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString(formatModifiers(m.Modifiers))
	sb.WriteString(" ")

	if len(m.GenericParameters) > 0 {
		sb.WriteString("<")
		for i, param := range m.GenericParameters {
			sb.WriteString(param.Format(ctx, ExcludeConstraints))
			if i < len(m.GenericParameters)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("> ")
	}

	sb.WriteString(m.ReturnType.Format(ctx, ExcludeConstraints))
	sb.WriteString(" ")

	sb.WriteString(m.Name)

	sb.WriteString("(")
	for i, param := range m.Parameters {
		sb.WriteString(param.Format(ctx))
		if i < len(m.Parameters)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(") ")
	if len(m.Throws) > 0 {
		sb.WriteString("throws ")
		for i, throw := range m.Throws {
			sb.WriteString(throw.Format(ctx, ExcludeConstraints, ExcludeParameters, ExcludeArrayBraces))

			if i < len(m.Throws)-1 {
				sb.WriteString(",")
			}

			sb.WriteString(" ")
		}
	}

	if code := m.Body.Format(ctx); strings.TrimSpace(code) != "" {
		sb.WriteString("{\n")
		sb.WriteString(ctx.indent(code))
		sb.WriteString("}")
	} else {
		sb.WriteString("{}")
	}

	return sb.String()
}

type MethodBuilder struct {
	method Method
}

func NewMethodBuilder(name string, returnType TypeName) *MethodBuilder {
	return &MethodBuilder{method: Method{Name: name, ReturnType: returnType}}
}

func (b *MethodBuilder) WithModifiers(modifiers ...Modifier) *MethodBuilder {
	b.method.Modifiers = appendModifiers(b.method.Modifiers, modifiers)
	return b
}

func (b *MethodBuilder) WithCode(code Code) *MethodBuilder {
	b.method.Body = code
	return b
}

func (b *MethodBuilder) WithGenericParameters(parameters ...TypeName) *MethodBuilder {
	b.method.GenericParameters = append(b.method.GenericParameters, parameters...)
	return b
}

func (b *MethodBuilder) WithParameters(params ...MethodParameter) *MethodBuilder {
	b.method.Parameters = append(b.method.Parameters, params...)
	return b
}

func (b *MethodBuilder) WithThrows(throws ...TypeName) *MethodBuilder {
	b.method.Throws = append(b.method.Throws, throws...)
	return b
}

func (b *MethodBuilder) Build() Method {
	b.method.Modifiers = maybeSetPackagePrivate(b.method.Modifiers)
	return b.method
}

type Constructor Method

func (c Constructor) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString(formatModifiers(c.Modifiers))
	if len(c.Modifiers) > 0 {
		sb.WriteString(" ")
	}

	sb.WriteString(c.Name)
	sb.WriteString("(")

	for i, param := range c.Parameters {
		sb.WriteString(param.Format(ctx))
		if i < len(c.Parameters)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(") ")

	if len(c.Throws) > 0 {
		sb.WriteString("throws ")
		for i, throw := range c.Throws {
			sb.WriteString(throw.Format(ctx, ExcludeConstraints, ExcludeParameters, ExcludeArrayBraces))

			if i < len(c.Throws)-1 {
				sb.WriteString(",")
			}

			sb.WriteString(" ")
		}
	}

	if code := c.Body.Format(ctx); strings.TrimSpace(code) != "" {
		sb.WriteString(") {\n")
		sb.WriteString(ctx.indent(code))
		sb.WriteString("\n}")
	} else {
		sb.WriteString(") {}")
	}

	return sb.String()
}

type ConstructorBuilder struct {
	constructor Constructor
}

func NewConstructorBuilder() *ConstructorBuilder {
	return &ConstructorBuilder{constructor: Constructor{}}
}

func (b *ConstructorBuilder) WithModifiers(modifiers ...Modifier) *ConstructorBuilder {
	b.constructor.Modifiers = appendModifiers(b.constructor.Modifiers, modifiers)
	return b
}

func (b *ConstructorBuilder) WithParameters(params ...MethodParameter) *ConstructorBuilder {
	b.constructor.Parameters = append(b.constructor.Parameters, params...)
	return b
}

func (b *ConstructorBuilder) WithThrows(throws ...TypeName) *ConstructorBuilder {
	b.constructor.Throws = append(b.constructor.Throws, throws...)
	return b
}

func (b *ConstructorBuilder) Build() Constructor {
	b.constructor.Modifiers = maybeSetPackagePrivate(b.constructor.Modifiers)
	return b.constructor
}
