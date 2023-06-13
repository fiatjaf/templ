package parser

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/a-h/parse"
)

// package parser
//
// import "strings"
// import strs "strings"
//
// css AddressLineStyle() {
//   background-color: #ff0000;
//   color: #ffffff;
// }
//
// templ RenderAddress(addr Address) {
// 	<div style={ AddressLineStyle() }>{ addr.Address1 }</div>
// 	<div>{ addr.Address2 }</div>
// 	<div>{ addr.Address3 }</div>
// 	<div>{ addr.Address4 }</div>
// }
//
// templ Render(p Person) {
//    <div>
//      <div>{ p.Name() }</div>
//      <a href={ p.URL }>{ strings.ToUpper(p.Name()) }</a>
//      <div>
//          if p.Type == "test" {
//             <span>{ "Test user" }</span>
//          } else {
// 	    	<span>{ "Not test user" }</span>
//          }
//          for _, v := range p.Addresses {
//             {! call RenderAddress(v) }
//          }
//      </div>
//    </div>
// }

// Source mapping to map from the source code of the template to the
// in-memory representation.
type Position struct {
	Index int64
	Line  uint32
	Col   uint32
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, col %d (index %d)", p.Line, p.Col, p.Index)
}

// NewPosition initialises a position.
func NewPosition(index int64, line, col uint32) Position {
	return Position{
		Index: index,
		Line:  line,
		Col:   col,
	}
}

// NewExpression creates a Go expression.
func NewExpression(value string, from, to parse.Position) Expression {
	return Expression{
		Value: value,
		Range: Range{
			From: Position{
				Index: int64(from.Index),
				Line:  uint32(from.Line),
				Col:   uint32(from.Col),
			},
			To: Position{
				Index: int64(to.Index),
				Line:  uint32(to.Line),
				Col:   uint32(to.Col),
			},
		},
	}
}

// Range of text within a file.
type Range struct {
	From Position
	To   Position
}

// Expression containing Go code.
type Expression struct {
	Value string
	Range Range
}

type TemplateFile struct {
	Package Package
	Nodes   []TemplateFileNode
}

func (tf TemplateFile) Write(w io.Writer) error {
	iw := NewIndentWriter(w)
	if err := iw.WriteStrings(tf.Package.String(), "\n\n"); err != nil {
		return err
	}
	for _, c := range tf.Nodes {
		if err := c.Write(iw); err != nil {
			return err
		}
		if err := iw.WriteString("\n\n"); err != nil {
			return err
		}
	}
	return nil
}

// TemplateFileNode can be a Template, CSS, Script or Go.
type TemplateFileNode interface {
	IsTemplateFileNode() bool
	Write(w *IndentWriter) error
}

// GoExpression within a TemplateFile
type GoExpression struct {
	Expression Expression
}

func (exp GoExpression) IsTemplateFileNode() bool { return true }
func (exp GoExpression) Write(w *IndentWriter) error {
	return w.WriteString(exp.Expression.Value)
}

type Package struct {
	Expression Expression
}

func (p Package) String() string {
	return p.Expression.Value
}

// Whitespace.
type Whitespace struct {
	Value string
}

func (ws Whitespace) IsNode() bool { return true }
func (ws Whitespace) Write(w *IndentWriter) error {
	return w.WriteString(ws.Value)
}

// CSS definition.
//
//	css Name() {
//	  color: #ffffff;
//	  background-color: { constants.BackgroundColor };
//	  background-image: url('./somewhere.png');
//	}
type CSSTemplate struct {
	Name       Expression
	Properties []CSSProperty
}

func (css CSSTemplate) IsTemplateFileNode() bool { return true }
func (css CSSTemplate) Write(w *IndentWriter) error {
	if err := w.WriteStrings("css ", css.Name.Value, "() {\n"); err != nil {
		return err
	}
	{
		w.indent++
		for _, p := range css.Properties {
			if err := w.WriteStrings(p.String(), "\n"); err != nil {
				return err
			}
		}
		w.indent--
	}
	if err := w.WriteString("}\n"); err != nil {
		return err
	}
	return nil
}

