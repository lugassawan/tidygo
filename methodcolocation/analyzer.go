// Package methodcolocation defines an analyzer that requires a method to be
// declared in the same file as its receiver type. Test, generated, //go:build
// and GOOS/GOARCH-suffixed files are exempt on either side, as are aliases;
// Config.AllowFileSuffix also accepts <stem>_<suffix>.go for a type in <stem>.go.
package methodcolocation

import (
	"go/ast"
	"go/build/constraint"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Known GOOS and GOARCH values: a file name ending in one of them (or a GOOS_GOARCH pair) is
// built for that platform only.
var knownGOOS = []string{
	"aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "js", "linux",
	"nacl", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "zos",
}

var knownGOARCH = []string{
	"386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "loong64", "mips", "mipsle",
	"mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "ppc64", "ppc64le", "riscv", "riscv64",
	"s390", "s390x", "sparc", "sparc64", "wasm",
}

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// AllowFileSuffix permits methods in <stem>_<anything>.go when the receiver
	// type is declared in <stem>.go.
	AllowFileSuffix bool
}

// Analyzer reports methods declared in a different file than their receiver
// type, using the default (zero) Config.
var Analyzer = New(Config{})

// New builds a methodcolocation analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "methodcolocation",
		Doc:  "requires a method to be declared in the same file as its receiver type",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

// typeDecl records the file that declares a named type.
type typeDecl struct {
	file   string // base name of the declaring file
	exempt bool   // the declaring file is exempt from the check
}

func run(pass *analysis.Pass, cfg Config) (any, error) {
	decls := declaredTypes(pass)
	for _, file := range pass.Files {
		name := fileName(pass, file)
		if isExemptFile(file, name) {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil && len(fn.Recv.List) > 0 {
				checkMethod(pass, cfg, decls, name, fn)
			}
		}
	}
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// declaredTypes maps every file-scope type name of the package to its declaring file.
func declaredTypes(pass *analysis.Pass) map[string]typeDecl {
	decls := map[string]typeDecl{}
	for _, file := range pass.Files {
		name := fileName(pass, file)
		decl := typeDecl{file: name, exempt: isExemptFile(file, name)}
		for _, spec := range typeSpecs(file) {
			decls[spec.Name.Name] = decl
		}
	}
	return decls
}

// typeSpecs returns the file-scope type declarations of file, skipping aliases:
// a method on an alias belongs to the aliased type, which cannot be resolved
// without type information.
func typeSpecs(file *ast.File) []*ast.TypeSpec {
	var specs []*ast.TypeSpec
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok && !ts.Assign.IsValid() {
				specs = append(specs, ts)
			}
		}
	}
	return specs
}

func checkMethod(pass *analysis.Pass, cfg Config, decls map[string]typeDecl, file string, fn *ast.FuncDecl) {
	typ := recvTypeName(fn.Recv.List[0].Type)
	decl, ok := decls[typ]
	if !ok || decl.exempt || decl.file == file {
		return
	}
	if cfg.AllowFileSuffix && isSuffixedSibling(file, decl.file) {
		return
	}
	pass.Reportf(fn.Name.Pos(), "method %s.%s should be declared in %s, the file that declares %s",
		typ, fn.Name.Name, decl.file, typ)
}

// fileName returns the base name of file, ignoring //line directives.
func fileName(pass *analysis.Pass, file *ast.File) string {
	return filepath.Base(pass.Fset.PositionFor(file.Package, false).Filename)
}

// isExemptFile reports whether a file takes no part in the check: a test,
// generated, build-constrained or platform-specific file.
func isExemptFile(file *ast.File, name string) bool {
	return strings.HasSuffix(name, "_test.go") ||
		ast.IsGenerated(file) ||
		hasBuildConstraint(file) ||
		hasPlatformSuffix(name)
}

// hasBuildConstraint reports whether a //go:build line precedes the package clause.
func hasBuildConstraint(file *ast.File) bool {
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			return false
		}
		for _, comment := range group.List {
			if constraint.IsGoBuild(comment.Text) {
				return true
			}
		}
	}
	return false
}

// hasPlatformSuffix reports whether name ends in _<GOOS>, _<GOARCH> or
// _<GOOS>_<GOARCH> before .go. The last element decides all three forms.
func hasPlatformSuffix(name string) bool {
	stem := strings.TrimSuffix(name, ".go")
	i := strings.LastIndex(stem, "_")
	if i < 0 {
		return false
	}
	suffix := stem[i+1:]
	return slices.Contains(knownGOOS, suffix) || slices.Contains(knownGOARCH, suffix)
}

// isSuffixedSibling reports whether file is <stem>_<anything>.go for a type
// declared in <stem>.go.
func isSuffixedSibling(file, typeFile string) bool {
	return strings.HasPrefix(file, strings.TrimSuffix(typeFile, ".go")+"_")
}

// recvTypeName returns the base type name of a receiver expression, unwrapping
// pointers, parentheses and generic instantiations (T, *T, T[K], *T[K, V]).
func recvTypeName(expr ast.Expr) string {
	for {
		switch e := expr.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.StarExpr:
			expr = e.X
		case *ast.ParenExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		default:
			return ""
		}
	}
}
