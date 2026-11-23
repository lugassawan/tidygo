package methodcolocation

import (
	"go/parser"
	"go/token"
	"testing"
)

const (
	userFile    = "user.go"
	userSQLFile = "user_sql.go"
)

// analysistest only loads the testdata files that match the host's GOOS/GOARCH
// and build tags, so these tables pin the exemption logic on every host.

func TestHasPlatformSuffix(t *testing.T) {
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"GOOS suffix", "user_linux.go", true},
		{"GOOS suffix of another platform", "user_windows.go", true},
		{"GOARCH suffix", "user_amd64.go", true},
		{"GOARCH suffix with digits", "user_386.go", true},
		{"GOOS and GOARCH suffix", "user_linux_amd64.go", true},
		{"plain name", userFile, false},
		{"non-platform suffix", userSQLFile, false},
		{"platform name without a prefix", "linux.go", false},
		{"platform word inside the stem", "linuxuser.go", false},
		{"platform word before another suffix", "user_linux_sql.go", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasPlatformSuffix(c.file); got != c.want {
				t.Errorf("hasPlatformSuffix(%q) = %v, want %v", c.file, got, c.want)
			}
		})
	}
}

func TestIsSuffixedSibling(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		typeFile string
		want     bool
	}{
		{"stem followed by an underscore suffix", userSQLFile, userFile, true},
		{"a longer suffix", "user_json_codec.go", userFile, true},
		{"stem that merely starts the same", "username.go", userFile, false},
		{"plural stem", "users_x.go", userFile, false},
		{"unrelated file", "other.go", userFile, false},
		{"the type's own file", userFile, userFile, false},
		{"stem taken from the declaring file", "account_sql.go", "account.go", true},
		{"type name is not the stem", userSQLFile, "account.go", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isSuffixedSibling(c.file, c.typeFile); got != c.want {
				t.Errorf("isSuffixedSibling(%q, %q) = %v, want %v", c.file, c.typeFile, got, c.want)
			}
		})
	}
}

func TestHasBuildConstraint(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"go:build before the package clause", "//go:build linux\n\npackage p\n", true},
		{"go:build after a license header", "// Copyright.\n\n//go:build !windows\n\npackage p\n", true},
		{"go:build for a tag that never matches", "//go:build never_set_tag\n\npackage p\n", true},
		{"no constraint", "package p\n", false},
		{"ordinary comment", "// Package p does things.\npackage p\n", false},
		{"go:build text after the package clause", "package p\n\n//go:build linux\n", false},
		{"legacy plus-build line only", "// +build linux\n\npackage p\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "p.go", c.src, parser.ParseComments)
			if err != nil {
				t.Fatalf("parsing fixture: %v", err)
			}
			if got := hasBuildConstraint(file); got != c.want {
				t.Errorf("hasBuildConstraint(%q) = %v, want %v", c.src, got, c.want)
			}
		})
	}
}
