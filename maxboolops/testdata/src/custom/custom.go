package custom

// Not reported: exactly 1 operator.
func atLimit(a, b bool) bool {
	return a && b
}

// Reported: 2 operators exceeds the configured max of 1.
func overLimit(a, b, c bool) bool {
	return a && b || c // want `boolean expression has 2 && / \|\| operators, max is 1`
}

// Reported: a nested chain is its own expression, counted against the configured max.
func nested(f func(bool) bool, a, b, c bool) bool {
	return f(a && b && c) // want `boolean expression has 2 && / \|\| operators, max is 1`
}
