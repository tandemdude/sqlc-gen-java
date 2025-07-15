package poet

import (
	"strings"
)

type Code struct {
	RawCode    string
	IsFlow     bool
	IsTryCatch bool
	IsIfElse   bool

	Arguments []any

	Statements []Code
}

func formatRawCode(ctx *Context, rawCode string, arguments []any) string {
	return ""
}

func formatStatements(ctx *Context, statements []Code) string {
	var sb strings.Builder

	for i, stmt := range statements {
		sb.WriteString(stmt.Format(ctx))

		inlineNextStmt := false
		if stmt.IsFlow && len(statements) > i+1 {
			nextStmt := statements[i+1]

			inlineNextStmt = nextStmt.IsFlow && ((stmt.IsIfElse && nextStmt.IsIfElse) || (stmt.IsTryCatch && nextStmt.IsTryCatch))
		}

		if inlineNextStmt {
			sb.WriteString(" ")
		} else {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (c *Code) Format(ctx *Context) string {
	var sb strings.Builder

	if c.IsFlow {
		// Control flow statement
		sb.WriteString(c.RawCode)
		sb.WriteString(" {\n")
		sb.WriteString(ctx.indent(formatStatements(ctx, c.Statements)))
		sb.WriteString("}")

		return sb.String()
	}

	if c.RawCode != "" && !c.IsFlow {
		// Simple statement
		sb.WriteString(c.RawCode)
		if !strings.HasSuffix(c.RawCode, ";") {
			sb.WriteRune(';')
		}

		return sb.String()
	}

	// List of statements
	sb.WriteString(formatStatements(ctx, c.Statements))

	return sb.String()
}

type CodeBuilder struct {
	code Code
}

func NewCodeBuilder() *CodeBuilder {
	return &CodeBuilder{}
}

func (b *CodeBuilder) AddStatement(stmt string, args ...any) *CodeBuilder {
	b.code.Statements = append(b.code.Statements, Code{RawCode: stmt, Arguments: args})
	return b
}

func (b *CodeBuilder) WithControlFlow(stmt string, blockBuilderFn func(*CodeBuilder), args ...any) *CodeBuilder {
	builder := NewCodeBuilder()
	builder.code.Arguments = args
	builder.code.RawCode = stmt
	builder.code.IsFlow = true

	if strings.HasPrefix(stmt, "if") || strings.HasPrefix(stmt, "else") {
		builder.code.IsIfElse = true
	} else if strings.HasPrefix(stmt, "try") || strings.HasPrefix(stmt, "catch") || strings.HasPrefix(stmt, "finally") {
		builder.code.IsTryCatch = true
	}

	blockBuilderFn(builder)

	b.code.Statements = append(b.code.Statements, builder.Build())
	return b
}

func (b *CodeBuilder) Build() Code {
	return b.code
}
