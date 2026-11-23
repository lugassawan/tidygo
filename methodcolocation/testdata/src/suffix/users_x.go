package suffix

// Failing: users_x.go starts with users_, not user_.

func (u User) Plural() string { return u.name + "s" } // want `method User\.Plural should be declared in user\.go, the file that declares User`
