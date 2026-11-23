package example

func use(int) {}

// Not reported: a plain if has no else-if chain.
func plain(a bool) {
	if a {
		use(1)
	}
}

// Not reported: if/else is not an else-if chain.
func ifElse(a bool) {
	if a {
		use(1)
	} else {
		use(2)
	}
}

// Reported: the chain ends without a final else.
func chain(a, b bool) {
	if a {
		use(1)
	} else if b { // want `if-else-if chain has no final else`
		use(2)
	}
}

// Reported once, at the last link: a longer chain still gives a single finding.
func longChain(a, b, c bool) {
	if a {
		use(1)
	} else if b {
		use(2)
	} else if c { // want `if-else-if chain has no final else`
		use(3)
	}
}

// Not reported: the chain ends in else.
func chainWithElse(a, b bool) {
	if a {
		use(1)
	} else if b {
		use(2)
	} else {
		use(3)
	}
}

// Not reported: a long chain that ends in else.
func longChainWithElse(a, b, c bool) {
	if a {
		use(1)
	} else if b {
		use(2)
	} else if c {
		use(3)
	} else {
		use(4)
	}
}

// Not reported: an if nested inside an else block is not an else-if link.
func nestedInElse(a, b bool) {
	if a {
		use(1)
	} else {
		if b {
			use(2)
		}
	}
}

// Reported: a chain inside an else block is its own chain.
func chainInElse(a, b, c bool) {
	if a {
		use(1)
	} else {
		if b {
			use(2)
		} else if c { // want `if-else-if chain has no final else`
			use(3)
		}
	}
}

// Reported: a chain nested in the body of another chain is checked separately.
func chainInBody(a, b, c, d bool) {
	if a {
		if b {
			use(1)
		} else if c { // want `if-else-if chain has no final else`
			use(2)
		}
	} else if d {
		use(3)
	} else {
		use(4)
	}
}

// Reported: an init statement does not change the chain.
func withInit(a bool) {
	if x := 1; a {
		use(x)
	} else if y := 2; y > x { // want `if-else-if chain has no final else`
		use(y)
	}
}
