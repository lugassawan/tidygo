package example

// Reported: identical body in another file; comments and layout do not matter.
func sumAll(items []string) int { // want `function sumAll has the same body as total \(a\.go:3\)`
	sum := 0 // running total
	for _, s := range items {
		// add each length
		sum += len(s)
	}
	return sum
}
