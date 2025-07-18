package poet

import (
	"fmt"
	"strings"
)

// TODO - annotation support

type ClassField struct {
	Name      string
	Type      TypeName
	Modifiers []Modifier
}

type Class struct {
	Name string

	Modifiers         []Modifier
	GenericParameters []TypeName
	Constructor       *Constructor
	Fields            []ClassField
	Methods           []Method
}

func (c Class) name() string {
	return c.Name
}

func (c Class) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString(formatModifiers(c.Modifiers))
	if sb.Len() > 0 {
		sb.WriteString(" ")
	}

	sb.WriteString("class ")
	sb.WriteString(c.Name)
	writeGenericParamList(ctx, &sb, c.GenericParameters, false)
	sb.WriteString(" {\n")

	for i, field := range c.Fields {
		sb.WriteString(ctx.indent(fmt.Sprintf(
			"%s %s %s;\n",
			formatModifiers(field.Modifiers),
			field.Type.Format(ctx, ExcludeConstraints),
			field.Name,
		)))

		if i == len(c.Fields)-1 {
			sb.WriteString("\n")
		}
	}

	if c.Constructor != nil {
		c.Constructor.Name = c.Name
		sb.WriteString(ctx.indent(c.Constructor.Format(ctx)))
		sb.WriteString("\n")
	}

	if len(c.Methods) > 0 {
		sb.WriteString("\n")
	}

	for i, method := range c.Methods {
		sb.WriteString(ctx.indent(method.Format(ctx)))
		sb.WriteString("\n")
		if i != len(c.Methods)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("}")
	return sb.String()
}

type ClassBuilder struct {
	class Class
}

func NewClassBuilder(name string) *ClassBuilder {
	return &ClassBuilder{class: Class{Name: name}}
}

func (c *ClassBuilder) WithModifiers(modifiers ...Modifier) *ClassBuilder {
	c.class.Modifiers = appendModifiers(c.class.Modifiers, modifiers)
	return c
}

func (c *ClassBuilder) WithGenericParameters(parameters ...TypeName) *ClassBuilder {
	c.class.GenericParameters = append(c.class.GenericParameters, parameters...)
	return c
}

func (c *ClassBuilder) WithConstructor(constructor Constructor) *ClassBuilder {
	c.class.Constructor = &constructor
	return c
}

func (c *ClassBuilder) WithFields(fields ...ClassField) *ClassBuilder {
	c.class.Fields = append(c.class.Fields, fields...)
	return c
}

func (c *ClassBuilder) WithMethods(methods ...Method) *ClassBuilder {
	c.class.Methods = append(c.class.Methods, methods...)
	return c
}

func (c *ClassBuilder) Build() Class {
	c.class.Modifiers = maybeSetPackagePrivate(c.class.Modifiers)
	return c.class
}

type EnumValue struct {
	Name string
	// for now, only support string enums given that is the only type supported by the databases
	Value string
}

func NewEnumValue(name string, value string) EnumValue {
	return EnumValue{Name: name, Value: value}
}

type Enum struct {
	Name string

	Modifiers []Modifier
	Values    []EnumValue
	Methods   []Method
}

func (e Enum) name() string {
	return e.Name
}

func (e Enum) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString(formatModifiers(e.Modifiers))
	if sb.Len() > 0 {
		sb.WriteString(" ")
	}

	sb.WriteString("enum ")
	sb.WriteString(e.Name)
	sb.WriteString(" {\n")

	for i, v := range e.Values {
		sb.WriteString(ctx.indent(fmt.Sprintf("%s(\"%s\")", v.Name, v.Value)))

		if i != len(e.Values)-1 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString(";\n\n")
		}
	}

	sb.WriteString(ctx.indent("private final String value;\n\n"))
	sb.WriteString(ctx.indent(e.Name))
	sb.WriteString("(final String value) { this.value = value; }\n")

	if len(e.Methods) > 0 {
		sb.WriteString("\n")
	}

	for i, method := range e.Methods {
		sb.WriteString(ctx.indent(method.Format(ctx)))
		sb.WriteString("\n")
		if i != len(e.Methods)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("}")

	return sb.String()
}

type EnumBuilder struct {
	enum Enum
}

func NewEnumBuilder(name string) *EnumBuilder {
	return &EnumBuilder{enum: Enum{Name: name}}
}

func (b *EnumBuilder) WithModifiers(modifiers ...Modifier) *EnumBuilder {
	b.enum.Modifiers = appendModifiers(b.enum.Modifiers, modifiers)
	return b
}

func (b *EnumBuilder) WithValue(name string, value string) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, EnumValue{Name: name, Value: value})
	return b
}

func (b *EnumBuilder) WithValues(values ...EnumValue) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, values...)
	return b
}

func (b *EnumBuilder) WithMethods(methods ...Method) *EnumBuilder {
	b.enum.Methods = append(b.enum.Methods, methods...)
	return b
}

func (b *EnumBuilder) Build() Enum {
	b.enum.Modifiers = maybeSetPackagePrivate(b.enum.Modifiers)
	return b.enum
}

type Record struct {
	Name string

	Modifiers  []Modifier
	Parameters []MethodParameter
	Methods    []Method
}

func (r Record) name() string {
	return r.Name
}

func (r Record) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteString(formatModifiers(r.Modifiers))
	if sb.Len() > 0 {
		sb.WriteString(" ")
	}

	// TODO - generic parameter support?

	sb.WriteString("record ")
	sb.WriteString(r.Name)
	sb.WriteString("(")

	for i, param := range r.Parameters {
		sb.WriteString(param.Format(ctx))
		if i != len(r.Parameters)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(") ")

	if len(r.Methods) == 0 {
		sb.WriteString("{}")
		return sb.String()
	}

	for i, method := range r.Methods {
		sb.WriteString(ctx.indent(method.Format(ctx)))
		sb.WriteString("\n")
		if i != len(r.Methods)-1 {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("}")

	return sb.String()
}

type RecordBuilder struct {
	record Record
}

func NewRecordBuilder(name string) *RecordBuilder {
	return &RecordBuilder{record: Record{Name: name}}
}

func (b *RecordBuilder) WithModifiers(modifiers ...Modifier) *RecordBuilder {
	b.record.Modifiers = appendModifiers(b.record.Modifiers, modifiers)
	return b
}

func (b *RecordBuilder) WithParameters(params ...MethodParameter) *RecordBuilder {
	b.record.Parameters = append(b.record.Parameters, params...)
	return b
}

func (b *RecordBuilder) WithMethods(methods ...Method) *RecordBuilder {
	b.record.Methods = append(b.record.Methods, methods...)
	return b
}

func (b *RecordBuilder) Build() Record {
	b.record.Modifiers = maybeSetPackagePrivate(b.record.Modifiers)
	return b.record
}
