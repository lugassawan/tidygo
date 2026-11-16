package splitfile

type T struct{}

// State seen in this file must not leak into b.go.

func (t T) aHelper() {} // OK

func aFunc() {} // OK
