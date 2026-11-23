// Package ifacename defines an analyzer that requires a single-method interface
// to be named after its method plus an agent-noun suffix (Read → Reader).
//
// Only interfaces declared in the linted packages whose body is exactly one
// exported method are checked; embedded interfaces, type-set elements and
// unexported (sealed or marker) methods are skipped. Config.AllowedNames exempts
// names that cannot be derived mechanically. See docs/analyzers/ifacename.md for the
// accepted spellings.
package ifacename

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// AllowedNames lists interface names that are exempt from the check.
	AllowedNames []string
}

// Analyzer reports single-method interfaces that are not named after their
// method, using the default (zero) Config.
var Analyzer = New(Config{})

// New builds an ifacename analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	allowed := make(map[string]bool, len(cfg.AllowedNames))
	for _, name := range cfg.AllowedNames {
		allowed[name] = true
	}
	return &analysis.Analyzer{
		Name:     "ifacename",
		Doc:      "requires single-method interfaces to be named after their method plus an agent-noun suffix",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, allowed)
		},
	}
}

func run(pass *analysis.Pass, allowed map[string]bool) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}
		checkSpec(pass, allowed, ts)
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func checkSpec(pass *analysis.Pass, allowed map[string]bool, ts *ast.TypeSpec) {
	iface, ok := ts.Type.(*ast.InterfaceType)
	if !ok || ts.Name.Name == "_" {
		return
	}
	method, ok := singleMethod(iface)
	if !ok || !ast.IsExported(method) {
		return // unexported sole method: a sealed or marker interface (isNode, apply)
	}
	name := ts.Name.Name
	if allowed[name] || isAcceptedName(method, name) {
		return
	}
	pass.Reportf(ts.Name.Pos(),
		"single-method interface %s should be named after its method %s (e.g. %s)",
		name, method, suggest(method, ast.IsExported(name)),
	)
}

// singleMethod returns the method name when the interface body is exactly one
// named method. Embedded interfaces and type-set elements are fields without
// names, so any of them (or a second method) yields false.
func singleMethod(iface *ast.InterfaceType) (string, bool) {
	if iface.Methods == nil || len(iface.Methods.List) != 1 {
		return "", false
	}
	field := iface.Methods.List[0]
	if len(field.Names) != 1 {
		return "", false
	}
	return field.Names[0].Name, true
}

// isAcceptedName reports whether name is an agent noun of method, ignoring the
// case of the first letter on both sides.
func isAcceptedName(method, name string) bool {
	want := upperFirst(name)
	for _, candidate := range acceptedNames(method) {
		if upperFirst(candidate) == want {
			return true
		}
	}
	return false
}

// acceptedNames derives every name accepted for method: agent nouns of the whole
// method name, plus agent nouns of its first camel-case word with the remainder
// appended unchanged (ReadFrom → ReaderFrom).
func acceptedNames(method string) []string {
	names := agentNouns(method)
	if word := firstWord(method); word != method {
		rest := method[len(word):]
		for _, noun := range agentNouns(word) {
			names = append(names, noun+rest)
		}
	}
	return names
}

// agentNouns returns the common English agent-noun spellings of word.
func agentNouns(word string) []string {
	r := []rune(word)
	n := len(r)
	if n == 0 {
		return nil
	}
	nouns := []string{word + "er", word + "or"}
	switch last := unicode.ToLower(r[n-1]); {
	case last == 'e':
		nouns = append(nouns, word+"r", string(r[:n-1])+"or")
	case last == 'y' && n > 1 && isConsonant(r[n-2]):
		nouns = append(nouns, string(r[:n-1])+"ier")
	case endsConsonantVowelConsonant(r):
		nouns = append(nouns, word+string(r[n-1])+"er")
	}
	return nouns
}

// suggest builds the example name shown in the diagnostic. It picks one spelling
// by the first rule that applies, then cases its first letter to match whether
// the offending interface is exported.
func suggest(method string, exported bool) string {
	word := firstWord(method)
	noun := primaryNoun(word) + method[len(word):]
	return caseFirst(noun, exported)
}

func primaryNoun(word string) string {
	r := []rune(word)
	n := len(r)
	if n == 0 {
		return word
	}
	switch last := unicode.ToLower(r[n-1]); {
	case isLatinateAte(word):
		return string(r[:n-1]) + "or"
	case last == 'e':
		return word + "r"
	case last == 'y' && n > 1 && isConsonant(r[n-2]):
		return string(r[:n-1]) + "ier"
	case endsConsonantVowelConsonant(r) && vowelGroups(r) == 1:
		return word + string(r[n-1]) + "er"
	default:
		return word + "er"
	}
}

// firstWord returns the first camel-case word of name, keeping a leading
// acronym together (HTTPServe → HTTP, ReadFrom → Read, ID → ID).
func firstWord(name string) string {
	r := []rune(name)
	if len(r) < 2 {
		return name
	}
	i := 1
	if unicode.IsUpper(r[0]) && unicode.IsUpper(r[1]) {
		for i < len(r) && unicode.IsUpper(r[i]) {
			i++
		}
		if i < len(r) {
			i-- // the last capital starts the next word
		}
		return string(r[:i])
	}
	for i < len(r) && !unicode.IsUpper(r[i]) {
		i++
	}
	return string(r[:i])
}

// isLatinateAte reports whether word is a verb like Validate or Create whose
// agent noun takes -or (Validator, Creator) rather than -r. Update is the
// common exception (Updater).
func isLatinateAte(word string) bool {
	lower := strings.ToLower(word)
	return len(lower) > 4 && strings.HasSuffix(lower, "ate") && !strings.HasSuffix(lower, "pdate")
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

func isConsonant(r rune) bool {
	return unicode.IsLetter(r) && !isVowel(r)
}

// endsConsonantVowelConsonant reports whether r ends in a consonant-vowel-consonant
// triple whose final letter is not w, x or y (which do not double: Show → Shower).
func endsConsonantVowelConsonant(r []rune) bool {
	n := len(r)
	if n < 3 {
		return false
	}
	switch unicode.ToLower(r[n-1]) {
	case 'w', 'x', 'y':
		return false
	}
	return isConsonant(r[n-3]) && isVowel(r[n-2]) && isConsonant(r[n-1])
}

// vowelGroups counts runs of consecutive vowels, a rough syllable count.
func vowelGroups(r []rune) int {
	groups := 0
	inGroup := false
	for _, c := range r {
		if isVowel(c) {
			if !inGroup {
				groups++
			}
			inGroup = true
		} else {
			inGroup = false
		}
	}
	return groups
}

func upperFirst(s string) string {
	return caseFirst(s, true)
}

func caseFirst(s string, upper bool) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	if upper {
		r[0] = unicode.ToUpper(r[0])
	} else {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}
