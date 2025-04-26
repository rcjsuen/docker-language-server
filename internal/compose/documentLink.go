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

func DocumentLink(ctx context.Context, documentURI protocol.URI, doc document.ComposeDocument) ([]protocol.DocumentLink, error) {
	url, err := url.Parse(string(documentURI))
	if err != nil {
		return nil, fmt.Errorf("LSP client sent invalid URI: %v", string(documentURI))
	}

	links := []protocol.DocumentLink{}
	results, _ := DocumentSymbol(ctx, doc)
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

	root := doc.RootNode()
	if len(root.Content) > 0 {
		for i := range root.Content[0].Content {
			switch root.Content[0].Content[i].Value {
			case "services":
				for j := 0; j < len(root.Content[0].Content[i+1].Content); j += 2 {
					serviceProperties := root.Content[0].Content[i+1].Content[j+1].Content
					for k := 0; k < len(serviceProperties); k += 2 {
						if serviceProperties[k].Value == "image" {
							imageNode := serviceProperties[k+1]
							linkedText, link := extractImageLink(imageNode.Value)
							links = append(links, protocol.DocumentLink{
								Range: protocol.Range{
									Start: protocol.Position{
										Line:      protocol.UInteger(imageNode.Line) - 1,
										Character: protocol.UInteger(imageNode.Column) - 1,
									},
									End: protocol.Position{
										Line:      protocol.UInteger(imageNode.Line) - 1,
										Character: protocol.UInteger(imageNode.Column - 1 + len(linkedText)),
									},
								},
								Target:  types.CreateStringPointer(link),
								Tooltip: types.CreateStringPointer(link),
							})
						}
					}
				}
			}
		}
	}
	return links, nil
}

func extractImageLink(nodeValue string) (string, string) {
	if strings.HasPrefix(nodeValue, "ghcr.io/") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			return nodeValue, fmt.Sprintf("https://%v", nodeValue)
		}
		return nodeValue[0:idx], fmt.Sprintf("https://%v", nodeValue[0:idx])
	}

	if strings.HasPrefix(nodeValue, "mcr.microsoft.com") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			return nodeValue, fmt.Sprintf("https://mcr.microsoft.com/artifact/mar/%v", nodeValue[18:])
		}
		return nodeValue[0:idx], fmt.Sprintf("https://mcr.microsoft.com/artifact/mar/%v", nodeValue[18:idx])
	}

	if strings.HasPrefix(nodeValue, "quay.io/") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			return nodeValue, fmt.Sprintf("https://quay.io/repository/%v", nodeValue[8:])
		}
		return nodeValue[0:idx], fmt.Sprintf("https://quay.io/repository/%v", nodeValue[8:idx])
	}

	idx := strings.LastIndex(nodeValue, ":")
	if idx == -1 {
		idx := strings.Index(nodeValue, "/")
		if idx == -1 {
			return nodeValue, fmt.Sprintf("https://hub.docker.com/_/%v", nodeValue)
		}
		return nodeValue, fmt.Sprintf("https://hub.docker.com/r/%v", nodeValue)
	}

	slashIndex := strings.Index(nodeValue, "/")
	if slashIndex == -1 {
		return nodeValue[0:idx], fmt.Sprintf("https://hub.docker.com/_/%v", nodeValue[0:idx])
	}
	return nodeValue[0:idx], fmt.Sprintf("https://hub.docker.com/r/%v", nodeValue[0:idx])
}
