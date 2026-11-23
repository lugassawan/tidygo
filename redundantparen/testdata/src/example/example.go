package example

type item struct{ ok bool }

func use(...any) {}

// Reported: parentheses around the whole returned value.
func returned(a, b int) int {
	return (a + b) // want `redundant parentheses`
}

// Reported: each returned operand is its own top-level position.
func returnedMulti(a, b int) (int, int) {
	return (a), (b) // want `redundant parentheses` `redundant parentheses`
}

// Reported: every assignment form has a top-level right-hand side.
func assigned(a, b int) {
	x := (a + b) // want `redundant parentheses`
	x = (a)      // want `redundant parentheses`
	x += (b)     // want `redundant parentheses`
	var y = (a)  // want `redundant parentheses`
	var (
		z = (b) // want `redundant parentheses`
	)
	use(x, y, z)
}

// Reported: a call argument is a top-level position.
func callArgs(f func(int) int, a int) {
	use((a))         // want `redundant parentheses`
	_ = f((a + 1))   // want `redundant parentheses`
	_ = f(a + (1)*2) // Not reported: the parentheses sit inside a binary operand.
}

// Not reported: parentheses on the callee are needed for conversions and are never flagged.
func conversions(f func(int), p *int) {
	_ = (*int)(nil)
	(f)(1)
	_ = (func())(nil)
	use(p)
}

// Not reported: parentheses that change grouping or clarify mixed operators.
func needed(a, b, c int, p, q, r bool) {
	_ = (a + b) * c
	_ = (p && q) || r
	_ = !(p)
	var t (int)
	_ = t
	_ = item{ok: (p)}
}

// Not reported: without the parentheses a composite literal in a statement header does not parse.
func headers(p bool) {
	if x := (item{ok: p}); x.ok {
		use(x)
	}
	for i := (item{ok: p}); i.ok; {
		use(i)
		break
	}
	switch x := (item{ok: p}); x.ok {
	default:
		use(x)
	}
}

// Reported: a header without a composite literal has no parsing ambiguity.
func headerPlain(a int) {
	if x := (a); x > 0 { // want `redundant parentheses`
		use(x)
	}
}

// Reported: inside the body the composite literal is unambiguous again.
func inBody(p bool) {
	if p {
		x := (item{ok: p}) // want `redundant parentheses`
		use(x)
	}
}
