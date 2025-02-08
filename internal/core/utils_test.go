package core

import "testing"

func TestResolveImportAndType(t *testing.T) {
	cases := []string{
		"foo.bar.baz.Bork.Qux",
		"com.example.MyClass.InnerClass",
		"java.util.List",
		"SingleClass",
		"Nested.Class",
	}
	expected := [][]string{
		{"foo.bar.baz.Bork", "Bork.Qux"},
		{"com.example.MyClass", "MyClass.InnerClass"},
		{"java.util.List", "List"},
		{"", "SingleClass"},
		{"", "Nested.Class"},
	}

	for i, tc := range cases {
		imp, typ, err := ResolveImportAndType(tc)
		if err != nil {
			t.Fatal(err)
		}

		if imp != expected[i][0] {
			t.Errorf("case %d: expected '%s', got '%s'", i, expected[i][0], imp)
		}
		if typ != expected[i][1] {
			t.Errorf("case %d: expected '%s', got '%s'", i, expected[i][1], typ)
		}
	}
}

func TestAnnotate(t *testing.T) {
	cases := [][]string{
		{"Foo", "@Annotation", "@Annotation Foo"},
		{"Foo.Bar", "@Annotation", "Foo.@Annotation Bar"},
		{"org.example.Foo", "@Annotation", "org.example.@Annotation Foo"},
		{"Foo[]", "@Annotation", "Foo @Annotation []"},
	}

	for i, c := range cases {
		typ, annotation, expected := c[0], c[1], c[2]

		out := Annotate(typ, annotation)
		if out != expected {
			t.Errorf("case %d: expected '%s', got '%s'", i, expected, out)
		}
	}
}

type tc struct {
	Type            string
	Nullable        bool
	ExpectedType    string
	ExpectedUnboxed bool
}

func TestMaybeUnbox(t *testing.T) {
	cases := []tc{
		{"Integer", false, "int", true},
		{"Long", false, "long", true},
		{"Short", false, "short", true},
		{"Boolean", false, "boolean", true},
		{"Float", false, "float", true},
		{"Double", false, "double", true},
		{"FooBar", false, "FooBar", false},
		{"Integer", true, "Integer", false},
		{"Long", true, "Long", false},
		{"Short", true, "Short", false},
		{"Boolean", true, "Boolean", false},
		{"Float", true, "Float", false},
		{"Double", true, "Double", false},
		{"FooBar", true, "FooBar", false},
	}

	for i, c := range cases {
		newType, unboxed := MaybeUnbox(c.Type, c.Nullable)
		if newType != c.ExpectedType || unboxed != c.ExpectedUnboxed {
			t.Errorf("case %d: expected '%s' '%v', got '%s' '%v'", i, c.ExpectedType, c.ExpectedUnboxed, newType, unboxed)
		}
	}
}
