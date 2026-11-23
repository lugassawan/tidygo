package dupfunc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// methodOfT is how a method a on type T is rendered.
const methodOfT = "T.a"

func TestFuncName(t *testing.T) {
	cases := []struct {
		name string
		decl string
		want string
	}{
		{name: "function", decl: "func a() {}", want: "a"},
		{name: "method", decl: "func (r T) a() {}", want: methodOfT},
		{name: "pointer receiver", decl: "func (r *T) a() {}", want: methodOfT},
		{name: "generic receiver", decl: "func (r *T[K]) a() {}", want: methodOfT},
		// The parser accepts these receivers but the type checker rejects them, and neither
		// has a local type name. The name falls back to the bare method name, not ".a".
		{name: "qualified receiver", decl: "func (r pkg.T) a() {}", want: "a"},
		{name: "composite receiver", decl: "func (r []T) a() {}", want: "a"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := funcName(parseFuncDecl(t, c.decl)); got != c.want {
				t.Errorf("funcName(%q) = %q, want %q", c.decl, got, c.want)
			}
		})
	}
}

// A method with an empty receiver list is malformed and rejected by the parser's
// own checks, so only a hand-built declaration can reach the guard.
func TestFuncNameEmptyReceiverList(t *testing.T) {
	fd := &ast.FuncDecl{Name: ast.NewIdent("a"), Recv: &ast.FieldList{}}
	if got := funcName(fd); got != "a" {
		t.Errorf("funcName(empty receiver list) = %q, want %q", got, "a")
	}
}

func parseFuncDecl(t *testing.T, decl string) *ast.FuncDecl {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "p.go", "package p\n"+decl+"\n", 0)
	if err != nil {
		t.Fatalf("parsing %q: %v", decl, err)
	}
	fd, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("decl is %T, want *ast.FuncDecl", file.Decls[0])
	}
	return fd
}
