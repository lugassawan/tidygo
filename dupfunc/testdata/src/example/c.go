package example

// Reported: a third copy is reported against the first function, not against sumAll.
func sumEvery(items []string) int { // want `function sumEvery has the same body as total \(a\.go:3\)`
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}
