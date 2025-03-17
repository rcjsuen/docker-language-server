package compose

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

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
			var abs, tooltip string
			if runtime.GOOS == "windows" {
				abs, err = filepath.Abs(path.Join(url.Path[1:], fmt.Sprintf("../%v", symbol.Name)))
				tooltip = abs
				abs = "/" + filepath.ToSlash(abs)
			} else {
				abs, err = filepath.Abs(path.Join(url.Path, fmt.Sprintf("../%v", symbol.Name)))
				tooltip = abs
			}
			if err == nil {
				links = append(links, protocol.DocumentLink{
					Range:   symbol.Range,
					Target:  types.CreateStringPointer(protocol.URI(fmt.Sprintf("file://%v", abs))),
					Tooltip: types.CreateStringPointer(tooltip),
				})
			}
		}
	}
	return links, nil
}
