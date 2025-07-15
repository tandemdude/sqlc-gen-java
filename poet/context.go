package poet

import (
	"context"
	"slices"
	"strings"
)

type Context struct {
	context.Context

	CurrentPackage  string
	CurrentTypeName string

	Imports []string
	Types   map[string]TypeName

	Indent string
}

func WithIndent(indent string) func(*Context) {
	return func(ctx *Context) {
		ctx.Indent = indent
	}
}

func NewContextFromContext(ctx context.Context, currentPackage string, options ...func(*Context)) *Context {
	newContext := &Context{
		Context:        ctx,
		CurrentPackage: currentPackage,
		Types:          make(map[string]TypeName),
	}

	for _, option := range options {
		option(newContext)
	}

	return newContext
}

func NewContext(currentPackage string, options ...func(*Context)) *Context {
	return NewContextFromContext(context.Background(), currentPackage, options...)
}

func (ctx *Context) Import(imports ...string) {
	// we don't need to import anything that is available within the current package
	for _, import_ := range imports {
		if import_ == ctx.CurrentPackage {
			continue
		}

		ctx.Imports = append(ctx.Imports, import_)
	}

	slices.Sort(ctx.Imports)
	ctx.Imports = slices.Compact(ctx.Imports)
}

func (ctx *Context) indent(text string) string {
	if len(strings.TrimSpace(text)) == 0 {
		return ""
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			lines[i] = ""
		} else {
			lines[i] = ctx.Indent + line
		}
	}

	return strings.Join(lines, "\n")
}
