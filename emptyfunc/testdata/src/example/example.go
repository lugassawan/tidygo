package example

type T struct{}

func (T) Empty() {} // want `function "Empty" has an empty body; add a comment explaining why`

func Free() {} // want `function "Free" has an empty body; add a comment explaining why`

// want +1 `function "MultiLine" has an empty body; add a comment explaining why`
func MultiLine() {
}

// OutsideAbove has a comment, but it sits outside the braces.
func OutsideAbove() {} // want `function "OutsideAbove" has an empty body; add a comment explaining why`

func Literal() {
	_ = func() {} // want `function has an empty body; add a comment explaining why`

	_ = func() {
		// intentionally empty
	}

	go func() {}() // want `function has an empty body; add a comment explaining why`
}

func (T) Commented() {
	// intentionally empty: satisfies an interface
}

func InlineBlock() { /* nothing to do */ }

func TrailingAfterBrace() { // intentionally empty
}

func WithStatements() int {
	return 1
}

func NestedEmpty() {
	if true {
	}
}
