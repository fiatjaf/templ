package proxy

import (
	"strings"

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

func (this *FullTextDocument) getText(r *lsp.Range) string {
	if r != nil {
		start := this.offsetAt(r.Start)
		end := this.offsetAt(r.End)
		return this.content[start:end]
	}
	return this.content
}

func splice[T any](array []T, index int, count int, items ...T) {
	// Delete.
	array = array[:index+copy(array[index:], array[index+1:])]
	// Insert.
	array = append(array[:index], append(items, array[index:]...)...)
	return
}

func spliceLineOffsets(slice []int, start int, deleteCount int, itemsToAdd []int) []int {
	// Remove elements
	removed := append(slice[:start], slice[start+deleteCount:]...)

	// Add new elements
	result := make([]int, 0, len(removed)+len(itemsToAdd))
	result = append(result, removed[:start]...)
	result = append(result, itemsToAdd...)
	result = append(result, removed[start:]...)

	return result
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

func (this *FullTextDocument) positionAt(offset uint32) lsp.Position {
	offset = max(min(offset, uint32(len(this.content))), 0)

	lineOffsets := this.getLineOffsets()
	low := 0
	high := len(lineOffsets)
	if high == 0 {
		return lsp.Position{Line: 0, Character: offset}
	}
	for low < high {
		mid := int(float64(low+high) / 2.0)
		if lineOffsets[mid] > offset {
			high = mid
		} else {
			low = mid + 1
		}
	}
	// low is the least x for which the line offset is larger than the current offset
	// or array.length if no line offset is larger than the current offset
	line := uint32(low - 1)
	return lsp.Position{Line: line, Character: offset - lineOffsets[line]}
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

func (this *FullTextDocument) lineCount() int {
	return len(this.getLineOffsets())
}

func (this *FullTextDocument) isIncremental(event lsp.TextDocumentContentChangeEvent) bool {
	return event.Range != nil
}

func (this *FullTextDocument) isFull(event lsp.TextDocumentContentChangeEvent) bool {
	return event.Range == nil
}

func applyEdits(document *FullTextDocument, edits []lsp.TextEdit) (text string) {
	text = document.getText(nil)
	compare := func(a lsp.TextEdit, b lsp.TextEdit) int {
		diff := int(a.Range.Start.Line - b.Range.Start.Line)
		if diff == 0 {
			return int(a.Range.Start.Character - b.Range.Start.Character)
		}
		return diff
	}
	wellFormedEdits := make([]lsp.TextEdit, len(edits))
	for i := 0; i < len(edits); i++ {
		wellFormedEdits[i] = getWellformedEdit(edits[i])
	}
	sortedEdits := mergeSort(wellFormedEdits, compare)
	var lastModifiedOffset uint32
	var spans []string
	for _, e := range sortedEdits {
		startOffset := document.offsetAt(e.Range.Start)
		if startOffset < lastModifiedOffset {
			//TODO: Replace panic.
			panic("Overlapping edit")
		} else if startOffset > lastModifiedOffset {
			spans = append(spans, text[lastModifiedOffset:startOffset])
		}
		if len(e.NewText) > 0 {
			spans = append(spans, e.NewText)
		}
		lastModifiedOffset = document.offsetAt(e.Range.End)
	}
	spans = append(spans, text[lastModifiedOffset:])
	return strings.Join(spans, "")
}

func mergeSort[T any](data []T, compare func(a T, b T) int) []T {
	if len(data) <= 1 {
		// sorted
		return data
	}
	p := (len(data) / 2)
	left := data[0:p]
	right := data[p:]

	mergeSort(left, compare)
	mergeSort(right, compare)

	var leftIdx, rightIdx, i int
	for leftIdx < len(left) && rightIdx < len(right) {
		ret := compare(left[leftIdx], right[rightIdx])
		if ret <= 0 {
			// smaller_equal -> take left to preserve order
			i++
			leftIdx++
			data[i] = left[leftIdx]
		} else {
			// greater -> take right
			i++
			leftIdx++
			data[i] = right[rightIdx]
		}
	}
	for leftIdx < len(left) {
		i++
		leftIdx++
		data[i] = left[leftIdx]
	}
	for rightIdx < len(right) {
		i++
		rightIdx++
		data[i] = right[rightIdx]
	}
	return data
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
	start := r.Start
	end := r.End
	if start.Line > end.Line || (start.Line == end.Line && start.Character > end.Character) {
		return lsp.Range{Start: end, End: start}
	}
	return r
}

func getWellformedEdit(textEdit lsp.TextEdit) lsp.TextEdit {
	r := getWellformedRange(textEdit.Range)
	if r != textEdit.Range {
		return lsp.TextEdit{NewText: textEdit.NewText, Range: r}
	}
	return textEdit
}
