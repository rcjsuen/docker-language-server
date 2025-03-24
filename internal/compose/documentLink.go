package compose

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
)

func DocumentLink(ctx context.Context, documentURI protocol.URI, document document.ComposeDocument) ([]protocol.DocumentLink, error) {
	url, err := url.Parse(string(documentURI))
	if err != nil {
		return nil, fmt.Errorf("LSP client sent invalid URI: %v", string(documentURI))
	}

	links := []protocol.DocumentLink{}
	results, _ := DocumentSymbol(ctx, document)
	for _, result := range results {
		if symbol, ok := result.(*protocol.DocumentSymbol); ok && symbol.Kind == protocol.SymbolKindModule {
			abs, err := types.AbsolutePath(url, symbol.Name)
			if err == nil {
				links = append(links, protocol.DocumentLink{
					Range:   symbol.Range,
					Target:  types.CreateStringPointer(protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(abs), "/")))),
					Tooltip: types.CreateStringPointer(abs),
				})
			}
		}
	}
	return links, nil
}
