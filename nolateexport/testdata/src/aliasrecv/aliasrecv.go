package aliasrecv

// Receivers are keyed by the type name as written, so methods declared through
// an alias are tracked apart from methods declared on the aliased type. This
// fixture records that limitation: mixed ordering across the two is not reported.

type T struct{}

type A = T

func (a A) helper() {} // OK - keyed as "A"

func (t T) Pub() {} // OK - keyed as "T", which has no unexported method before it