// CSSProperty is a CSS property and value pair.
type CSSProperty interface {
	IsCSSProperty() bool
	String() string
}

// color: #ffffff;
type ConstantCSSProperty struct {
	Name  string
	Value string
}

func (c ConstantCSSProperty) IsCSSProperty() bool { return true }
func (c ConstantCSSProperty) String() string {
	return fmt.Sprintf("%s: %s;\n", c.Name, c.Value)
}

// background-color: { constants.BackgroundColor };
type ExpressionCSSProperty struct {
	Name  string
	Value StringExpression
}

func (c ExpressionCSSProperty) IsCSSProperty() bool { return true }
func (c ExpressionCSSProperty) String() string {
	return fmt.Sprintf("%s: %s;\n", c.Name, c.Value.Expression.Value)
}

// <!DOCTYPE html>
type DocType struct {
	Value string
}

func (dt DocType) IsNode() bool { return true }
func (dt DocType) Write(w *IndentWriter) error {
	return w.WriteString("<!DOCTYPE " + dt.Value + ">")
}

// HTMLTemplate definition.
//
//	templ Name(p Parameter) {
//	  if ... {
//	      <Element></Element>
//	  }
//	}
type HTMLTemplate struct {
	Expression Expression
	Children   []Node
}

func (t HTMLTemplate) IsTemplateFileNode() bool { return true }
func (t HTMLTemplate) Write(w *IndentWriter) error {
	if err := w.WriteString("templ " + t.Expression.Value + " {"); err != nil {
		return err
	}
	w.indent++
	if err := w.WriteString("\n"); err != nil {
		return err
	}
	for _, c := range t.Children {
		if err := c.Write(w); err != nil {
			return err
		}
	}
	w.indent--
	if err := w.WriteString("\n}"); err != nil {
		return err
	}
	return nil
}

// A Node appears within a template, e.g. an StringExpression, Element, IfExpression etc.
type Node interface {
	IsNode() bool
	Write(w *IndentWriter) error
}

// Text node within the document.
type Text struct {
	// Value is the raw HTML encoded value.
	Value string
}

func (t Text) IsNode() bool { return true }
func (t Text) Write(w *IndentWriter) error {
	if w.block {
		return w.WriteStrings(t.Value, "\n")
	}
	return w.WriteString(t.Value)
}

// <a .../> or <div ...>...</div>
type Element struct {
	Name       string
	Attributes []Attribute
	Children   []Node
}

var voidElements = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "command": {}, "embed": {}, "hr": {}, "img": {}, "input": {}, "keygen": {}, "link": {}, "meta": {}, "param": {}, "source": {}, "track": {}, "wbr": {}}

// https://www.w3.org/TR/2011/WD-html-markup-20110113/syntax.html#void-element
func (e Element) IsVoidElement() bool {
	_, ok := voidElements[e.Name]
	return ok
}

var blockElements = map[string]struct{}{
	"address": {}, "article": {}, "aside": {}, "body": {}, "blockquote": {}, "canvas": {}, "dd": {}, "div": {}, "dl": {}, "dt": {}, "fieldset": {}, "figcaption": {}, "figure": {}, "footer": {}, "form": {}, "h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {}, "head": {}, "header": {}, "hr": {}, "html": {}, "li": {}, "main": {}, "meta": {}, "nav": {}, "noscript": {}, "ol": {}, "p": {}, "pre": {}, "script": {}, "section": {}, "table": {}, "tr": {}, "th": {}, "td": {}, "template": {}, "tfoot": {}, "turbo-stream": {}, "ul": {}, "video": {},
}

func (e Element) isBlockElement() bool {
	_, ok := blockElements[e.Name]
	return ok
}

func (e Element) containsBlockElement() bool {
	for _, c := range e.Children {
		switch n := c.(type) {
		case Whitespace:
			continue
		case Element:
			if n.isBlockElement() {
				return true
			}
			continue
		case StringExpression:
			continue
		case Text:
			continue
		case TemplElementExpression:
			if len(n.Children) > 0 {
				return true
			}
			continue
		}
		// Any template elements should be considered block.
		return true
	}
	return false
}

