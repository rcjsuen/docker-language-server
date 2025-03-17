package hcl

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl/v2"
)

func DocumentSymbol(ctx context.Context, filename string, doc document.BakeHCLDocument) (result []any, err error) {
	symbols, err := doc.Decoder().SymbolsInFile(filename)
	if err != nil {
		return nil, err
	}
	for _, symbol := range symbols {
		symbolRange := symbol.Range()
		if blockSymbol, ok := symbol.(*decoder.BlockSymbol); ok {
			name := blockSymbol.Type
			if len(blockSymbol.Labels) > 0 {
				name = blockSymbol.Labels[0]
			}
			kind := protocol.SymbolKindFunction
			if blockSymbol.Type == "variable" {
				kind = protocol.SymbolKindVariable
			}
			result = append(result, createSymbol(name, kind, symbolRange))
		} else if symbol, ok := symbol.(*decoder.AttributeSymbol); ok {
			result = append(result, createSymbol(symbol.AttrName, protocol.SymbolKindProperty, symbolRange))
		}
	}
	return result, nil
}

func createSymbol(name string, kind protocol.SymbolKind, rng hcl.Range) *protocol.DocumentSymbol {
	return &protocol.DocumentSymbol{
		Name: name,
		Kind: kind,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(rng.Start.Line - 1),
				Character: 0,
			},
			End: protocol.Position{
				Line:      uint32(rng.Start.Line - 1),
				Character: 0,
			},
		},
		SelectionRange: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(rng.Start.Line - 1),
				Character: 0,
			},
			End: protocol.Position{
				Line:      uint32(rng.Start.Line - 1),
				Character: 0,
			},
		},
	}
}
