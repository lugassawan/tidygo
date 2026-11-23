package example

import (
	"strings"
	"time"
)

// Matching bodies are only reported when the signatures are identical too:
// functions whose parameter or result types differ cannot be merged without generics.

// Not reported: the same body over int and float64 needs generics to merge.
func clampInt(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func clampFloat(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// Not reported: only the result type differs.
func resultInt(n int) int {
	y := n + 7
	sink(y)
	return y
}

func resultAny(n int) any {
	y := n + 7
	sink(y)
	return y
}

// Not reported: a variadic parameter is not a slice parameter.
func sumVariadic(v ...int) int {
	n := len(v)
	sink(n, 11)
	return n
}

func sumSlice(v []int) int {
	n := len(v)
	sink(n, 11)
	return n
}

// Not reported: the constraints of the type parameters differ.
func keepAny[T any](v T) T {
	sink(v, "keep")
	sink(v)
	return v
}

func keepComparable[T comparable](v T) T {
	sink(v, "keep")
	sink(v)
	return v
}

type duration time.Duration

// Not reported: a defined type is not its underlying type.
func waitStd(d time.Duration) time.Duration {
	sink(d, "wait")
	sink(d)
	return d
}

func waitLocal(d duration) duration {
	sink(d, "wait")
	sink(d)
	return d
}

// Reported: the same qualified type in both signatures.
func waitStdAgain(
	d time.Duration,
) time.Duration { // want `function waitStdAgain has the same body as waitStd \(sig\.go:74\)`
	sink(d, "wait")
	sink(d)
	return d
}

type stack[T any] struct{ v []T }

type queue[K comparable, V any] struct {
	v []K
	w []V
}

// Not reported: the receivers declare unrelated type parameters, so x has a different type in each.
func (s *stack[T]) push(x T) {
	s.v = append(s.v, x)
	sink(x)
	sink(s)
}

func (s *queue[K, V]) push(x K) {
	s.v = append(s.v, x)
	sink(x)
	sink(s)
}

// Reported: each method of a generic type names its own receiver type parameters, which must not matter.
func (s *stack[T]) pushA(x T) {
	s.v = append(s.v, x)
	sink(x, "a")
	sink(s)
}

func (s *stack[E]) pushB(x E) { // want `function stack\.pushB has the same body as stack\.pushA \(sig\.go:114\)`
	s.v = append(s.v, x)
	sink(x, "a")
	sink(s)
}

// Reported: byte is uint8.
func headBytes(b []byte) []byte {
	n := len(b) / 2
	sink(n, "head")
	return b[:n]
}

func headUint8(b []uint8) []uint8 { // want `function headUint8 has the same body as headBytes \(sig\.go:127\)`
	n := len(b) / 2
	sink(n, "head")
	return b[:n]
}

type ints = []int

// Reported: an alias is its target type.
func lastAlias(v ints) int {
	n := len(v) - 1
	sink(n, "last")
	return v[n]
}

func lastSlice(v []int) int { // want `function lastSlice has the same body as lastAlias \(sig\.go:142\)`
	n := len(v) - 1
	sink(n, "last")
	return v[n]
}

// Reported: generic functions match up to the names of their type parameters.
func firstA[T any](v []T) T {
	x := v[0]
	sink(x, "first")
	return x
}

func firstB[U any](v []U) U { // want `function firstB has the same body as firstA \(sig\.go:155\)`
	x := v[0]
	sink(x, "first")
	return x
}

type user struct{ name string }

// Reported: methods with the same signature, as in the Label/Title pair.
func (u user) Label() string {
	s := strings.TrimSpace(u.name)
	sink(s)
	return s
}

func (u user) Title() string { // want `function user\.Title has the same body as user\.Label \(sig\.go:170\)`
	s := strings.TrimSpace(u.name)
	sink(s)
	return s
}

// Reported: blank functions are recorded by the type checker like any other, and
// have the same (empty) signature.
func _() {
	sink(1, "blank")
	sink(2)
	sink(3)
}

func _() { // want `function _ has the same body as _ \(sig\.go:184\)`
	sink(1, "blank")
	sink(2)
	sink(3)
}

// Reported: a body can hold several signature classes. clampFloat opened the second
// one, so a later float64 copy is reported against it and not against clampInt, and
// an int copy after it still finds the first class.
func clampFloatAgain(
	x, lo, hi float64,
) float64 { // want `function clampFloatAgain has the same body as clampFloat \(sig\.go:22\)`
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func clampIntAgain(x, lo, hi int) int { // want `function clampIntAgain has the same body as clampInt \(sig\.go:12\)`
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// Reported: a value receiver and a pointer receiver on the same generic type share a signature.
func (s stack[T]) pushC(x T) {
	sink(x, "c")
	sink(s)
	sink(s.v)
}

func (s *stack[E]) pushD(x E) { // want `function stack\.pushD has the same body as stack\.pushC \(sig\.go:220\)`
	sink(x, "c")
	sink(s)
	sink(s.v)
}

// Reported: the receiver is not part of the signature, so pointer and value receivers match.
func (u *user) Name() string {
	s := strings.ToUpper(u.name)
	sink(s)
	return s
}

func (u user) Upper() string { // want `function user\.Upper has the same body as user\.Name \(sig\.go:233\)`
	s := strings.ToUpper(u.name)
	sink(s)
	return s
}
