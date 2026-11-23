package example

import (
	"errors"
	"io/fs"
	"os"
)

const (
	confPath = "/etc/app.conf"
	perm     = 0o644
)

type config struct {
	Path string
}

func touch() {}

// ---- reported ----

func initForm(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.Create(
			p,
		) // want `os\.Create after os\.Stat existence check is racy; use os\.OpenFile with O_CREATE\|O_EXCL`
	}
}

func separateStatements(p string) {
	_, err := os.Stat(p)
	touch()
	if os.IsNotExist(err) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func errorsIsFS(p string) {
	if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func errorsIsOS(p string) {
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func lstat(p string) {
	if _, err := os.Lstat(p); os.IsNotExist(err) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func writeFile(p string, data []byte) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_ = os.WriteFile(
			p,
			data,
			perm,
		) // want `os\.WriteFile after os\.Stat existence check is racy; use os\.OpenFile with O_CREATE\|O_EXCL`
	}
}

func openFileCreate(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.OpenFile(
			p,
			os.O_CREATE|os.O_WRONLY,
			0o644,
		) // want `os\.OpenFile after os\.Stat existence check is racy; use os\.OpenFile with O_CREATE\|O_EXCL`
	}
}

func constantLiteralPath() {
	if _, err := os.Stat("/tmp/app.lock"); os.IsNotExist(err) {
		_, _ = os.Create("/tmp/app.lock") // want `os\.Create after os\.Stat existence check is racy`
	}
}

func constantNamedPath() {
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		_, _ = os.Create("/etc/app.conf") // want `os\.Create after os\.Stat existence check is racy`
	}
}

func selectorPath(cfg config) {
	if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
		_, _ = os.Create(cfg.Path) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func conjunctionTrailing(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) && len(p) > 0 {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func conjunctionLeading(p string, enabled bool) {
	if _, err := os.Stat(p); enabled && os.IsNotExist(err) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func nestedBlock(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if len(p) > 0 {
			_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
		}
	}
}

func elseIfBranch(p string) {
	_, err := os.Stat(p)
	if err == nil {
		return
	} else if os.IsNotExist(err) {
		_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
	}
}

func insideFuncLit(p string) func() {
	return func() {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			_, _ = os.Create(p) // want `os\.Create after os\.Stat existence check is racy`
		}
	}
}

// ---- not reported ----

func exclusiveOpen(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	}
}

func openFileWithoutCreate(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.OpenFile(p, os.O_WRONLY, perm)
	}
}

func nonConstantFlag(p string) {
	flag := os.O_CREATE | os.O_WRONLY
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.OpenFile(p, flag, perm)
	}
}

func differentPath(p, q string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		_, _ = os.Create(q)
	}
}

func differentConstants() {
	if _, err := os.Stat("/tmp/a"); os.IsNotExist(err) {
		_, _ = os.Create("/tmp/b")
	}
}

func pathReassigned(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		p += ".new"
		_, _ = os.Create(p)
	}
}

func selectorReassigned(cfg *config) {
	if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
		cfg.Path = "/tmp/other"
		_, _ = os.Create(cfg.Path)
	}
}

func elseBranch(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		touch()
	} else {
		_, _ = os.Create(p)
	}
}

func negatedCheck(p string) {
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		_, _ = os.Create(p)
	}
}

func disjunction(p string, force bool) {
	if _, err := os.Stat(p); force || os.IsNotExist(err) {
		_, _ = os.Create(p)
	}
}

func errorNotNil(p string) {
	if _, err := os.Stat(p); err != nil {
		_, _ = os.Create(p)
	}
}

func errorsIsOtherSentinel(p string) {
	if _, err := os.Stat(p); errors.Is(err, fs.ErrExist) {
		_, _ = os.Create(p)
	}
}

func createOutsideIf(p string) {
	_, _ = os.Create(p)
}

func createBeforeCheck(p string) {
	_, _ = os.Create(p)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		touch()
	}
}

func nestedFuncLit(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		defer func() {
			_, _ = os.Create(p)
		}()
	}
}

func errNotFromStat(p string) {
	err := os.Remove(p)
	if os.IsNotExist(err) {
		_, _ = os.Create(p)
	}
}

func errReassignedBeforeCheck(p string) {
	_, err := os.Stat(p)
	err = os.Remove(p)
	if os.IsNotExist(err) {
		_, _ = os.Create(p)
	}
}

func shadowedErr(p string) {
	_, err := os.Stat(p)
	_ = err
	{
		err := os.Remove(p)
		if os.IsNotExist(err) {
			_, _ = os.Create(p)
		}
	}
}
