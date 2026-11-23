package dupfunc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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

// A function the type checker recorded no object for has no signature, and a
// missed report costs less than a false one, so it matches nothing in either role.
func TestFindMatchWithoutSignature(t *testing.T) {
	pkg := typeCheck(t, "func a(x int) int { return x }\nfunc b(x int) int { return x }")
	withSig := func(name string) *candidate {
		fn, ok := pkg.Scope().Lookup(name).(*types.Func)
		if !ok {
			t.Fatalf("%s is not a function", name)
		}
		return &candidate{fn: fn}
	}
	noSig := &candidate{}

	// Control: two functions with identical signatures do match.
	if findMatch([]*candidate{withSig("a")}, withSig("b")) == nil {
		t.Fatal("findMatch(a, b) = nil, want a match for identical signatures")
	}
	cases := []struct {
		name  string
		group []*candidate
		c     *candidate
	}{
		{name: "empty group", group: nil, c: withSig("a")},
		{name: "candidate without signature", group: []*candidate{withSig("a")}, c: noSig},
		{name: "group member without signature", group: []*candidate{noSig}, c: withSig("a")},
		{name: "neither has a signature", group: []*candidate{noSig}, c: &candidate{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := findMatch(c.group, c.c); got != nil {
				t.Errorf("findMatch = %v, want nil", got)
			}
		})
	}
}

// typeCheck type-checks src, a package body without imports, as package p.
func typeCheck(t *testing.T, src string) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n"+src+"\n", 0)
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}
	pkg, err := (&types.Config{}).Check("p", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("type-checking: %v", err)
	}
	return pkg
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
