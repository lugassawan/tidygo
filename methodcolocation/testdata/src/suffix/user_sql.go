package suffix

// Passing: user_sql.go is a <stem>_<suffix>.go sibling of user.go.

func (u User) Scan() error { return nil }

// Failing: the sibling relationship belongs to the type's file, so user_ does not cover Order.

func (o Order) Cost() int { return o.total } // want `method Order\.Cost should be declared in order\.go, the file that declares Order`
