package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestFormatting(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		edits   []protocol.TextEdit
	}{
		{
			name: "correct indentation of one level indented more than expected",
			content: `
services:
   web:`,
			edits: []protocol.TextEdit{
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 3},
					},
				},
			},
		},
		{
			name: "correct indentation of the second level indented less than expected",
			content: `
services:
  web:
   image: alpine`,
			edits: []protocol.TextEdit{
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 3},
					},
				},
			},
		},
		{
			name: "correct indentation of the second level indented more than expected",
			content: `
services:
  web:
     image: alpine`,
			edits: []protocol.TextEdit{
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 5},
					},
				},
			},
		},
		{
			name: "indentation is reset and can be corrected in future nodes",
			content: `
services:
  web:
    image: alpine

networks:
   testNetwork:`,
			edits: []protocol.TextEdit{
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 0},
						End:   protocol.Position{Line: 6, Character: 3},
					},
				},
			},
		},
		{
			name: "indentation is reset for sub nodes",
			content: `
services:
  web:
   image: alpine
  web2:
      image: alpine`,
			edits: []protocol.TextEdit{
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 3},
					},
				},
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 6},
					},
				},
			},
		},
		{
			name: "comment does not reset stored indentations",
			content: `
topLevelNode:
  attribute:
           # comment
        # comment
   attribute2: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 11},
					},
				},
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 0},
						End:   protocol.Position{Line: 4, Character: 8},
					},
				},
				{
					NewText: "    ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 3},
					},
				},
			},
		},
		{
			name: "comment is formatted even if the node is fine",
			content: `
topLevelNode:
  attribute: true
           # comment
  attribute2: false`,
			edits: []protocol.TextEdit{
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 11},
					},
				},
			},
		},
		{
			name: "correct comment is ignored",
			content: `
topLevelNode:
  attribute: true
  # comment
  attribute2: false`,
			edits: []protocol.TextEdit{},
		},
		{
			name: "comment information is reset",
			content: `
topLevelNode:
    # comment
  attribute: true
topLevelNode2:
      # comment
    attribute: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 4},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 6},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 0},
						End:   protocol.Position{Line: 6, Character: 4},
					},
				},
			},
		},
		{
			name: "comment pushed to the front",
			content: `
topLevelNode:
  attribute: true
  # comment
topLevelNode2:`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 2},
					},
				},
			},
		},
		{
			name: "comments are reset when a comment is pushed to the front",
			content: `
topLevelNode:
  attribute: true
 # comment
topLevelNode2:
    # comment
  attribute: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 1},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 4},
					},
				},
			},
		},
		{
			name: "top node adjusted",
			content: `
# comment
 topLevelNode:
  attribute: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 1},
					},
				},
			},
		},
		{
			name: "document separator does not indent if correct",
			content: `---
# comment
topLevelNode:
  attribute: true`,
			edits: []protocol.TextEdit{},
		},
		{
			name: "top node adjusted even with a document separator",
			content: `---
# comment
 topLevelNode:
  attribute: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 1},
					},
				},
			},
		},
		{
			name: "document separator will reset indentations",
			content: `---
topLevelNode:
  attribute: true
---
  abc:
    def: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 0},
						End:   protocol.Position{Line: 4, Character: 2},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 4},
					},
				},
			},
		},
		{
			name: "comments fixed when encountering a document separator",
			content: `---
topLevelNode:
  # comment
---`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 2},
					},
				},
			},
		},
		{
			name: "comments cleared and reset when encountering a document separator",
			content: `---
topLevelNode:
  # comment
---
first:
    # comment
  second: true`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 2},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 4},
					},
				},
			},
		},
		{
			name: "top level nodes' indentation remains consistent",
			content: `
 topLevelNode:
   attribute: true

 topAgain:
   attribute: false`,
			edits: []protocol.TextEdit{
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 1},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 3},
					},
				},
				{
					NewText: "",
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 0},
						End:   protocol.Position{Line: 4, Character: 1},
					},
				},
				{
					NewText: "  ",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 3},
					},
				},
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			edits, err := Formatting(doc, protocol.FormattingOptions{
				protocol.FormattingOptionTabSize: float64(2),
			})
			require.NoError(t, err)
			require.Equal(t, tc.edits, edits)
		})
	}
}