// Validate that no invalid expressions have been used.
func (e Element) Validate() (msgs []string, ok bool) {
	// Validate that style attributes are constant.
	for _, attr := range e.Attributes {
		if exprAttr, isExprAttr := attr.(ExpressionAttribute); isExprAttr {
			if strings.EqualFold(exprAttr.Name, "style") {
				msgs = append(msgs, "invalid style attribute: style attributes cannot be a templ expression")
			}
		}
	}
	// Validate that script and style tags don't contain expressions.
	if strings.EqualFold(e.Name, "script") || strings.EqualFold(e.Name, "style") {
		if containsNonTextNodes(e.Children) {
			msgs = append(msgs, "invalid node contents: script and style attributes must only contain text")
		}
	}
	return msgs, len(msgs) == 0
}

func containsNonTextNodes(nodes []Node) bool {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		switch n.(type) {
		case Text:
			continue
		case Whitespace:
			continue
		default:
			return true
		}
	}
	return false
}

func (e Element) IsNode() bool { return true }
func (e Element) Write(w *IndentWriter) error {
	if e.isBlockElement() || e.containsBlockElement() {
		if err := w.WriteString("\n"); err != nil {
			return err
		}
	}
	if err := w.WriteString("<" + e.Name); err != nil {
		return err
	}

	var closeAngleBracketIndent string
	{
		w.indent++
		var previousWasMultiline bool
		for i := 0; i < len(e.Attributes); i++ {
			a := e.Attributes[i]
			currentIsMultiline := a.IsMultilineAttr()
			// Only the conditional attributes get put on a newline.
			if !(previousWasMultiline || currentIsMultiline) {
				if err := w.WriteString(" "); err != nil {
					return err
				}
			}
			if err := a.Write(w); err != nil {
				return err
			}
			previousWasMultiline = currentIsMultiline
		}
		if previousWasMultiline {
			closeAngleBracketIndent = "\n"
		}
		w.indent--
	}

	// Exit early for void tags.
	if e.IsVoidElement() {
		if err := w.WriteString(closeAngleBracketIndent + "/>"); err != nil {
			return err
		}
		if e.isBlockElement() || e.containsBlockElement() {
			if err := w.WriteString("\n"); err != nil {
				return err
			}
		}
		return nil
	}

	// Complete the open tag.
	if err := w.WriteString(closeAngleBracketIndent + ">"); err != nil {
		return err
	}

	// If it's block, do a newline.
	if e.isBlockElement() || e.containsBlockElement() {
		if err := w.WriteString("\n"); err != nil {
			return err
		}
	}

	// Write out the nodes.
	{
		w.indent++
		for _, c := range e.Children {
			//TODO: Handle multiple nesting levels.
			if err := c.Write(w); err != nil {
				return err
			}
		}
		w.indent--
	}

	// Close up.
	if e.isBlockElement() || e.containsBlockElement() {
		if err := w.WriteString("\n"); err != nil {
			return err
		}
	}
	if err := w.WriteString("</" + e.Name + ">"); err != nil {
		return err
	}
	if e.isBlockElement() || e.containsBlockElement() {
		if err := w.WriteString("\n"); err != nil {
			return err
		}
	}

	return nil
}

type RawElement struct {
	Name       string
	Attributes []Attribute
	Contents   string
}

func (e RawElement) IsNode() bool { return true }
func (e RawElement) Write(w *IndentWriter) error {
	if err := w.WriteString("<" + e.Name); err != nil {
		return err
	}
	for i := 0; i < len(e.Attributes); i++ {
		if err := w.WriteString(" "); err != nil {
			return err
		}
		if err := e.Attributes[i].Write(w); err != nil {
			return err
		}
	}
	if err := w.WriteString(">"); err != nil {
		return err
	}
	// Contents.
	if err := w.WriteString(e.Contents); err != nil {
		return err
	}
	// Close.
	if err := w.WriteString("</" + e.Name + ">"); err != nil {
		return err
	}
	return nil
}

