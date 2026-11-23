package example

import . "context"

// A dot import brings the function in without a qualifier; it still resolves.
func dotImport() {
	_, c := WithCancel(Background()) // want `func: context.WithCancel`
	c()
}
