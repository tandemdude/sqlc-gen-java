package poet

import (
	"strings"
)

type formattable interface {
	name() string
	Format(ctx *Context) string
}

type FileOptions struct {
	Comment string
}

func WithFileComment(comment string) func(*FileOptions) {
	return func(o *FileOptions) {
		o.Comment = comment
	}
}

func FormatFile(ctx *Context, member formattable, options ...func(*FileOptions)) string {
	ctx.CurrentTypeName = member.name()

	opts := &FileOptions{}
	for _, o := range options {
		o(opts)
	}

	var sb strings.Builder

	if opts.Comment != "" {
		sb.WriteString(opts.Comment)
		if opts.Comment[len(opts.Comment)-1] != '\n' {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("package " + ctx.CurrentPackage + ";\n\n")

	memberString := member.Format(ctx)
	for i, imp := range ctx.Imports {
		sb.WriteString("import " + imp + ";\n")
		if i == len(ctx.Imports)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString(memberString)
	return sb.String()
}
