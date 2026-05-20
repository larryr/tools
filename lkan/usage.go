package lkan

import (
	_ "embed"
	"io"
	"strings"
)

//go:embed doc.go
var docSource string

// Usage writes the package help text — the leading block comment of doc.go —
// to w. Output is plain text suitable for a terminal or for an LLM to ingest
// verbatim.
func Usage(w io.Writer) error {
	_, after, _ := strings.Cut(docSource, "/*")
	body, _, _ := strings.Cut(after, "*/")
	_, err := io.WriteString(w, strings.TrimLeft(body, "\n"))
	return err
}
