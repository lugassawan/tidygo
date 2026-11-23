//go:build go1.1

package example

// Passing: files with a build constraint are exempt on both sides.

func (u User) TaggedMethod() {}

type Tagged struct{}
