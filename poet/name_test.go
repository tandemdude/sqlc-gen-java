package poet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatClassName_InCurrentPackage(t *testing.T) {
	assert := assert.New(t)

	name := NewClassName("io.github.tandemdude", "Foo")

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("Foo", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatClassName_NotInCurrentPackage(t *testing.T) {
	assert := assert.New(t)

	name := NewClassName("io.github.tandemdude", "Foo")

	ctx := NewContext("com.example")
	assert.Equal("Foo", name.Format(ctx))
	assert.Len(ctx.Imports, 1)
	assert.Equal("io.github.tandemdude", ctx.Imports[0])
}

func TestFormatClassName_ConflictingNames(t *testing.T) {
	assert := assert.New(t)

	name := NewClassName("io.github.tandemdude", "Foo")
	name2 := NewClassName("io.github.davfsa", "Foo")

	ctx := NewContext("com.example")
	assert.Equal("Foo", name.Format(ctx))
	assert.Equal("io.github.davfsa.Foo", name2.Format(ctx))
	assert.Len(ctx.Imports, 1)
	assert.Equal("io.github.tandemdude", ctx.Imports[0])
}

func TestFormatParameterizedClassName_NoParameters(t *testing.T) {
	assert := assert.New(t)

	name := NewParameterizedClassName("java.util", "List")

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("List<>", name.Format(ctx))
	assert.Len(ctx.Imports, 1)
	assert.Equal("java.util", ctx.Imports[0])
}

func TestFormatParameterizedClassName_SingleParameter(t *testing.T) {
	assert := assert.New(t)

	name := NewParameterizedClassName("java.util", "List", NewClassName("io.github.tandemdude", "Foo"))

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("List<Foo>", name.Format(ctx))
	assert.Len(ctx.Imports, 1)
	assert.Equal("java.util", ctx.Imports[0])
}

func TestFormatParameterizedClassName_MultipleParameters(t *testing.T) {
	assert := assert.New(t)

	name := NewParameterizedClassName(
		"java.util", "Map",
		NewClassName("io.github.tandemdude", "Foo"),
		NewClassName("io.github.tandemdude", "Bar"),
	)

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("Map<Foo, Bar>", name.Format(ctx))
	assert.Len(ctx.Imports, 1)
	assert.Equal("java.util", ctx.Imports[0])
}

func TestFormatGenericParam_NoConstraints(t *testing.T) {
	assert := assert.New(t)

	name := NewGenericParam("T")

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("T", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatGenericParam_OneConstraint(t *testing.T) {
	assert := assert.New(t)

	name := NewGenericParam("T", NewClassName("io.github.tandemdude", "Foo"))

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("T extends Foo", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatGenericParam_MultipleConstraints(t *testing.T) {
	assert := assert.New(t)

	name := NewGenericParam(
		"T",
		NewClassName("io.github.tandemdude", "Foo"),
		NewClassName("io.github.tandemdude", "Bar"),
	)

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("T extends Foo & Bar", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatClassName_IsArray(t *testing.T) {
	assert := assert.New(t)

	name := NewClassName("io.github.tandemdude", "Foo").Array()

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("Foo[]", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatParameterizedClassName_IsArray(t *testing.T) {
	assert := assert.New(t)

	name := NewParameterizedClassName("io.github.tandemdude", "Foo", String).Array()

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("Foo<String>[]", name.Format(ctx))
	assert.Empty(ctx.Imports)
}

func TestFormatGenericParam_IsArray(t *testing.T) {
	assert := assert.New(t)

	name := NewGenericParam("T").Array()

	ctx := NewContext("io.github.tandemdude")
	assert.Equal("T[]", name.Format(ctx, ExcludeConstraints))
	assert.Empty(ctx.Imports)
}

func TestClassName_SimpleEquals(t *testing.T) {
	assert := assert.New(t)

	name := NewClassName("", "Foo")
	pName := NewParameterizedClassName("", "Foo")
	gName := NewGenericParam("Foo")
	aName := name.Array()

	// the same instances are equal to each other
	assert.Equal(name, name)
	assert.Equal(pName, pName)
	assert.Equal(gName, gName)
	assert.Equal(aName, aName)

	// different types of class names are not equal
	assert.NotEqual(name, pName)
	assert.NotEqual(name, gName)
	assert.NotEqual(name, aName)

	assert.NotEqual(pName, gName)
	assert.NotEqual(pName, aName)

	assert.NotEqual(gName, aName)
}

func TestParameterizedClassName_Equals(t *testing.T) {
	assert := assert.New(t)

	name := NewParameterizedClassName("", "Foo")
	name2 := NewParameterizedClassName("", "Foo", String)

	// types with different parameter counts are not equal
	assert.NotEqual(name, name2)

	name = NewParameterizedClassName("", "Foo", String)
	name2 = NewParameterizedClassName("", "Foo", IntBoxed)

	// types with the same parameter count, but different parameters are not equal
	assert.NotEqual(name, name2)
}

func TestGenericParam_Equals(t *testing.T) {
	assert := assert.New(t)

	name := NewGenericParam("T")
	name2 := NewGenericParam("T", String)

	// generic parameters with different constraint counts are not equal
	assert.NotEqual(name, name2)

	name = NewGenericParam("T", String)
	name2 = NewGenericParam("T", IntBoxed)

	// generic parameters with the same constraint counts, but different constraints are not equal
	assert.NotEqual(name, name2)
}
