package example

import "io"

// Passing: the interface is an agent noun of its only method.

type Reader interface{ Read() }

type Closer interface{ Close() error }

type Writer interface{ Write() }

type Stopper interface{ Stop() }

type Runner interface{ Run() }

type Copier interface{ Copy() }

type Validator interface{ Validate() error }

type Visitor interface{ Visit() }

type Stringer interface{ String() string }

type Marshaler interface{ Marshal() ([]byte, error) }

// ReaderFrom uses the standard library's "first word + -er + rest" form.
type ReaderFrom interface{ ReadFrom() }

// The interface name's first-letter case is ignored.
type reader interface{ Read() }

type Getter[T any] interface{ Get() T }

func localPassing() {
	type flusher interface{ Flush() }
}

// Skipped: not exactly one named method.

type Multi interface {
	Read()
	Write()
}

type Empty interface{}

type EmbeddedOnly interface{ Reader }

type EmbeddedAndMethod interface {
	Reader
	Extra()
}

type EmbedsStdlib interface{ io.ReaderFrom }

type Number interface{ ~int | ~string }

type ConstraintAndMethod interface {
	~int | ~string
	Describe()
}

type _ interface{ Anything() }

// Sealed and marker interfaces have an unexported sole method and are skipped.

type Node interface{ isNode() }

type Option interface{ apply() }

// Reported.

type FileRead interface{ Read() } // want `single-method interface FileRead should be named after its method Read \(e\.g\. Reader\)`

type Handler interface{ ServeHTTP() } // want `single-method interface Handler should be named after its method ServeHTTP \(e\.g\. ServerHTTP\)`

type closeable interface{ Close() } // want `single-method interface closeable should be named after its method Close \(e\.g\. closer\)`

// The first word gets the suffix and the rest is kept, so UserGetter does not match GetUser.
type UserGetter interface{ GetUser() } // want `single-method interface UserGetter should be named after its method GetUser \(e\.g\. GetterUser\)`

type Aliased = interface{ Read() } // want `single-method interface Aliased should be named after its method Read \(e\.g\. Reader\)`

type Box[T any] interface{ Get() T } // want `single-method interface Box should be named after its method Get \(e\.g\. Getter\)`

func localReported() {
	type badLocal interface{ Flush() } // want `single-method interface badLocal should be named after its method Flush \(e\.g\. flusher\)`
}
