package example

// Failing: the receiver type is declared in another file.

func (u User) Display() string { return u.name } // want `method User\.Display should be declared in user\.go, the file that declares User`

func (u *User) Reset() { u.name = "" } // want `method User\.Reset should be declared in user\.go, the file that declares User`

func (t *Team) Add(u User) { // want `method Team\.Add should be declared in user\.go`
	t.members = append(t.members, u)
}

func (b Box[T]) Peek() T { return b.v } // want `method Box\.Peek should be declared in user\.go, the file that declares Box`

func (b *Box[K]) Set(v K) { // want `method Box\.Set should be declared in user\.go`
	b.v = v
}

func (p *Pair[K, V]) Value() V { return p.v } // want `method Pair\.Value should be declared in user\.go, the file that declares Pair`

// Passing: the type is declared in an exempt file (generated or build-constrained).

func (g Generated) Touch() {}

func (t Tagged) Touch() {}

// Passing: a method on an alias is not tied to the alias's own file.

func (a Alias) ViaAlias() {}
