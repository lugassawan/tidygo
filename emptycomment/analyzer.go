// Package emptycomment defines an analyzer that reports block comments holding
// nothing but whitespace and asterisks.
package emptycomment

import (
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports empty /* */ block comments.
var Analyzer = &analysis.Analyzer{
	Name: "emptycomment",
	Doc:  "reports block comments that contain only whitespace and asterisks",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, c := range group.List {
				// "//" comments are never reported: an empty "//" line
				// separates the paragraphs of a doc comment.
				if !strings.HasPrefix(c.Text, "/*") {
					continue
				}
				if isEmptyBlock(c.Text) {
					pass.Reportf(c.Pos(), "empty block comment; remove it")
				}
			}
		}
	}
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// isEmptyBlock reports whether the block comment text, delimiters included,
// has nothing left after dropping whitespace and "*" characters.
func isEmptyBlock(text string) bool {
	body := strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/")
	return strings.TrimFunc(body, func(r rune) bool {
		return r == '*' || unicode.IsSpace(r)
	}) == ""
}
