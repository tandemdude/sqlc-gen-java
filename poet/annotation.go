package poet

import "strings"

type Annotation struct {
	Class   TypeName
	Members map[string]Code
}

func (a Annotation) Format(ctx *Context) string {
	var sb strings.Builder

	sb.WriteRune('@')
	sb.WriteString(a.Class.Format(ctx))

	// FIXME: Add support for embedded annotations
	// .addAnnotation(AnnotationSpec.builder(HeaderList.class)
	//        .addMember("value", "$L", AnnotationSpec.builder(Header.class)
	//            .addMember("name", "$S", "Accept")
	//            .addMember("value", "$S", "application/json; charset=utf-8")
	//            .build())
	//        .addMember("value", "$L", AnnotationSpec.builder(Header.class)
	//            .addMember("name", "$S", "User-Agent")
	//            .addMember("value", "$S", "Square Cash")
	//            .build())
	//        .build())
	if value, ok := a.Members["value"]; len(a.Members) == 1 && ok {
		sb.WriteRune('(')
		sb.WriteString(value.Format(ctx))
		sb.WriteRune(')')
	} else if len(a.Members) > 0 {
		sb.WriteString("(\n")
		for name, code := range a.Members {
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(code.Format(ctx))
		}
		sb.WriteString("\n)")
	}

	return sb.String()
}

type AnnotationBuilder struct {
	annotation Annotation
}

func NewAnnotationBuilder(class TypeName) *AnnotationBuilder {
	return &AnnotationBuilder{annotation: Annotation{Class: class, Members: make(map[string]Code)}}
}

func (b *AnnotationBuilder) WithMember(name string, value string, args ...any) *AnnotationBuilder {
	b.annotation.Members[name] = Code{RawCode: value, Arguments: args}
	return b
}

func (b *AnnotationBuilder) Build() Annotation {
	return b.annotation
}
