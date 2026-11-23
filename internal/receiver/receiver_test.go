package receiver_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/lugassawan/tidygo/internal/receiver"
)

func TestTypeName(t *testing.T) {
	cases := []struct {
		recv string
		want string
	}{
		{recv: "T", want: "T"},
		{recv: "*T", want: "T"},
		{recv: "(*T)", want: "T"},
		{recv: "(T)", want: "T"},
		{recv: "T[K]", want: "T"},
		{recv: "*T[K, V]", want: "T"},
		{recv: "(*T[K, V])", want: "T"},
		// A qualified or composite type has no local name to report.
		{recv: "pkg.T", want: ""},
		{recv: "[]T", want: ""},
	}
	for _, c := range cases {
		t.Run(c.recv, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "p.go", "package p\nfunc (r "+c.recv+") M() {}\n", 0)
			if err != nil {
				t.Fatalf("parsing source: %v", err)
			}
			fn, ok := file.Decls[0].(*ast.FuncDecl)
			if !ok {
				t.Fatalf("decl is %T, want *ast.FuncDecl", file.Decls[0])
			}
			if got := receiver.TypeName(fn.Recv.List[0].Type); got != c.want {
				t.Errorf("TypeName(%s) = %q, want %q", c.recv, got, c.want)
			}
		})
	}
}
