package example

// Failing with the default config: only Config.AllowFileSuffix permits user_sql.go.

func (u User) Scan() error { return nil } // want `method User\.Scan should be declared in user\.go, the file that declares User`
