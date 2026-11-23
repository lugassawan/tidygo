package funcs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/funcs"
)

const src = `package p

func declared() {
	_ = func() {
		_ = func() {}
	}
}

func external() // no body

type T struct{}

func (T) method() {}
`

func TestEach(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}

	var got []int // line of each visited body
	funcs.Each(inspector.New([]*ast.File{file}), func(ftype *ast.FuncType, body *ast.BlockStmt) {
		if ftype == nil || body == nil {
			t.Errorf("Each passed a nil signature or body: %v %v", ftype, body)
			return
		}
		got = append(got, fset.Position(body.Lbrace).Line)
	})

	// declared, then the literal inside it, then the literal inside that, then the
	// method. external has no body and is skipped.
	want := []int{3, 4, 5, 13}
	if !slices.Equal(got, want) {
		t.Errorf("Each visited bodies at lines %v, want %v", got, want)
	}
}
