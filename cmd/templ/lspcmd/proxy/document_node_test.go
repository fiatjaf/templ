package proxy

import (
	"testing"

	lsp "github.com/a-h/protocol"
)

//test('Incrementally replacing single-line content, more chars', () => {
//const document = newDocument('function abc() {\n  console.log("hello, world!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: 'hello, test case!!!', range: Ranges.forSubstring(document, 'hello, world!') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'function abc() {\n  console.log("hello, test case!!!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing single-line content, less chars', () => {
//const document = newDocument('function abc() {\n  console.log("hello, world!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: 'hey', range: Ranges.forSubstring(document, 'hello, world!') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'function abc() {\n  console.log("hey");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing single-line content, same num of chars', () => {
//const document = newDocument('function abc() {\n  console.log("hello, world!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: 'world, hello!', range: Ranges.forSubstring(document, 'hello, world!') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'function abc() {\n  console.log("world, hello!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing multi-line content, more lines', () => {
//const document = newDocument('function abc() {\n  console.log("hello, world!");\n}');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: '\n//hello\nfunction d(){', range: Ranges.forSubstring(document, 'function abc() {') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), '\n//hello\nfunction d(){\n  console.log("hello, world!");\n}');
//assert.equal(document.lineCount, 5);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing multi-line content, less lines', () => {
//const document = newDocument('a1\nb1\na2\nb2\na3\nb3\na4\nb4\n');
//assert.equal(document.lineCount, 9);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: 'xx\nyy', range: Ranges.forSubstring(document, '\na3\nb3\na4\nb4\n') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'a1\nb1\na2\nb2xx\nyy');
//assert.equal(document.lineCount, 5);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing multi-line content, same num of lines and chars', () => {
//const document = newDocument('a1\nb1\na2\nb2\na3\nb3\na4\nb4\n');
//assert.equal(document.lineCount, 9);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: '\nxx1\nxx2', range: Ranges.forSubstring(document, 'a2\nb2\na3') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'a1\nb1\n\nxx1\nxx2\nb3\na4\nb4\n');
//assert.equal(document.lineCount, 9);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing multi-line content, same num of lines but diff chars', () => {
//const document = newDocument('a1\nb1\na2\nb2\na3\nb3\na4\nb4\n');
//assert.equal(document.lineCount, 9);
//assertValidLineNumbers(document);
//TextDocument.update(document, [{ text: '\ny\n', range: Ranges.forSubstring(document, 'a2\nb2\na3') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'a1\nb1\n\ny\n\nb3\na4\nb4\n');
//assert.equal(document.lineCount, 9);
//assertValidLineNumbers(document);
//});

//test('Incrementally replacing multi-line content, huge number of lines', () => {
//const document = newDocument('a1\ncc\nb1');
//assert.equal(document.lineCount, 3);
//assertValidLineNumbers(document);
//const text = new Array(20000).join('\ndd'); // a string with 19999 `\n`
//TextDocument.update(document, [{ text, range: Ranges.forSubstring(document, '\ncc') }], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'a1' + text + '\nb1');
//assert.equal(document.lineCount, 20001);
//assertValidLineNumbers(document);
//});

//test('Several incremental content changes', () => {
//const document = newDocument('function abc() {\n  console.log("hello, world!");\n}');
//TextDocument.update(document, [
//{ text: 'defg', range: Ranges.create(0, 12, 0, 12) },
//{ text: 'hello, test case!!!', range: Ranges.create(1, 15, 1, 28) },
//{ text: 'hij', range: Ranges.create(0, 16, 0, 16) },
//], 1);
//assert.strictEqual(document.version, 1);
//assert.strictEqual(document.getText(), 'function abcdefghij() {\n  console.log("hello, test case!!!");\n}');
//assertValidLineNumbers(document);
//});

//test('Basic append', () => {
//let document = newDocument('foooo\nbar\nbaz');

//assert.equal(document.offsetAt(Positions.create(2, 0)), 10);

//TextDocument.update(document, [{ text: ' some extra content', range: Ranges.create(1, 3, 1, 3) }], 1);
//assert.equal(document.getText(), 'foooo\nbar some extra content\nbaz');
//assert.equal(document.version, 1);
//assert.equal(document.offsetAt(Positions.create(2, 0)), 29);
//assertValidLineNumbers(document);
//});

//test('Multi-line append', () => {
//let document = newDocument('foooo\nbar\nbaz');

//assert.equal(document.offsetAt(Positions.create(2, 0)), 10);

//TextDocument.update(document, [{ text: ' some extra\ncontent', range: Ranges.create(1, 3, 1, 3) }], 1);
//assert.equal(document.getText(), 'foooo\nbar some extra\ncontent\nbaz');
//assert.equal(document.version, 1);
//assert.equal(document.offsetAt(Positions.create(3, 0)), 29);
//assert.equal(document.lineCount, 4);
//assertValidLineNumbers(document);
//});

