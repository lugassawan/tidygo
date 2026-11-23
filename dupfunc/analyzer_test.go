package dupfunc_test

import (
	"strconv"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/dupfunc"
)

// parenReceiverSource declares a method on a parenthesised receiver type. gofmt
// strips those parentheses, so the source lives in a string instead of a
// checked-in (and formatted) fixture.
const parenReceiverSource = `package paren

type left struct{ v []int }

type right struct{ v []int }

func (x *left) count() int {
	n := 0
	for range x.v {
		n++
	}
	return n
}

func (x (*right)) count() int { // want ` + "`right\\.count has the same body as left\\.count \\(paren\\.go:7\\)`" + `
	n := 0
	for range x.v {
		n++
	}
	return n
}
`

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, dupfunc.Analyzer, "example")
}

func TestAnalyzerParenthesisedReceiver(t *testing.T) {
	dir, cleanup, err := analysistest.WriteFiles(map[string]string{"paren/paren.go": parenReceiverSource})
	if err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, dir, dupfunc.Analyzer, "paren")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  dupfunc.Config
		pkg  string
	}{
		{
			name: "a positive minimum replaces the default",
			cfg:  dupfunc.Config{MinStmts: 2},
			pkg:  "custom",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, dupfunc.New(c.cfg), c.pkg)
		})
	}
}

func TestNewNonPositiveMinStmtsUsesDefault(t *testing.T) {
	testdata := analysistest.TestData()
	for _, minStmts := range []int{0, -1} {
		// The example package is written against the default minimum of 3 statements.
		t.Run(strconv.Itoa(minStmts), func(t *testing.T) {
			analysistest.Run(t, testdata, dupfunc.New(dupfunc.Config{MinStmts: minStmts}), "example")
		})
	}
}
