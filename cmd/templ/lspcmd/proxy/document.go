package proxy

import (
	lsp "github.com/a-h/protocol"
)

type FullTextDocument struct {
	version     int
	content     string
	lineOffsets []uint32
}

func NewFullTextDocument(version int, content string) (this *FullTextDocument) {
	return &FullTextDocument{
		version:     version,
		content:     content,
		lineOffsets: nil,
	}
}

func (this *FullTextDocument) Apply(r *lsp.Range, text string) {
	this.update([]lsp.TextDocumentContentChangeEvent{
		{
			Range: r,
			Text:  text,
		},
	}, this.version)
}

func (this *FullTextDocument) String() string {
	return this.content
}

func (this *FullTextDocument) normalizeRange(r *lsp.Range) {
	if r == nil {
		return
	}
	idxs := computeLineOffsets(this.String(), true, 0)
	var lens []uint32
	for i := 0; i < len(idxs)-1; i++ {
		lens = append(lens, idxs[i+1]-idxs[i])
	}
	if len(lens) == 0 {
		lens = append(lens, uint32(len(this.String())))
	}

	if r.Start.Line >= uint32(len(lens)) {
		r.Start.Line = uint32(len(lens) - 1)
		r.Start.Character = uint32(lens[r.Start.Line])
	}
	if r.Start.Character > uint32(lens[r.Start.Line]) {
		r.Start.Character = uint32(lens[r.Start.Line])
	}
	if r.End.Line >= uint32(len(lens)) {
		r.End.Line = uint32(len(lens) - 1)
		r.End.Character = uint32(lens[r.End.Line])
	}
	if r.End.Character > uint32(lens[r.End.Line]) {
		r.End.Character = uint32(lens[r.End.Line])
	}
}

func (this *FullTextDocument) update(changes []lsp.TextDocumentContentChangeEvent, version int) {
	for _, change := range changes {
		if this.isIncremental(change) {
			// makes sure start is before end
			r := getWellformedRange(*change.Range)

			// Normalize.
			this.normalizeRange(&r)

			// update content
			startOffset := this.offsetAt(r.Start)
			endOffset := this.offsetAt(r.End)
			this.content = this.content[0:startOffset] + change.Text + this.content[endOffset:len(this.content)]

			// update the offsets
			startLine := max(r.Start.Line, 0)
			endLine := max(r.End.Line, 0)
			lineOffsets := this.lineOffsets
			addedLineOffsets := computeLineOffsets(change.Text, false, startOffset)
			if endLine-startLine == uint32(len(addedLineOffsets)) {
				length := len(addedLineOffsets)
				for i := 0; i < length; i++ {
					lineOffsets[uint32(i)+startLine+1] = addedLineOffsets[i]
				}
			} else {
				lineOffsets = append(lineOffsets[0:startLine+1], append(addedLineOffsets, lineOffsets[endLine+1:]...)...)
				this.lineOffsets = lineOffsets
			}
			diff := len(change.Text) - int(endOffset-startOffset)
			if diff != 0 {
				length := len(lineOffsets)
				for i := int(startLine) + 1 + len(addedLineOffsets); i < length; i++ {
					lineOffsets[i] += uint32(diff)
				}
			}
		} else if this.isFull(change) {
			this.content = change.Text
			this.lineOffsets = nil
		} else {
			//TODO: Fix panic.
			panic("Unknown change event received")
		}
	}
	this.version = version
}

func (this *FullTextDocument) getLineOffsets() []uint32 {
	if this.lineOffsets == nil {
		this.lineOffsets = computeLineOffsets(this.content, true, 0)
	}
	return this.lineOffsets
}

func (this *FullTextDocument) offsetAt(position lsp.Position) uint32 {
	lineOffsets := this.getLineOffsets()
	if position.Line >= uint32(len(lineOffsets)) {
		return uint32(len(this.content))
	} else if position.Line < 0 {
		return 0
	}
	lineOffset := lineOffsets[position.Line]
	nextLineOffset := uint32(len(this.content))
	if position.Line+1 < uint32(len(lineOffsets)) {
		nextLineOffset = lineOffsets[position.Line+1]
	}
	return max(min(lineOffset+position.Character, nextLineOffset), lineOffset)
}

func min[T int | uint32](a, b T) T {
	if a < b {
		return a
	}
	return b
}
func max[T int | uint32](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func (this *FullTextDocument) isIncremental(event lsp.TextDocumentContentChangeEvent) bool {
	return event.Range != nil
}

func (this *FullTextDocument) isFull(event lsp.TextDocumentContentChangeEvent) bool {
	return event.Range == nil
}

func computeLineOffsets(text string, isAtLineStart bool, textOffset uint32) (result []uint32) {
	if isAtLineStart {
		result = append(result, textOffset)
	}
	var i int
	var ch rune
	for i, ch = range text {
		if ch == '\r' || ch == '\n' {
			if ch == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				i++
			}
			result = append(result, textOffset+uint32(i)+1)
		}
	}
	if i != len(text) {
		result = append(result, uint32(len(text)))
	}
	return result
}

func getWellformedRange(r lsp.Range) lsp.Range {
	start, end := r.Start, r.End
	if start.Line > end.Line || (start.Line == end.Line && start.Character > end.Character) {
		return lsp.Range{Start: end, End: start}
	}
	return r
}