//test('Basic delete', () => {
//let document = newDocument('foooo\nbar\nbaz');

//assert.equal(document.offsetAt(Positions.create(2, 0)), 10);

//TextDocument.update(document, [{ text: '', range: Ranges.create(1, 0, 1, 3) }], 1);
//assert.equal(document.getText(), 'foooo\n\nbaz');
//assert.equal(document.version, 1);
//assert.equal(document.offsetAt(Positions.create(2, 0)), 7);
//assertValidLineNumbers(document);
//});

//test('Multi-line delete', () => {
//let lm = newDocument('foooo\nbar\nbaz');

//assert.equal(lm.offsetAt(Positions.create(2, 0)), 10);

//TextDocument.update(lm, [{ text: '', range: Ranges.create(0, 5, 1, 3) }], 1);
//assert.equal(lm.getText(), 'foooo\nbaz');
//assert.equal(lm.version, 1);
//assert.equal(lm.offsetAt(Positions.create(1, 0)), 6);
//assertValidLineNumbers(lm);
//});

//test('Single character replace', () => {
//let document = newDocument('foooo\nbar\nbaz');

//assert.equal(document.offsetAt(Positions.create(2, 0)), 10);

//TextDocument.update(document, [{ text: 'z', range: Ranges.create(1, 2, 1, 3) }], 2);
//assert.equal(document.getText(), 'foooo\nbaz\nbaz');
//assert.equal(document.version, 2);
//assert.equal(document.offsetAt(Positions.create(2, 0)), 10);
//assertValidLineNumbers(document);
//});

//test('Multi-character replace', () => {
//let lm = newDocument('foo\nbar');

//assert.equal(lm.offsetAt(Positions.create(1, 0)), 4);

//TextDocument.update(lm, [{ text: 'foobar', range: Ranges.create(1, 0, 1, 3) }], 1);
//assert.equal(lm.getText(), 'foo\nfoobar');
//assert.equal(lm.version, 1);
//assert.equal(lm.offsetAt(Positions.create(1, 0)), 4);
//assertValidLineNumbers(lm);
//});

//test('Invalid update ranges', () => {
//// Before the document starts -> before the document starts
//let document = newDocument('foo\nbar');
//TextDocument.update(document, [{ text: 'abc123', range: Ranges.create(-2, 0, -1, 3) }], 2);
//assert.equal(document.getText(), 'abc123foo\nbar');
//assert.equal(document.version, 2);
//assertValidLineNumbers(document);

//// Before the document starts -> the middle of document
//document = newDocument('foo\nbar');
//TextDocument.update(document, [{ text: 'foobar', range: Ranges.create(-1, 0, 0, 3) }], 2);
//assert.equal(document.getText(), 'foobar\nbar');
//assert.equal(document.version, 2);
//assert.equal(document.offsetAt(Positions.create(1, 0)), 7);
//assertValidLineNumbers(document);

//// The middle of document -> after the document ends
//document = newDocument('foo\nbar');
//TextDocument.update(document, [{ text: 'foobar', range: Ranges.create(1, 0, 1, 10) }], 2);
//assert.equal(document.getText(), 'foo\nfoobar');
//assert.equal(document.version, 2);
//assert.equal(document.offsetAt(Positions.create(1, 1000)), 10);
//assertValidLineNumbers(document);

//// After the document ends -> after the document ends
//document = newDocument('foo\nbar');
//TextDocument.update(document, [{ text: 'abc123', range: Ranges.create(3, 0, 6, 10) }], 2);
//assert.equal(document.getText(), 'foo\nbarabc123');
//assert.equal(document.version, 2);
//assertValidLineNumbers(document);

//// Before the document starts -> after the document ends
//document = newDocument('foo\nbar');
//TextDocument.update(document, [{ text: 'entirely new content', range: Ranges.create(-1, 1, 2, 10000) }], 2);
//assert.equal(document.getText(), 'entirely new content');
//assert.equal(document.version, 2);
//assert.equal(document.lineCount, 1);
//assertValidLineNumbers(document);
//});
//});

func BenchmarkDocumentNodeContents(b *testing.B) {
	start := `package testcall

templ personTemplate(p person) {
	<div>
		<h1>{ p.name }</h1>
	</div>
}
`
	operations := func(d *FullTextDocument) {
		d.Apply(&lsp.Range{
			Start: lsp.Position{
				Line:      4,
				Character: 21,
			},
			End: lsp.Position{
				Line:      4,
				Character: 21,
			},
		}, "\n\t\t")
	}
	expected := `package testcall

templ personTemplate(p person) {
	<div>
		<h1>{ p.name }</h1>
		
	</div>
}
`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := NewFullTextDocument(1, start)
		operations(d)
		if d.String() != expected {
			b.Fatalf("comparison failed: %v", d.String())
		}
	}
}
