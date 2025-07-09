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

func TestDefinition_Services(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range serviceReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Networks(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range networkReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Volumes(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range volumeReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Configs(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range configReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Secrets(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range secretReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Models(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range modelReferenceTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
		})
	}
}

func TestDefinition_Fragments(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range fragmentTestCases {
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
			require.Equal(t, tc.locations(composeFileURI), locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			links, err := Definition(context.Background(), true, doc, &params)
			require.NoError(t, err)
			require.Equal(t, tc.links(composeFileURI), links)
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
		{
			name: "dependent file is not defined properly",
			content: `
include:
  - compose.other.yaml
services:
  test:
    networks:
      - other`,
			otherContent: `
networks: string`,
			line:      6,
			character: 11,
			locations: nil,
			links:     nil,
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
