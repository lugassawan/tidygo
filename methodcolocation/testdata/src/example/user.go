package example

// Passing: every method lives in the file that declares its receiver type.

type User struct{ name string }

func (u User) Name() string { return u.name }

func (u *User) SetName(name string) { u.name = name }

type Team struct{ members []User }

func (t Team) Size() int { return len(t.members) }

// Box is generic, so its receivers are written Box[T].
type Box[T any] struct{ v T }

func (b *Box[T]) Get() T { return b.v }

type Pair[K comparable, V any] struct {
	k K
	v V
}

func (p Pair[K, V]) Key() K { return p.k }
