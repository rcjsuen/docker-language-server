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

func TestDocumentHighlight_Services(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a service",
			content: `
services:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined service array item",
			content: `
services:
  test:
    depends_on:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "cursor not on anything meaningful",
			content: `
services:
  test:
    depends_on:
      - test2
      - test2`,
			line:      3,
			character: 9,
			ranges:    nil,
		},
		{
			name: "read highlight on an undefined service array item, duplicated",
			content: `
services:
  test:
    depends_on:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a service array item",
			content: `
services:
  test:
    depends_on:
      - test2
  test2:`,
			line:      5,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 2, 5, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}
