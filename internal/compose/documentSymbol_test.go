package compose

import (
	"context"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
)

func TestDocumentSymbol(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		symbols []*protocol.DocumentSymbol
	}{
		{
			name:    "empty file",
			content: "",
			symbols: []*protocol.DocumentSymbol{},
		},
		{
			name:    "empty services block",
			content: "services:",
			symbols: []*protocol.DocumentSymbol{},
		},
		{
			name: "services block",
			content: `services:
  web:
    build: .
  redis:
    image: "redis:alpine"`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "web",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
				{
					Name: "redis",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 2},
						End:   protocol.Position{Line: 3, Character: 7},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 2},
						End:   protocol.Position{Line: 3, Character: 7},
					},
				},
			},
		},
		{
			name: "duplicated services block",
			content: `services:
  web:
    build: .
services:
  redis:
    image: "redis:alpine"`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "web",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
				{
					Name: "redis",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 2},
						End:   protocol.Position{Line: 4, Character: 7},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 2},
						End:   protocol.Position{Line: 4, Character: 7},
					},
				},
			},
		},
		{
			name: "services block with a piped scalar value",
			content: `services:
  web: |
    this is a string`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "web",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 5},
					},
				},
			},
		},
		{
			name: "networks block",
			content: `networks:
  frontend:`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "frontend",
					Kind: protocol.SymbolKindInterface,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 10},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 10},
					},
				},
			},
		},
		{
			name: "volumes block",
			content: `volumes:
  myapp:`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "myapp",
					Kind: protocol.SymbolKindFile,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 7},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 7},
					},
				},
			},
		},
		{
			name: "configs block",
			content: `configs:
  http_config:`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "http_config",
					Kind: protocol.SymbolKindVariable,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 13},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 13},
					},
				},
			},
		},
		{
			name: "secrets block",
			content: `secrets:
  server-certificate:`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "server-certificate",
					Kind: protocol.SymbolKindKey,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 20},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 20},
					},
				},
			},
		},
		{
			name: "include array",
			content: `include:
  - file.yml`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "file.yml",
					Kind: protocol.SymbolKindModule,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 12},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 12},
					},
				},
			},
		},
		{
			name: "include array, path with list of strings",
			content: `include:
  - path:
    - ../commons/compose.yaml
    - ./commons-override.yaml`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "../commons/compose.yaml",
					Kind: protocol.SymbolKindModule,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 6},
						End:   protocol.Position{Line: 2, Character: 29},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 6},
						End:   protocol.Position{Line: 2, Character: 29},
					},
				},
				{
					Name: "./commons-override.yaml",
					Kind: protocol.SymbolKindModule,
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 6},
						End:   protocol.Position{Line: 3, Character: 29},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 6},
						End:   protocol.Position{Line: 3, Character: 29},
					},
				},
			},
		},
		{
			name: "include array, wrong name with list of strings",
			content: `include:
  - path2:
    - ../commons/compose.yaml
    - ./commons-override.yaml`,
			symbols: []*protocol.DocumentSymbol{},
		},
		{
			name: "include array, long syntax",
			content: `include:
  - path: ../commons/compose.yaml
    project_directory: ..
    env_file: ../another/.env`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "../commons/compose.yaml",
					Kind: protocol.SymbolKindModule,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 33},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 33},
					},
				},
			},
		},
		{
			name: "regular file",
			content: `
services:
  web:
    build: .
  redis:
    image: redis

networks:
  testNetwork:`,
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "web",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 5},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 2},
						End:   protocol.Position{Line: 2, Character: 5},
					},
				},
				{
					Name: "redis",
					Kind: protocol.SymbolKindClass,
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 2},
						End:   protocol.Position{Line: 4, Character: 7},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 2},
						End:   protocol.Position{Line: 4, Character: 7},
					},
				},
				{
					Name: "testNetwork",
					Kind: protocol.SymbolKindInterface,
					Range: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 13},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 2},
						End:   protocol.Position{Line: 8, Character: 13},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument("docker-compose.yml", 1, []byte(tc.content))
			symbols, err := DocumentSymbol(context.Background(), doc)
			require.NoError(t, err)
			var result []any
			for _, symbol := range tc.symbols {
				result = append(result, symbol)
			}
			require.Equal(t, result, symbols)
		})
	}
}
