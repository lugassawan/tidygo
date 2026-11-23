package example

func total(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

// Reported: identical body in the same file, reported against the first function.
func totalAgain(items []string) int { // want `function totalAgain has the same body as total \(a\.go:3\)`
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

// Not reported: a renamed local changes the body.
func totalRenamed(items []string) int {
	acc := 0
	for _, s := range items {
		acc += len(s)
	}
	return acc
}

// Not reported: a different literal changes the body.
func totalFromOne(items []string) int {
	sum := 1
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

// Not reported: a different operator changes the body.
func totalMinus(items []string) int {
	sum := 0
	for _, s := range items {
		sum -= len(s)
	}
	return sum
}

// Not reported: s[n:] and s[:n] are different slices even though both hold one operand.
func tailFrom(s []int) []int {
	n := 1
	t := s[n:]
	return t
}

func headTo(s []int) []int {
	n := 1
	t := s[:n]
	return t
}

// Not reported: bodies shorter than 3 statements are too small to matter.
func shortOne(x int) int {
	y := x + 1
	return y
}

func shortTwo(x int) int {
	y := x + 1
	return y
}

// Not reported: function literals have no name and are never compared.
var literalOne = func(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

var literalTwo = func(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}
