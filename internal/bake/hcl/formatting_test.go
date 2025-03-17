package hcl

import (
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
)

func TestFormatting(t *testing.T) {
	testCases := []struct {
		name             string
		content          string
		indentations     []int
		indentationEdits []protocol.TextEdit
		spacingEdits     []protocol.TextEdit
	}{
		{
			name:             "formats single-line block",
			content:          "target t {  call  =  true  }",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 12},
					},
				},
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 16},
						End:   protocol.Position{Line: 0, Character: 18},
					},
				},
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 19},
						End:   protocol.Position{Line: 0, Character: 21},
					},
				},
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 25},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
			},
		},
		{
			name:         "truncate whitespace before a block type",
			content:      " target t {\n}",
			indentations: []int{0},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 1},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "indents block attributes",
			content:      "target t {\na = true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "indenting block attributes preserved comments",
			content:      "target t {\n/* */ a = true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "increases insufficient indentation from block attributes",
			content:      "target t {\n a = true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "removes excessive indentation from block attributes",
			content:      "target t {\n        a = true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 8},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "preserve whitespace if spacing correct for embedded comment before the equals",
			content:      "target t {\na /**/ = true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "remove extra whitespace after equals sign for boolean",
			content:      "target t {\na =  true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 3},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
			},
		},
		{
			name:         "remove extra whitespace after equals sign for quoted string",
			content:      "target t {\na =  \"abc\"\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 3},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
			},
		},
		{
			name:         "truncate whitespace correctly if comments found after the equals sign",
			content:      "target t {\na =  /**/  true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " /**/ ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 3},
						End:   protocol.Position{Line: 1, Character: 11},
					},
				},
			},
		},
		{
			name:         "preserve whitespace if spacing correct for embedded comment after the equals",
			content:      "target t {\na = /**/ true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "add whitespace after equals sign",
			content:      "target t {\na =true\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 3},
						End:   protocol.Position{Line: 1, Character: 3},
					},
				},
			},
		},
		{
			name:             "correct whitespace between block type and unquoted block name",
			content:          "target  t {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 6},
						End:   protocol.Position{Line: 0, Character: 8},
					},
				},
			},
		},
		{
			name:             "correct whitespace between block type and quoted block name",
			content:          "target  \"t\" {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 6},
						End:   protocol.Position{Line: 0, Character: 8},
					},
				},
			},
		},
		{
			name:             "correct whitespace between multiple unquoted block labels",
			content:          "target t1  t2  t3 {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 9},
						End:   protocol.Position{Line: 0, Character: 11},
					},
				},
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 13},
						End:   protocol.Position{Line: 0, Character: 15},
					},
				},
			},
		},
		{
			name:             "correct whitespace between multiple quoted block labels",
			content:          "target \"t1\"  \"t2\"  \"t3\" {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 11},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 17},
						End:   protocol.Position{Line: 0, Character: 19},
					},
				},
			},
		},
		{
			name:             "correct whitespace between unquoted block and the opening brace",
			content:          "target t  {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 8},
						End:   protocol.Position{Line: 0, Character: 10},
					},
				},
			},
		},
		{
			name:             "correct whitespace between quoted block and the opening brace",
			content:          "target \"t1\"  {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 11},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:             "correct whitespace between block name and the opening brace",
			content:          "target  {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: " ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 6},
						End:   protocol.Position{Line: 0, Character: 8},
					},
				},
			},
		},
		{
			name:             "remove excessive whitespace preceding a block's closing brace on a newline",
			content:          "target {\n }",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
			},
		},
		{
			name:             "comment preserved before a block's closing brace on a newline",
			content:          "target {\n /**/ }",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
			},
		},
		{
			name:         "handle args object inside a target block",
			content:      "target t {\n args = {\n var = value\n }\n}",
			indentations: []int{1, 2, 1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 1},
					},
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 1},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "handle args object inside a target block with the first key on the same line as the attribute",
			content:      "target t {\n args = { var = \"value\"\n var2 = \"value2\"\n}\n}",
			indentations: []int{1, 2, 1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 1},
					},
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 0},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:         "handle args object on a single line",
			content:      "target t {\n args = { var = value }\n}",
			indentations: []int{1},
			indentationEdits: []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
			},
			spacingEdits: []protocol.TextEdit{},
		},
		{
			name:             "remove incorrectly inserted newline before opening brace",
			content:          "target t {}",
			indentations:     []int{},
			indentationEdits: []protocol.TextEdit{},
			spacingEdits:     []protocol.TextEdit{},
		},
	}

	options := []protocol.FormattingOptions{
		{protocol.FormattingOptionInsertSpaces: false},
		{protocol.FormattingOptionInsertSpaces: true, protocol.FormattingOptionTabSize: float64(2)},
		{protocol.FormattingOptionInsertSpaces: true, protocol.FormattingOptionTabSize: float64(4)},
	}
	indentations := []string{"\t", "  ", "    "}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			for i := range 3 {
				edits, err := Formatting(doc, options[i])
				for j := range tc.indentationEdits {
					indentation := ""
					for range tc.indentations[j] {
						indentation += indentations[i]
					}
					tc.indentationEdits[j].NewText = indentation
				}

				require.NoError(t, err)
				expected := []protocol.TextEdit{}
				expected = append(expected, tc.indentationEdits...)
				expected = append(expected, tc.spacingEdits...)
				require.Equal(t, expected, edits)
			}
		})
	}
}

func TestFormattingCustom(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		edits   []protocol.TextEdit
	}{
		{
			name:    "handle whitespace after an args key's equals sign",
			content: "target t {\n  args = {\n    var  =  value\n  }\n}",
			edits: []protocol.TextEdit{
				{
					NewText: "  = ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 7},
						End:   protocol.Position{Line: 2, Character: 12},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			edits, err := Formatting(doc, protocol.FormattingOptions{
				protocol.FormattingOptionInsertSpaces: true, protocol.FormattingOptionTabSize: float64(2),
			})
			require.NoError(t, err)
			require.Equal(t, tc.edits, edits)
		})
	}
}

func TestFormattingUnchanged(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		edits   []protocol.TextEdit
	}{
		{
			name:    "spacing preserved",
			content: "target t {\n  call = \"check\"\n}",
			edits:   []protocol.TextEdit{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			edits, err := Formatting(doc, protocol.FormattingOptions{
				protocol.FormattingOptionInsertSpaces: true, protocol.FormattingOptionTabSize: float64(2),
			})
			require.NoError(t, err)
			require.Len(t, edits, 0)
		})
	}
}
