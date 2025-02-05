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
