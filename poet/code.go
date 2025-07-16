package poet

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	replaceTypeLiteral rune = 'L'
	replaceTypeString  rune = 'S'
	replaceTypeType    rune = 'T'
)

var stringFormatRegex = regexp.MustCompile(`\$(?<pos>\d*)?(?<type>[LST])`)

type Code struct {
	RawCode    string
	IsFlow     bool
	IsTryCatch bool
	IsIfElse   bool

	Arguments []any

	Statements []Code
}

func stringify(raw any) string {
	if raw == nil {
		return "null"
	}

	return fmt.Sprintf("%v", raw)
}

func formatRawCode(ctx *Context, rawCode string, arguments []any) string {
	matchIndex := 0

	return stringFormatRegex.ReplaceAllStringFunc(rawCode, func(match string) string {
		replaceType := rune(match[len(match)-1])
		argumentIndex := 0
		for i := 1; i < len(match)-1; i++ {
			argumentIndex = (argumentIndex * 10) + int(match[i]-'0')
		}

		if argumentIndex > 0 {
			// So they dont have to be 0 indexed
			argumentIndex -= 1
		} else {
			argumentIndex = matchIndex
		}

		if argumentIndex >= len(arguments) {
			// Tried to access an argument that is not there
			return match
		}

		replacement := match

		switch replaceType {
		case replaceTypeLiteral:
			replacement = stringify(arguments[argumentIndex])
		case replaceTypeString:
			replacement = fmt.Sprintf("%q", stringify(arguments[argumentIndex]))
		case replaceTypeType:
			replacement = arguments[argumentIndex].(TypeName).Format(ctx)
		}

		if len(match) == 2 {
			matchIndex += 1
		}

		return replacement
	})
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
		sb.WriteString(formatRawCode(ctx, c.RawCode, c.Arguments))
		sb.WriteString(" {\n")
		sb.WriteString(ctx.indent(formatStatements(ctx, c.Statements)))
		sb.WriteString("}")

		return sb.String()
	}

	if c.RawCode != "" && !c.IsFlow {
		// Simple statement
		sb.WriteString(formatRawCode(ctx, c.RawCode, c.Arguments))
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
