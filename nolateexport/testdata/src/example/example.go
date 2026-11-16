package example

type T struct{}

func init() {} // OK - init is exempt

func Exported() {} // OK - exported before unexported

func unexported() {} // OK - unexported after exported

// OK - no unexported method of T precedes it; standalone functions don't count against methods.
func (t T) ExportedMethod() {}

func AlsoExported() {} // want `exported function AlsoExported should appear before unexported functions`

func init() {} // OK - second init is also exempt

func alsoUnexported() {} // OK

func (t T) unexportedMethod() {} // OK

func (t T) LateMethod() {} // want `exported method T.LateMethod should appear before unexported methods of T`

// Two types interleaved: each type is tracked on its own, so nothing here is reported.

type P struct{}

type Q struct{}

func (p P) PExported() {} // OK

func (p P) pHelper() {} // OK

func (q Q) QExported() {} // OK - Q has no unexported method before it

func (q Q) qHelper() {} // OK

// Pointer and value receivers share the same base type.

type Ptr struct{}

func (p *Ptr) helper() {} // OK

func (p Ptr) Value() {} // want `exported method Ptr.Value should appear before unexported methods of Ptr`

// Generic receivers are unwrapped to the base type.

type One[K comparable] struct{}

func (o *One[K]) helper() {} // OK

func (o *One[K]) Get() {} // want `exported method One.Get should appear before unexported methods of One`

type Two[K comparable, V any] struct{}

func (t Two[K, V]) helper() {} // OK

func (t Two[K, V]) Get() {} // want `exported method Two.Get should appear before unexported methods of Two`

// Non-struct receiver types behave the same way.

type Names []string

func (n Names) helper() {} // OK

func (n Names) Len() int { return len(n) } // want `exported method Names.Len should appear before unexported methods of Names`
