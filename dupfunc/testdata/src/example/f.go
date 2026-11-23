package example

type box[T any] struct{ v []T }

type pair[K comparable, V any] struct {
	k []K
	v []V
}

// Reported: a generic receiver is named by its base type, here through an IndexExpr.
func (x *box[T]) count() int { // want `function box\.count has the same body as left\.count \(d\.go:5\)`
	n := 0
	for range x.v {
		n++
	}
	return n
}

// Reported: a receiver with several type parameters is an IndexListExpr.
func (x *pair[K, V]) count() int { // want `function pair\.count has the same body as left\.count \(d\.go:5\)`
	n := 0
	for range x.v {
		n++
	}
	return n
}
