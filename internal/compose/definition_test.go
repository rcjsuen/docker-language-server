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
			name: "short syntax quoted form of depends_on in services",
			content: `
services:
  web:
    build: .
    depends_on:
      - "redis"
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
						Start: protocol.Position{Line: 5, Character: 9},
						End:   protocol.Position{Line: 5, Character: 14},
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
		{
			name: "long syntax form of depends_on in services",
			content: `
services:
  web:
    build: .
    depends_on:
      db:
        condition: service_healthy
        restart: true
      redis:
        condition: service_started
  db:
    image: postgres
  redis:
    image: redis`,
			line:      8,
			character: 9,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 8, Character: 6},
						End:   protocol.Position{Line: 8, Character: 11},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 7},
					},
				},
			},
		},
		{
			name: "configs reference",
			content: `
services:
  test:
    image: alpine:3.21
    depends_on:
      - redis
    configs:
      - def
  redis:
    image: redis

configs:
  def:
    file: ./httpd.conf`,
			line:      7,
			character: 10,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 5},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 11},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 5},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 12, Character: 2},
						End:   protocol.Position{Line: 12, Character: 5},
					},
				},
			},
		},
		{
			name: "networks reference",
			content: `
services:
  test:
    image: alpine:3.21
    networks:
    - abc

networks:
  abc:`,
			line:      5,
			character: 8,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 5},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 5, Character: 6},
						End:   protocol.Position{Line: 5, Character: 9},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 5},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 5},
					},
				},
			},
		},
		{
			name: "secrets reference",
			content: `
services:
  test:
    image: alpine:3.21
    secrets:
    - abcd

secrets:
  abcd:
    environment: "PATH"`,
			line:      5,
			character: 8,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 6},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 5, Character: 6},
						End:   protocol.Position{Line: 5, Character: 10},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 6},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 6},
					},
				},
			},
		},
		{
			name: "volumes array item reference without a mount path",
			content: `
services:
  test:
    volumes:
      - test2
volumes:
  test2:`,
			line:      4,
			character: 10,
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
						Start: protocol.Position{Line: 4, Character: 8},
						End:   protocol.Position{Line: 4, Character: 13},
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
			name: "volumes array item reference with a mount path",
			content: `
services:
  test:
    volumes:
      - test2:/mount/path
volumes:
  test2:`,
			line:      4,
			character: 10,
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
						Start: protocol.Position{Line: 4, Character: 8},
						End:   protocol.Position{Line: 4, Character: 13},
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
			name: "volumes array item reference on the mount path itself",
			content: `
services:
  test:
    volumes:
      - test2:/mount/path
volumes:
  test2:`,
			line:      4,
			character: 18,
			locations: nil,
			links:     nil,
		},
		{
			name: "volume array item object's with the source attribute",
			content: `
services:
  test:
    volumes:
      - source: test2
volumes:
  test2:`,
			line:      4,
			character: 18,
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
						Start: protocol.Position{Line: 4, Character: 16},
						End:   protocol.Position{Line: 4, Character: 21},
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
			name: "extends as a string attribute",
			content: `
services:
  test:
    image: alpine
  test2:
    extends: test`,
			line:      5,
			character: 15,
			locations: []protocol.Location{
				{
					URI: composeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 6},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 5, Character: 13},
						End:   protocol.Position{Line: 5, Character: 17},
					},
					TargetURI: composeFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 6},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 6},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		u := uri.URI(composeFileURI)
		doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
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
func TestDefinition_ExternalReference(t *testing.T) {
	folder := os.TempDir()
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(folder, "compose.yaml")), "/"))
	otherFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(folder, "compose.other.yaml")), "/"))

	testCases := []struct {
		name         string
		content      string
		otherContent string
		line         uint32
		character    uint32
		locations    any
		links        any
	}{
		{
			name: "dependent service in another file",
			content: `
include:
  - compose.other.yaml
services:
  web:
    build: .
    depends_on:
      - redis`,
			otherContent: `
services:
  redis:
    image: redis:alpine`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 13},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
		},
		{
			name: "dependent network in another file",
			content: `
include:
  - compose.other.yaml
services:
  test:
    image: alpine:3.21
    networks:
      - otherNetwork`,
			otherContent: `
networks:
  otherNetwork:`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 14},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 20},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 14},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 14},
					},
				},
			},
		},
		{
			name: "dependent volume in another file",
			content: `
include:
  - compose.other.yaml
services:
  test:
    image: alpine:3.21
    volumes:
      - other:/mount/stuff`,
			otherContent: `
volumes:
  other:`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 13},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
		},
		{
			name: "dependent config in another file",
			content: `
include:
  - compose.other.yaml
services:
  test:
    image: alpine:3.21
    configs:
      - other`,
			otherContent: `
configs:
  other:
    file: ./httpd.conf`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 13},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
		},
		{
			name: "dependent secret in another file",
			content: `
include:
  - compose.other.yaml
services:
  test:
    image: alpine:3.21
    secrets:
      - other`,
			otherContent: `
secrets:
  other:
    file: ./server.cert`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 13},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
		},
		{
			name: "dependent type has the same name as something else",
			content: `
include:
  - compose.other.yaml
services:
  test:
    image: alpine:3.21
    networks:
      - other`,
			otherContent: `
networks:
  other:
services:
  other:
    image: alpine:3.21
`,
			line:      7,
			character: 11,
			locations: []protocol.Location{
				{
					URI: otherFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 7, Character: 8},
						End:   protocol.Position{Line: 7, Character: 13},
					},
					TargetURI: otherFileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 7},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		u := uri.URI(composeFileURI)
		mgr := document.NewDocumentManager()
		changed, err := mgr.Write(context.Background(), uri.URI(otherFileURI), protocol.DockerComposeLanguage, 1, []byte(tc.otherContent))
		require.NoError(t, err)
		require.True(t, changed)
		doc := document.NewComposeDocument(mgr, u, 1, []byte(tc.content))
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
