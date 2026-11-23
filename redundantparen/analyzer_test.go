package redundantparen_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/redundantparen"
)

// doubledSource holds doubled parentheses. gofmt collapses ((x)) to (x), so
// the source lives in a string instead of a checked-in (and formatted) fixture.
const doubledSource = `package doubled

func use(...any) {}

func returned(a int) int {
	return ((a)) // want ` + "`redundant parentheses`" + `
}

func tripled(a int) int {
	return (((a))) // want ` + "`redundant parentheses`" + `
}

func inOperand(a, b int) int {
	return a * ((b + 1)) // want ` + "`redundant parentheses`" + `
}

func inCondition(a int) {
	if ((a > 0)) { // want ` + "`redundant parentheses`" + `
		use(a)
	}
}

func single(a, b int) int {
	return a * (b + 1)
}
`

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, redundantparen.Analyzer, "example")
}

func TestAnalyzerDoubledParentheses(t *testing.T) {
	dir, cleanup, err := analysistest.WriteFiles(map[string]string{"doubled/doubled.go": doubledSource})
	if err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	t.Cleanup(cleanup)
	analysistest.Run(t, dir, redundantparen.Analyzer, "doubled")
}
