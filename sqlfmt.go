package sqlfmt

import (
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/parser"
	"github.com/labulakalia/sqlfmt/cockroach/pkg/sql/sem/tree"
	"regexp"
	"strings"
	"unicode"
)

var (
	ignoreComments = regexp.MustCompile(`^--.*\s*`)
)

func getPrettyCfg(lineWidth int) *tree.PrettyCfg {
	prettyCfg := &tree.PrettyCfg{
		LineWidth: lineWidth,
		Simplify:  true,
		TabWidth:  4,
		UseTabs:   true,
		Align:     tree.PrettyNoAlign,
		Case:      strings.ToUpper,
		JSONFmt:   true,
	}
	return prettyCfg
}

func FormatSQL(stmt string) (string, error) {
	prettyCfg := getPrettyCfg(len(stmt) / 2 + 10)
	var prettied strings.Builder
	stmt = strings.TrimSpace(stmt)
	hasContent := false
	// Trim comments, preserving whitespace after them.
	for {
		found := ignoreComments.FindString(stmt)
		if found == "" {
			break
		}
		// Remove trailing whitespace but keep up to 2 newlines.
		prettied.WriteString(strings.TrimRightFunc(found, unicode.IsSpace))
		newlines := strings.Count(found, "\n")
		if newlines > 2 {
			newlines = 2
		}
		prettied.WriteString(strings.Repeat("\n", newlines))
		stmt = stmt[len(found):]
		hasContent = true
	}
	// Split by semicolons
	next := stmt
	if pos, _ := parser.SplitFirstStatement(stmt); pos > 0 {
		next = stmt[:pos]
		stmt = stmt[pos:]
	} else {
		stmt = ""
	}
	// This should only return 0 or 1 responses.
	allParsed, err := parser.Parse(next)
	if err != nil {
		return "", err
	}
	for _, parsed := range allParsed {
		prettied.WriteString(prettyCfg.Pretty(parsed.AST))
		prettied.WriteString(";\n")
		hasContent = true
	}
	if hasContent {
		prettied.WriteString("\n")
	}
	return strings.TrimSpace(prettied.String()), nil
}
