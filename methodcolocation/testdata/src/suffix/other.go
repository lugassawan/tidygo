package suffix

// Failing: an unrelated file stays reported even with AllowFileSuffix.

func (u User) Other() string { return u.name } // want `method User\.Other should be declared in user\.go, the file that declares User`
