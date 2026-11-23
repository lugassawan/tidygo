package includevars

const pkg_const = 1 // want `Rename constant "pkg_const" to match the regular expression \^`

var pkg_var = 2 // want `Rename variable "pkg_var" to match the regular expression \^`

var (
	grouped_a, groupedOK int // want `Rename variable "grouped_a" to match the regular expression \^`
	_                    int
)

// T is a dummy type for receiver tests.
type T struct{}

func (re_cv T) Method() {} // want `Rename receiver "re_cv" to match the regular expression \^`

func params(p_a int, pOK int) {} // want `Rename parameter "p_a" to match the regular expression \^`

func results() (r_a int, rOK int) { // want `Rename result "r_a" to match the regular expression \^`
	return
}

func locals(ch chan int, m map[string]int, v any) {
	const loc_const = 1 // want `Rename constant "loc_const" to match the regular expression \^`
	var loc_var int     // want `Rename variable "loc_var" to match the regular expression \^`
	short_var := 1      // want `Rename variable "short_var" to match the regular expression \^`
	_, _, _ = loc_const, loc_var, short_var

	for key_k, val_v := range m { // want `Rename variable "key_k" to match the regular expression \^` `Rename variable "val_v" to match the regular expression \^`
		_, _ = key_k, val_v
	}

	switch ty_v := v.(type) { // want `Rename variable "ty_v" to match the regular expression \^`
	case int:
		_ = ty_v
	}

	select {
	case rv_x := <-ch: // want `Rename variable "rv_x" to match the regular expression \^`
		_ = rv_x
	}

	fn := func(lit_p int) (lit_r int) { // want `Rename parameter "lit_p" to match the regular expression \^` `Rename result "lit_r" to match the regular expression \^`
		return
	}
	_ = fn
}

// A := that redeclares a name checks it again: without type information the
// analyzer cannot tell a redeclaration from a new variable.
func redeclared() {
	a_b, err := 1, 2 // want `variable "a_b"`
	a_b, ok := 3, 4  // want `variable "a_b"`
	_, _, _ = a_b, err, ok
}

func inits() {
	if if_v := 1; if_v > 0 { // want `variable "if_v"`
	}
	switch sw_v := 2; sw_v { // want `variable "sw_v"`
	}
	for for_i := 0; for_i < 1; for_i++ { // want `variable "for_i"`
	}
}

// Box is a dummy generic type for receiver tests.
type Box[T any] struct{}

func (re_g Box[T]) M() {} // want `receiver "re_g"`

// Receiver type parameters are type parameters, so they are not checked.
func (b Box[T_p]) N() {}

// Sites that are deliberately not checked.

type S struct{ field_name int }

type I interface {
	Do(p_x int) (r_x int)
}

var fnVar func(p_y int) (r_y int)

func generic[T_p any](x T_p) {}

func notDeclarations(m map[int]int) {
	existing := 1
	existing = 2 // assignment, not a declaration
	_ = existing
	for existing = range m { // range with = reuses an existing variable
	}
	var _ int
	_ = 3
}
