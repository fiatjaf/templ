package parser

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNodeString(t *testing.T) {
	tests := []struct {
		name     string
		input    Node
		expected string
	}{
		{
			name:     "whitespace is not adjusted",
			input:    Whitespace{Value: `\t\n`},
			expected: `\t\n`,
		},
		{
			name:     "DocType values are preserved",
			input:    DocType{Value: `html`},
			expected: `<!DOCTYPE html>`,
		},
		{
			name:     "Text is preserved",
			input:    Text{Value: `This is a sentence.`},
			expected: `This is a sentence.`,
		},
		{
			name:     "Void elements are preserved",
			input:    Element{Name: "br"},
			expected: `<br/>`,
		},
		{
			name: "Void elements can have attributes",
			input: Element{
				Name: "br",
				Attributes: []Attribute{
					BoolConstantAttribute{
						Name: "noshade",
					},
				},
			},
			expected: `<br noshade/>`,
		},
		{
			name: "Void elements can have constant attributes",
			input: Element{
				Name: "br",
				Attributes: []Attribute{
					ConstantAttribute{
						Name:  "attr",
						Value: "value",
					},
				},
			},
			expected: `<br attr="value"/>`,
		},
		{
			name: "Void elements can have multiline attributes",
			input: Element{
				Name: "br",
				Attributes: []Attribute{
					ConditionalAttribute{
						Expression: Expression{
							Value: "true",
						},
						Then: []Attribute{
							ConstantAttribute{
								Name:  "class",
								Value: "truthy",
							},
						},
						Else: []Attribute{},
					},
					BoolConstantAttribute{
						Name: "noshade",
					},
				},
			},
			expected: `<br
	if true {
		class="truthy"
	}
	noshade/>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if isNode := test.input.IsNode(); !isNode {
				t.Errorf("unexpected false")
			}

			sb := new(strings.Builder)
			iw := &IndentWriter{
				w: sb,
			}
			if err := test.input.Write(iw); err != nil {
				t.Fatal(err)
			}
			actual := sb.String()
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("Input:\n\n%v", test.input)
				t.Errorf("Diff:\n\n%v", diff)
				t.Errorf("Expected:\n\n%s", showHiddenChars(test.expected))
				t.Errorf("Actual:\n\n%s", showHiddenChars(actual))
			}
		})
	}
}

func showHiddenChars(s string) string {
	s = strings.Replace(s, " ", ".", -1)
	s = strings.Replace(s, "\t", " →", -1)
	return s
}
