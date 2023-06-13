package parser

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIndentWriter(t *testing.T) {
	tests := []struct {
		name      string
		operation func(w *IndentWriter)
		expected  string
	}{
		{
			name: "It operates as a standard writer",
			operation: func(w *IndentWriter) {
				io.WriteString(w, "This is a test\n")
			},
			expected: "This is a test\n",
		},
		{
			name: "Text is indented",
			operation: func(w *IndentWriter) {
				w.indent++
				io.WriteString(w, "Line 1\n")
				io.WriteString(w, "Line 2\n")
			},
			expected: `	Line 1
	Line 2
`,
		},
		{
			name: "Indentation can vary",
			operation: func(w *IndentWriter) {
				io.WriteString(w, "a\n")
				w.indent++
				io.WriteString(w, "b\n")
				w.indent--
				io.WriteString(w, "c\n")
			},
			expected: `a
	b
c
`,
		},
		{
			name: "Multiline text is indented",
			operation: func(w *IndentWriter) {
				w.indent++
				w.WriteString("a\nb\nc\n")
			},
			expected: "\ta\n\tb\n\tc\n",
		},
		{
			name: "Multiple strings can be written in a single operation",
			operation: func(w *IndentWriter) {
				w.indent++
				w.WriteStrings("a\n", "b\n", "c\n")
			},
			expected: "\ta\n\tb\n\tc\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sb := new(strings.Builder)
			w := NewIndentWriter(sb)
			test.operation(w)
			if diff := cmp.Diff(test.expected, sb.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
