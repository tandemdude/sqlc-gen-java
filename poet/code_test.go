package poet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeBuilder_StringFormatting(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
		args     []any
	}{
		{
			name:     "simple replacement",
			code:     "$T $L = $S",
			expected: `String test = "a value"`,
			args:     []any{String, "test", "a value"},
		},
		{
			name:     "complex replacement",
			code:     "$T $3L = $2S + $3S",
			expected: `String test = "a value" + "test"`,
			args:     []any{String, "a value", "test"},
		},
		{
			name:     "nil values",
			code:     "$1L $1S",
			expected: `null "null"`,
			args:     []any{nil},
		},
		{
			name:     "bool values",
			code:     "$1L $1S",
			expected: `true "true"`,
			args:     []any{true},
		},
		{
			name:     "int values",
			code:     "$1L $1S",
			expected: `1 "1"`,
			args:     []any{1},
		},
		{
			name:     "float values",
			code:     "$1L $1S",
			expected: `1.1 "1.1"`,
			args:     []any{1.1},
		},
		{
			name:     "string values",
			code:     "$1L $1S",
			expected: `hello "hello"`,
			args:     []any{"hello"},
		},
		{
			name:     "quoted strings",
			code:     "$1L $1S",
			expected: `"hello" "\"hello\""`,
			args:     []any{`"hello"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			ctx := NewContext("io.github.tandemdude")

			code := NewCodeBuilder().
				WithStatement(tt.code, tt.args...).
				Build()

			assert.Equal(tt.expected+";\n", code.Format(ctx))
		})
	}
}

func TestCodeBuilder_MultipleStatements(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("io.github.tandemdude")

	code := NewCodeBuilder().
		WithStatement("$L $S", false, true).
		WithStatement("$L $S", true, false).
		Build()

	assert.Equal("false \"true\";\ntrue \"false\";\n", code.Format(ctx))
}

func TestCodeBuilder_NoStatements(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("io.github.tandemdude")

	code := NewCodeBuilder().Build()

	assert.Equal("", code.Format(ctx))
}
