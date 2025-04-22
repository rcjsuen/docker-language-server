package compose

import (
	"context"
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

func TestDefinition(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))

	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		locations any
		links     any
	}{
		{
			name: "short syntax form of depends_on in services",
			content: `
services:
  web:
    build: .
    depends_on:
      - redis
  redis:
    image: redis`,
			line:      5,
			character: 11,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 2},
						End:   protocol.Position{Line: 6, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 5, Character: 8},
						End:   protocol.Position{Line: 5, Character: 13},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 2},
						End:   protocol.Position{Line: 6, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 2},
						End:   protocol.Position{Line: 6, Character: 7},
					},
				},
			},
		},
		{
			name: "short syntax form of depends_on in services finding the right match",
			content: `
services:
  web:
    build: .
    depends_on:
      - postgres
      - redis
  postgres:
    image: postgres
  redis:
    image: redis`,
			line:      6,
			character: 11,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 9, Character: 2},
						End:   protocol.Position{Line: 9, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 6, Character: 8},
						End:   protocol.Position{Line: 6, Character: 13},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 9, Character: 2},
						End:   protocol.Position{Line: 9, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 9, Character: 2},
						End:   protocol.Position{Line: 9, Character: 7},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		u := uri.URI(composeFileURI)
		doc := document.NewComposeDocument(u, 1, []byte(tc.content))
		params := protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: composeFileURI},
				Position:     protocol.Position{Line: tc.line, Character: tc.character},
			},
		}

		t.Run(fmt.Sprintf("%v (Location)", tc.name), func(t *testing.T) {
			locations, err := Definition(context.Background(), false, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.locations, locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}
