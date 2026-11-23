package example

// Named struct types are depth 0, so one anonymous level inside is fine.
type Named struct{ A int }

type OneLevel struct {
	Inner struct{ B int }
}

// An alias of an anonymous struct counts like a named type.
type Alias = struct {
	Inner struct{ B int }
}

// Table-driven shape: a single anonymous struct is allowed.
var table = []struct{ name string }{{"a"}}

var single struct{ A int }

// Two levels in a var are depth 2, reported once at the inner struct.
var twoLevels struct {
	DB struct { // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
		Host string
	}
}

// Three levels are reported once, at level 2. Level 3 is not reported again.
var threeLevels struct {
	A struct { // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
		B struct {
			C int
		}
	}
}

// Nesting through slices, maps, pointers and channels is still nesting.
var viaSlice struct {
	Items []struct{ ID int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}

var viaMap struct {
	ByName map[string]struct{ ID int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}

var viaPointer struct {
	Ptr *struct{ ID int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}

var viaChan struct {
	Ch chan struct{ ID int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}

// A field of a named type may hold one anonymous level, not two.
type Config struct {
	Server struct {
		TLS struct { // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
			On bool
		}
	}
}

// Parameter and result types count too.
func param(p struct {
	A struct{ B int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}) {
}

func result() (r struct {
	A struct{ B int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}) {
	return
}

// Local variables and composite literals behave like package-level ones.
func local() {
	var v struct {
		A struct{ B int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
	}
	_ = v

	cases := []struct {
		name string
		want struct{ ok bool } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
	}{}
	_ = cases
}

// A named slice, map or pointer of anonymous structs counts the element struct
// as the first level, unlike a named struct, which counts as depth 0.
type Rows []struct {
	Addr struct{ City string } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
}

// A local type declaration stops the walk like a package-level one.
func localDecl() {
	type L struct {
		A struct {
			B struct{ C int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
		}
	}
	_ = L{}
}

// A function type starts a new scope, so its parameters do not inherit the outer depth.
var fv struct {
	F func(p struct{ A int })
}

// So does a generic instantiation.
type Box[T any] struct{}

var inBox struct {
	F Box[struct{ A int }]
}

var deepInBox struct {
	F Box[struct {
		A struct{ B int } // want `anonymous struct nested 2 levels deep, max is 1; extract a named type`
	}]
}

// An interface method signature is a different scope: the walk stops there.
type Iface interface {
	Do(p struct{ A int }) struct{ B int }
}
