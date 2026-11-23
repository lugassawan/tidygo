package example

type flags struct{ x, y, z, w bool }

// Not reported: exactly 3 operators.
func atLimit(a, b, c, d bool) bool {
	return a && b && c && d
}

// Reported: 4 operators exceeds the default max of 3.
func overLimit(a, b, c, d, e bool) bool {
	return a && b && c && d && e // want `boolean expression has 4 && / \|\| operators, max is 3`
}

// Reported: mixed && and || are counted together.
func mixed(a, b, c, d, e bool) bool {
	return a && b || c && d || e // want `boolean expression has 4 && / \|\| operators, max is 3`
}

// Reported: parenthesised groups belong to the same expression.
func parenthesised(a, b, c, d, e, f bool) bool {
	return (a && b) || (c && d) || (e && f) // want `boolean expression has 5 && / \|\| operators, max is 3`
}

// Reported once at the root: negation does not start a new expression.
func negated(a, b, c, d, e bool) bool {
	return !(a && b && c) || d && e // want `boolean expression has 4 && / \|\| operators, max is 3`
}

// Not reported: comparisons are leaves, only && and || are counted.
func comparisons(x, y, z, w int) bool {
	return x > 1 && y < 2 && z == 3 && w != 4
}

// Reported: an if condition is a boolean expression too.
func ifCondition(a, b, c, d, e bool) {
	if a && b && c && d && e { // want `boolean expression has 4 && / \|\| operators, max is 3`
		return
	}
}

// Reported: a for condition is a boolean expression too.
func forCondition(a, b, c, d, e bool) {
	for a && b && c && d && e { // want `boolean expression has 4 && / \|\| operators, max is 3`
		return
	}
}

// Reported: a call argument is its own expression.
func callArgument(f func(bool), a, b, c, d, e bool) {
	f(a && b && c && d && e) // want `boolean expression has 4 && / \|\| operators, max is 3`
}

// Not reported: the call argument is counted on its own, so neither side exceeds the limit.
func callInsideChain(g func(bool) bool, a, b, c, d, e bool) bool {
	return a && g(b && c && d) && e
}

// Reported: only the nested argument chain exceeds the limit.
func callInsideChainOver(g func(bool) bool, a, b, c, d, e, f bool) bool {
	return a && g(b && c && d && e && f) // want `boolean expression has 4 && / \|\| operators, max is 3`
}

// Reported: a closure body is its own expression; the outer chain has only 2 operators.
func closure(a, b bool) bool {
	return a && b && func() bool {
		return a && b && a && b && a // want `boolean expression has 4 && / \|\| operators, max is 3`
	}()
}

// Not reported: small expressions in separate struct fields do not add up.
func literal(a, b bool) flags {
	return flags{x: a && b, y: a || b, z: a && b, w: a || b}
}

// Not reported: no boolean operators at all.
func plain(a bool) bool {
	return a
}
