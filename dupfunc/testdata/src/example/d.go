package example

type left struct{ v []string }

func (x *left) count() int {
	n := 0
	for range x.v {
		n++
	}
	return n
}

type right struct{ v []string }

// Reported: methods are named with their receiver type.
func (x *right) count() int { // want `function right\.count has the same body as left\.count \(d\.go:5\)`
	n := 0
	for range x.v {
		n++
	}
	return n
}
