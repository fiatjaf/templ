package parser

import (
	"bytes"
	"io"
)

type IndentWriter struct {
	block            bool
	indent           int
	w                io.Writer
	justWroteNewLine bool
}

func NewIndentWriter(w io.Writer) *IndentWriter {
	return &IndentWriter{
		w:                w,
		justWroteNewLine: true,
	}
}

func (iw *IndentWriter) WriteStrings(ss ...string) (err error) {
	for _, s := range ss {
		if _, err = iw.Write([]byte(s)); err != nil {
			return err
		}
	}
	return err
}

func (iw *IndentWriter) WriteString(s string) (err error) {
	_, err = iw.Write([]byte(s))
	return err
}

func (iw *IndentWriter) Write(b []byte) (n int, err error) {
	indent := bytes.Repeat([]byte{'\t'}, iw.indent)
	for _, bb := range b {
		// Write indent if the last char was a newline.
		currentIsNewLine := bb == '\n'
		if iw.justWroteNewLine && !currentIsNewLine {
			// Only write the indent if there's something in the line.
			nn, err := iw.w.Write(indent)
			if err != nil {
				return nn, err
			}
			n += nn
		}
		// Write out the char.
		nn, err := iw.w.Write([]byte{bb})
		if err != nil {
			return nn, err
		}
		n += nn
		iw.justWroteNewLine = bb == '\n'
	}
	return n, err
}