type Attribute interface {
	IsMultilineAttr() bool
	Write(w *IndentWriter) error
}

// <hr noshade/>
type BoolConstantAttribute struct {
	Name string
}

func (bca BoolConstantAttribute) IsMultilineAttr() bool { return false }
func (bca BoolConstantAttribute) Write(w *IndentWriter) error {
	return w.WriteString(bca.Name)
}

// href=""
type ConstantAttribute struct {
	Name  string
	Value string
}

func (ca ConstantAttribute) IsMultilineAttr() bool { return false }
func (ca ConstantAttribute) Write(w *IndentWriter) error {
	return w.WriteStrings(ca.Name, `="`, html.EscapeString(ca.Value), `"`)
}

// href={ templ.Bool(...) }
type BoolExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea BoolExpressionAttribute) IsMultilineAttr() bool { return false }
func (ea BoolExpressionAttribute) Write(w *IndentWriter) error {
	return w.WriteStrings(ea.Name, `?={ `, ea.Expression.Value, ` }`)
}

// href={ ... }
type ExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea ExpressionAttribute) IsMultilineAttr() bool { return false }
func (ea ExpressionAttribute) Write(w *IndentWriter) error {
	return w.WriteStrings(ea.Name, `={ `, ea.Expression.Value, ` }`)
}

//	<a href="test" \
//		if active {
//	   class="isActive"
//	 }
type ConditionalAttribute struct {
	Expression Expression
	Then       []Attribute
	Else       []Attribute
}

func (ca ConditionalAttribute) IsMultilineAttr() bool { return true }
func (ca ConditionalAttribute) Write(w *IndentWriter) error {
	if err := w.WriteStrings("\nif ", ca.Expression.Value, " {"); err != nil {
		return err
	}
	// Then.
	{
		w.indent++
		if err := w.WriteString("\n"); err != nil {
			return err
		}
		for _, attr := range ca.Then {
			if err := attr.Write(w); err != nil {
				return err
			}
			if err := w.WriteString("\n"); err != nil {
				return err
			}
		}
		w.indent--
	}
	// No else.
	if len(ca.Else) == 0 {
		if err := w.WriteString("}\n"); err != nil {
			return err
		}
		return nil
	}
	// Else.
	if err := w.WriteString(" else {\n"); err != nil {
		return err
	}
	{
		w.indent++
		for _, attr := range ca.Else {
			if err := attr.Write(w); err != nil {
				return err
			}
			if err := w.WriteString("\n"); err != nil {
				return err
			}
		}
		w.indent--
	}
	if err := w.WriteString("}\n"); err != nil {
		return err
	}
	return nil
}

// Nodes.

// CallTemplateExpression can be used to create and render a template using data.
// {! Other(p.First, p.Last) }
// or it can be used to render a template parameter.
// {! v }
type CallTemplateExpression struct {
	// Expression returns a template to execute.
	Expression Expression
}

func (cte CallTemplateExpression) IsNode() bool { return true }
func (cte CallTemplateExpression) Write(w *IndentWriter) error {
	return w.WriteString(`{! ` + cte.Expression.Value + ` }`)
}

// TemplElementExpression can be used to create and render a template using data.
// @Other(p.First, p.Last)
// or it can be used to render a template parameter.
// @v
type TemplElementExpression struct {
	// Expression returns a template to execute.
	Expression Expression
	// Children returns the elements in a block element.
	Children []Node
}

func (tee TemplElementExpression) IsNode() bool { return true }
func (tee TemplElementExpression) Write(w *IndentWriter) error {
	if len(tee.Children) == 0 {
		return w.WriteString(fmt.Sprintf("@%s", tee.Expression.Value))
	}
	if err := w.WriteString(fmt.Sprintf("@%s {\n", tee.Expression.Value)); err != nil {
		return err
	}
	w.indent++
	for _, c := range tee.Children {
		if err := c.Write(w); err != nil {
			return err
		}
	}
	w.indent--
	if err := w.WriteString("\n}"); err != nil {
		return err
	}
	return nil
}

