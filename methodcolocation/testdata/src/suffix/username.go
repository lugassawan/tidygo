package suffix

// Failing: username.go shares the user prefix but has no underscore after the stem.

func (u User) Label() string { return u.name } // want `method User\.Label should be declared in user\.go, the file that declares User`
