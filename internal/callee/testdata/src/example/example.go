package example

import (
	"bytes"
	"context"
	ctxalias "context"
	"os"
	"strings"
	"time"
)

type local struct{}

func (local) WithCancel() {}

func fn() {}

func funcs() {
	_, c1 := context.WithCancel(context.Background()) // want `func: context.WithCancel`
	c1()
	_, c2 := ctxalias.WithTimeout(context.Background(), time.Second) // want `func: context.WithTimeout`
	c2()

	// None of the calls below are context.WithCancel/WithTimeout/WithDeadline.
	_ = os.Getpid()
	local{}.WithCancel()
	fn()
	f := fn
	f()
	_ = context.Background().Err()
}

func methods(b *bytes.Buffer, v bytes.Buffer, r *strings.Reader) {
	b.Reset() // want `method: bytes.Buffer.Reset`
	v.Reset() // want `method: bytes.Buffer.Reset`

	// Same method name on a different type, and a different method on the right type.
	r.Reset("")
	b.Len()
}

// A local named like the package shadows the import, so its method is no match.
func shadowedPackage() {
	context := local{}
	context.WithCancel()
}
