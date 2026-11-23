package custom

// With max 2, two anonymous levels are allowed.
var twoLevels struct {
	DB struct{ Host string }
}

// The third level is the first one over the limit, reported once.
var threeLevels struct {
	A struct {
		B struct { // want `anonymous struct nested 3 levels deep, max is 2; extract a named type`
			C struct{ D int }
		}
	}
}
