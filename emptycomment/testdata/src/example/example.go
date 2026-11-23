package example

// Documented has an empty "//" line between its paragraphs, which is how doc
// comments separate paragraphs and must not be reported.
//
// This is the second paragraph.
func Documented() int {
	/**/ // want `empty block comment; remove it`
	a := 1

	/* */ // want `empty block comment; remove it`
	b := 2

	/*	 */ // want `empty block comment; remove it`
	c := 3

	/***/ // want `empty block comment; remove it`
	d := 4

	// want +1 `empty block comment; remove it`
	/*
	 *
	 */
	e := 5

	/* text */
	f := 6

	/*
	 * text
	 */
	g := 7

	//
	h := 8

	return a + b + c + d + e + f + g + h
}
