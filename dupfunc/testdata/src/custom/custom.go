package custom

func first(x int) int {
	y := x + 1
	return y
}

// Reported: two statements reach the configured minimum of 2.
func second(x int) int { // want `function second has the same body as first \(custom\.go:3\)`
	y := x + 1
	return y
}

// Not reported: a single statement is below the configured minimum.
func oneA(x int) int { return x + 1 }

func oneB(x int) int { return x + 1 }
