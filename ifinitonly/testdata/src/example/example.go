package example

import "errors"

var errTooMany = errors.New("too many")

type T struct{ n int }

func use(...any) {}

func f() int { return 1 }

func g() error { return nil }

// ---- reported ----

func basic(items []int) error {
	if n := len(items); n > 10 { // want `variable "n" in if init is only used in the condition`
		return errTooMany
	}
	return nil
}

func emptyBody() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
	}
}

func bodyRedeclares() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		n := 5
		use(n)
	}
}

func bodyParamShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		h := func(n int) int { return n }
		use(h)
	}
}

func bodyNestedIf() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		if n := 3; n > 1 {
			use(n)
		}
	}
}

func bodyRange(xs []int) {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		for _, n := range xs {
			use(n)
		}
	}
}

func bodySelector(x T) {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		use(x.n)
	}
}

func bodyLabel() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
	n:
		for {
			break n
		}
	}
}

func bodyTypeSwitch(v any) {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		switch n := v.(type) {
		case int:
			use(n)
		}
	}
}

func bodySelect(ch chan int) {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		select {
		case n := <-ch:
			use(n)
		}
	}
}

// ---- not reported ----

func errIdiom() error {
	if err := g(); err != nil {
		return err
	}
	return nil
}

func usedInElse() {
	if n := f(); n > 0 {
	} else {
		use(n)
	}
}

func usedInElseIfCond() {
	if n := f(); n > 0 {
	} else if n < -5 {
	}
}

func usedTwiceInCond() {
	if n := f(); n > 0 && n < 10 {
	}
}

func bodyRHSUsesOuter() {
	if n := f(); n > 0 {
		n := n + 1
		use(n)
	}
}

func bodyVarUsesOuter() {
	if n := f(); n > 0 {
		var m = n
		use(m)
	}
}

func bodyClosureCaptures() {
	if n := f(); n > 0 {
		func() { use(n) }()
	}
}

func multiValueInit(m map[int]int) {
	if v, ok := m[1]; ok && v > 0 {
	}
}

func plainAssignInit() {
	var n int
	if n = f(); n > 0 {
	}
}

func structLiteralKey() {
	if n := f(); n > 0 {
		use(T{n: 1})
	}
}

func typeSwitchInitUsesOuter(v any) {
	if n := f(); n > 0 {
		switch m := n; n := v.(type) {
		case int:
			use(m, n)
		}
	}
}

func addressTaken() {
	if n := f(); take(&n) {
	}
}

func useBeforeRedeclare() {
	if n := f(); n > 0 {
		use(n)
		n := 2
		use(n)
	}
}

func forInitUsesOuter() {
	if n := f(); n > 0 {
		for i := n; i < 3; i++ {
			use(i)
		}
	}
}

func switchTagUsesOuter() {
	if n := f(); n > 0 {
		switch n {
		case 1:
		}
	}
}

func take(p *int) bool { return p != nil }

// ---- reported: more shadowing forms ----

func bodyVarShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		var n = 5
		use(n)
	}
}

func bodyTypeShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		type n int
		use(n(1))
	}
}

func elseShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
	} else {
		n := 1
		use(n)
	}
}

func elseIfRedeclares() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
	} else if n := 3; n > 1 {
		use(n)
	}
}

func namedResultShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		use(func() (n int) { return n }())
	}
}

func forInitShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		for n := 0; n < 3; n++ {
			use(n)
		}
	}
}

func switchInitShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		switch n := 2; n {
		case 2:
		}
	}
}

func bodyConstShadows() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		const n = 1
		use(n)
	}
}

func structFieldNameIsNotUse() {
	if n := f(); n > 0 { // want `variable "n" in if init is only used in the condition`
		var s struct{ n int }
		use(s)
	}
}

func shortCircuitLeft(flag bool) {
	if n := f(); n > 0 && flag { // want `variable "n" in if init is only used in the condition`
	}
}

// ---- not reported: more uses of the outer variable ----

func caseBodyUses(x int) {
	if n := f(); n > 0 {
		switch x {
		case 1:
			use(n)
		}
	}
}

func bodyVarInitUsesOuter2() {
	if n := f(); n > 0 {
		var n = n
		use(n)
	}
}

func groupedSpecUsesOuter() {
	if n := f(); n > 0 {
		var (
			a = n
			n = 5
		)
		use(a, n)
	}
}

func rangeAssignUsesOuter(xs []int) {
	if n := f(); n > 0 {
		for n = range xs {
		}
	}
}

// Inlining would make the call conditional (right operand of && or ||).
func shortCircuitRight(flag bool) {
	if n := f(); flag && n > 0 {
	}
}

// Inlining would run the call on every closure invocation.
func closureInCond() {
	if n := f(); func() bool { return n > 0 }() {
	}
}
