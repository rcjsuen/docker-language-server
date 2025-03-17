package hcl

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
			name:    "targets block",
			content: "target \"webapp\" {\n}\ntarget \"api\" {\n}\n",
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "webapp",
					Kind: protocol.SymbolKindFunction,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
				{
					Name: "api",
					Kind: protocol.SymbolKindFunction,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 0},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 0},
					},
				},
			},
		},
		{
			name:    "target block without name",
			content: "target{}",
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "target",
					Kind: protocol.SymbolKindFunction,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:    "variable block without name",
			content: "variable{}",
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "variable",
					Kind: protocol.SymbolKindVariable,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:    "attribute with a value",
			content: "attribute = \"value\"",
			symbols: []*protocol.DocumentSymbol{
				{
					Name: "attribute",
					Kind: protocol.SymbolKindProperty,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			symbols, err := DocumentSymbol(context.Background(), "docker-bake.hcl", doc)
			require.NoError(t, err)
			var result []any
			for _, symbol := range tc.symbols {
				result = append(result, symbol)
			}
			require.Equal(t, result, symbols)
		})
	}
}
