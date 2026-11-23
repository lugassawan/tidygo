package example

import "testing"

// Not reported: test files are skipped, table-test scaffolding repeats on purpose.
func totalInTest(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

func TestTotal(t *testing.T) {
	if totalInTest(nil) != 0 {
		t.Fatal("unexpected total")
	}
}