// ChildrenExpression can be used to rended the children of a templ element.
// { children ... }
type ChildrenExpression struct{}

func (ChildrenExpression) IsNode() bool { return true }
func (ChildrenExpression) Write(w *IndentWriter) error {
	return w.WriteString("{ children... }")
}

// if p.Type == "test" && p.thing {
// }
type IfExpression struct {
	Expression Expression
	Then       []Node
	ElseIfs    []ElseIfExpression
	Else       []Node
}

type ElseIfExpression struct {
	Expression Expression
	Then       []Node
}

func (n IfExpression) IsNode() bool { return true }
func (n IfExpression) Write(w *IndentWriter) error {
	if err := w.WriteString("if " + n.Expression.Value + " {\n"); err != nil {
		return err
	}
	{
		w.indent++
		for _, n := range n.Then {
			if err := n.Write(w); err != nil {
				return err
			}
		}
		w.indent--
	}
	for _, n := range n.ElseIfs {
		if err := w.WriteString("} else if "); err != nil {
			return err
		}
		if err := w.WriteString(n.Expression.Value); err != nil {
			return err
		}
		if err := w.WriteString(" {\n"); err != nil {
			return err
		}
		{
			w.indent++
			for _, n := range n.Then {
				if err := n.Write(w); err != nil {
					return err
				}
			}
			w.indent--
		}
	}
	if len(n.Else) > 0 {
		if err := w.WriteString("} else {\n"); err != nil {
			return err
		}
		{
			w.indent++
			for _, n := range n.Else {
				if err := n.Write(w); err != nil {
					return err
				}
			}
			w.indent--
		}
	}
	if err := w.WriteString("}"); err != nil {
		return err
	}
	return nil
}

//	switch p.Type {
//	 case "Something":
//	}
type SwitchExpression struct {
	Expression Expression
	Cases      []CaseExpression
}

func (se SwitchExpression) IsNode() bool { return true }
func (se SwitchExpression) Write(w *IndentWriter) error {
	if err := w.WriteStrings("switch ", se.Expression.Value, " {\n"); err != nil {
		return err
	}
	{
		w.indent++
		for _, c := range se.Cases {
			if err := w.WriteStrings(c.Expression.Value, "\n"); err != nil {
				return err
			}
			{
				w.indent++
				for _, child := range c.Children {
					if err := child.Write(w); err != nil {
						return err
					}
				}
				w.indent--
			}
		}
		w.indent--
	}
	if err := w.WriteString("}"); err != nil {
		return err
	}
	return nil
}

// case "Something":
type CaseExpression struct {
	Expression Expression
	Children   []Node
}

//	for i, v := range p.Addresses {
//	  {! Address(v) }
//	}
type ForExpression struct {
	Expression Expression
	Children   []Node
}

func (fe ForExpression) IsNode() bool { return true }
func (fe ForExpression) Write(w *IndentWriter) error {
	if err := w.WriteStrings("for ", fe.Expression.Value, " {\n"); err != nil {
		return err
	}
	{
		w.indent++
		for _, n := range fe.Children {
			if err := n.Write(w); err != nil {
				return err
			}
		}
		w.indent--
	}
	if err := w.WriteString("\n}"); err != nil {
		return err
	}
	return nil
}

// StringExpression is used within HTML elements, and for style values.
// { ... }
type StringExpression struct {
	Expression Expression
}

func (se StringExpression) IsNode() bool                  { return true }
func (se StringExpression) IsStyleDeclarationValue() bool { return true }
func (se StringExpression) Write(w *IndentWriter) error {
	return w.WriteStrings(`{ `, se.Expression.Value, ` }`)
}

// ScriptTemplate is a script block.
type ScriptTemplate struct {
	Name       Expression
	Parameters Expression
	Value      string
}

func (s ScriptTemplate) IsTemplateFileNode() bool { return true }
func (s ScriptTemplate) Write(w *IndentWriter) error {
	return w.WriteStrings("script ", s.Name.Value, "(", s.Parameters.Value, ") {\n",
		s.Value, "}")
}
