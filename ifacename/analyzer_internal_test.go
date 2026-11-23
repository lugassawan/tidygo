package ifacename

import (
	"slices"
	"testing"
)

const (
	read      = "Read"
	readLower = "read"
	readFrom  = "ReadFrom"
	serveHTTP = "ServeHTTP"
	closeM    = "Close"
	copyM     = "Copy"
	stopM     = "Stop"
	visitM    = "Visit"
	validateM = "Validate"

	reader     = "Reader"
	closer     = "Closer"
	copier     = "Copier"
	stopper    = "Stopper"
	visitor    = "Visitor"
	validator  = "Validator"
	readerFrom = "ReaderFrom"
)

func TestFirstWord(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: read, want: read},
		{name: readFrom, want: read},
		{name: serveHTTP, want: "Serve"},
		{name: readLower, want: readLower},
		{name: "HTTPServe", want: "HTTP"},
		{name: "ID", want: "ID"},
		{name: "X", want: "X"},
		{name: "", want: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := firstWord(c.name); got != c.want {
				t.Errorf("firstWord(%q) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}

func TestAgentNouns(t *testing.T) {
	cases := []struct {
		word string
		want []string
	}{
		{word: read, want: []string{reader, "Reador"}},
		{word: closeM, want: []string{closer, "Closor"}},
		{word: validateM, want: []string{"Validater", validator}},
		{word: copyM, want: []string{copier}},
		{word: stopM, want: []string{stopper}},
		{word: visitM, want: []string{visitor, "Visitter"}},
		{word: "Run", want: []string{"Runner"}},
		{word: "Show", want: []string{"Shower"}},
		{word: "Play", want: []string{"Player"}},
		{word: "Fix", want: []string{"Fixer"}},
	}
	for _, c := range cases {
		t.Run(c.word, func(t *testing.T) {
			got := agentNouns(c.word)
			for _, w := range c.want {
				if !slices.Contains(got, w) {
					t.Errorf("agentNouns(%q) = %v, missing %q", c.word, got, w)
				}
			}
		})
	}

	t.Run("never accepts an unrelated name", func(t *testing.T) {
		if got := agentNouns(read); slices.Contains(got, "Handler") {
			t.Errorf("agentNouns(%q) = %v, unexpectedly contains Handler", read, got)
		}
	})
	t.Run("empty word yields nothing", func(t *testing.T) {
		if got := agentNouns(""); len(got) != 0 {
			t.Errorf("agentNouns(%q) = %v, want none", "", got)
		}
	})
}

func TestAgentNounsExcludesBadSpellings(t *testing.T) {
	excluded := map[string]string{
		"Seek": "Seekker", // vowel pair before the last letter does not double
		"Show": "Showwer", // a trailing w does not double
		"Fix":  "Fixxer",  // a trailing x does not double
		"Play": "Playier", // vowel + y keeps the y
	}
	for word, bad := range excluded {
		if got := agentNouns(word); slices.Contains(got, bad) {
			t.Errorf("agentNouns(%q) = %v, must not contain %q", word, got, bad)
		}
	}
}

func TestIsAcceptedName(t *testing.T) {
	cases := []struct {
		method string
		iface  string
		want   bool
	}{
		{method: read, iface: reader, want: true},
		{method: read, iface: "reader", want: true},
		{method: readLower, iface: reader, want: true},
		{method: closeM, iface: closer, want: true},
		{method: validateM, iface: validator, want: true},
		{method: copyM, iface: copier, want: true},
		{method: stopM, iface: stopper, want: true},
		{method: visitM, iface: visitor, want: true},
		{method: "String", iface: "Stringer", want: true},
		{method: "Marshal", iface: "Marshaler", want: true},
		{method: readFrom, iface: readerFrom, want: true},
		{method: "WriteTo", iface: "WriterTo", want: true},
		{method: "ReadAt", iface: "ReaderAt", want: true},
		{method: read, iface: "FileRead", want: false},
		{method: serveHTTP, iface: "Handler", want: false},
		{method: closeM, iface: "closeable", want: false},
		{method: readFrom, iface: "FromReader", want: false},
	}
	for _, c := range cases {
		t.Run(c.method+"/"+c.iface, func(t *testing.T) {
			if got := isAcceptedName(c.method, c.iface); got != c.want {
				t.Errorf("isAcceptedName(%q, %q) = %v, want %v", c.method, c.iface, got, c.want)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	cases := []struct {
		method   string
		exported bool
		want     string
	}{
		{method: read, exported: true, want: reader},
		{method: closeM, exported: true, want: closer},
		{method: copyM, exported: true, want: copier},
		{method: stopM, exported: true, want: stopper},
		{method: "Open", exported: true, want: "Opener"},
		{method: validateM, exported: true, want: validator},
		{method: "Create", exported: true, want: "Creator"},
		{method: "Update", exported: true, want: "Updater"},
		{method: "Date", exported: true, want: "Dater"},
		{method: readFrom, exported: true, want: readerFrom},
		{method: serveHTTP, exported: true, want: "ServerHTTP"},
		{method: closeM, exported: false, want: "closer"},
		{method: readLower, exported: true, want: reader},
		{method: readLower, exported: false, want: "reader"},
	}
	for _, c := range cases {
		t.Run(c.method, func(t *testing.T) {
			if got := suggest(c.method, c.exported); got != c.want {
				t.Errorf("suggest(%q, %v) = %q, want %q", c.method, c.exported, got, c.want)
			}
		})
	}
}
