package example

import stdos "os"

func aliasImport(p string) {
	if _, err := stdos.Stat(p); stdos.IsNotExist(err) {
		_, _ = stdos.Create(
			p,
		) // want `os\.Create after os\.Stat existence check is racy; use os\.OpenFile with O_CREATE\|O_EXCL`
	}
}
