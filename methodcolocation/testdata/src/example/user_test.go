package example

// Passing: methods in test files are exempt.

func (u User) Fixture() User { return User{name: "fixture"} }
