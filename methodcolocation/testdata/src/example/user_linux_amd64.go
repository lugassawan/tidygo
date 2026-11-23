package example

// Passing: a GOOS and/or GOARCH file name suffix exempts the file. Only the files
// matching the host are compiled, so each declares a distinct method.

func (u User) OnLinuxAMD64() {}
