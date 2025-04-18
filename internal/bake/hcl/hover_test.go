package hcl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/bake/hcl/parser"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestHover(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		result    *protocol.Hover
	}{
		{
			name:      "target block",
			content:   "target \"webapp\" {}",
			line:      0,
			character: 0,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**target** _Block_\n\n" + parser.BakeSchema.Blocks["target"].Description.Value,
				},
			},
		},
		{
			name:      "target block with 10 preceding newlines",
			content:   "\n\n\n\n\n\n\n\n\n\ntarget \"webapp\" {}",
			line:      10,
			character: 0,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**target** _Block_\n\n" + parser.BakeSchema.Blocks["target"].Description.Value,
				},
			},
		},
		{
			name:      "target block with 10 preceding CRLFs",
			content:   "\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\ntarget \"webapp\" {}",
			line:      10,
			character: 0,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**target** _Block_\n\n" + parser.BakeSchema.Blocks["target"].Description.Value,
				},
			},
		},
		{
			name:      "args attribute (inside a target block)",
			content:   "target \"default\" {\n  args = {\n    VERSION = \"0.0.0+unknown\"\n  }\n}",
			line:      1,
			character: 4,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**args** _optional, map of string_\n\n" + parser.BakeSchema.Blocks["target"].Body.Attributes["args"].Description.Value,
				},
			},
		},
		{
			name:      "target attribute (inside a target block)",
			content:   "target \"default\" {\n  target = \"binaries\"\n}",
			line:      1,
			character: 4,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**target** _optional, string_\n\n" + parser.BakeSchema.Blocks["target"].Body.Attributes["target"].Description.Value,
				},
			},
		},
		{
			name:      "${variable} inside tags",
			content:   "target \"api\" {\n  tags = [\"${variable}\"]\n}",
			line:      1,
			character: 17,
			result:    nil,
		},
		{
			name:      "whitespace between attribute and value",
			content:   "target \"api\" {\n  tags  = [\"${variable}\"]\n}",
			line:      1,
			character: 7,
			result:    nil,
		},
		{
			name:      "variable lookup from the unquoted declaration",
			content:   "variable var { default = \"value\" }",
			line:      0,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from the quoted declaration",
			content:   "variable \"var\" { default = \"value\" }",
			line:      0,
			character: 12,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from before the quoted declaration",
			content:   "variable \"var\" { default = \"value\" }",
			line:      0,
			character: 9,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"var\" (variableName)",
				},
			},
		},
		{
			name:      "variable lookup from the reference",
			content:   "variable var { default = \"value\" }\ntarget t { name = var }",
			line:      1,
			character: 19,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from the reference",
			content:   "variable {}\nvariable var { default = \"value\" }\ntarget t { name = var }",
			line:      2,
			character: 19,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from a single templated reference",
			content:   "variable var { default = \"value\" }\ntarget t { name = \"${var}\" }",
			line:      1,
			character: 24,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from templated references within a quoted string (first reference)",
			content:   "variable var { default = \"value\" }\ntarget t { name = \"${var}${var}\" }",
			line:      1,
			character: 24,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from templated references within a quoted string (second reference)",
			content:   "variable var { default = \"value\" }\ntarget t { name = \"${var}${var}\" }",
			line:      1,
			character: 30,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup from a reference in an array",
			content:   "variable var { default = \"value\" }\ntarget t { annotations = [ var ] }",
			line:      1,
			character: 29,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "variable lookup for an attribute's value nested within a map",
			content:   "variable var { default = \"value\" }\ntarget t { args = { name = var } }",
			line:      1,
			character: 29,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "\"value\"",
				},
			},
		},
		{
			name:      "hover inside a variable with no label",
			content:   "variable {  }",
			line:      0,
			character: 11,
			result:    nil,
		},
	}

	temporaryBakeFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "docker-bake.hcl")), "/"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument(uri.URI(temporaryBakeFile), 1, []byte(tc.content))
			result, err := Hover(context.Background(), &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: temporaryBakeFile},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, doc)
			require.NoError(t, err)
			if tc.result == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Nil(t, result.Range)
				markupContent, ok := result.Contents.(protocol.MarkupContent)
				require.True(t, ok)
				require.Equal(t, protocol.MarkupKindMarkdown, markupContent.Kind)
				require.Equal(t, tc.result.Contents.(protocol.MarkupContent).Value, markupContent.Value)
			}
		})
	}

}
