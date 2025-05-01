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
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func createIncludeLink(u *url.URL, node *token.Token) *protocol.DocumentLink {
	abs, err := types.AbsolutePath(u, node.Value)
	if err != nil {
		return nil
	}
	offset := 0
	if node.Type == token.DoubleQuoteType {
		offset = 1
	}
	return &protocol.DocumentLink{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(node.Position.Line - 1),
				Character: protocol.UInteger(node.Position.Column + offset - 1),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(node.Position.Line - 1),
				Character: protocol.UInteger(node.Position.Column + offset + len(node.Value) - 1),
			},
		},
		Target:  types.CreateStringPointer(protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(abs), "/")))),
		Tooltip: types.CreateStringPointer(abs),
	}
}

func createImageLinks(serviceNode *ast.MappingValueNode) *protocol.DocumentLink {
	if service, ok := serviceNode.Value.(*ast.MappingValueNode); ok {
		if service.Key.GetToken().Value == "image" {
			value := service.Value.GetToken().Value
			linkedText, link := extractImageLink(value)
			return &protocol.DocumentLink{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      protocol.UInteger(service.Value.GetToken().Position.Line) - 1,
						Character: protocol.UInteger(service.Value.GetToken().Position.Column) - 1,
					},
					End: protocol.Position{
						Line:      protocol.UInteger(service.Value.GetToken().Position.Line) - 1,
						Character: protocol.UInteger(service.Value.GetToken().Position.Column - 1 + len(linkedText)),
					},
				},
				Target:  types.CreateStringPointer(link),
				Tooltip: types.CreateStringPointer(link),
			}
		}
	}
	return nil
}

func scanForLinks(u *url.URL, n *ast.MappingValueNode) []protocol.DocumentLink {
	if s, ok := n.Key.(*ast.StringNode); ok {
		links := []protocol.DocumentLink{}
		switch s.Value {
		case "include":
			if sequence, ok := n.Value.(*ast.SequenceNode); ok {
				for _, entry := range sequence.Values {
					if mappingNode, ok := entry.(*ast.MappingValueNode); ok {
						if mappingNode.Key.GetToken().Value != "path" {
							continue
						}
						entry = mappingNode.Value
					}
					if sequenceNode, ok := entry.(*ast.SequenceNode); ok {
						for _, entry := range sequenceNode.Values {
							link := createIncludeLink(u, entry.GetToken())
							if link != nil {
								links = append(links, *link)
							}
						}
					} else {
						link := createIncludeLink(u, entry.GetToken())
						if link != nil {
							links = append(links, *link)
						}
					}
				}
			}
		case "services":
			if mappingNode, ok := n.Value.(*ast.MappingNode); ok {
				for _, node := range mappingNode.Values {
					link := createImageLinks(node)
					if link != nil {
						links = append(links, *link)
					}
				}
			} else if serviceNode, ok := n.Value.(*ast.MappingValueNode); ok {
				link := createImageLinks(serviceNode)
				if link != nil {
					links = append(links, *link)
				}
			}
		}
		return links
	}
	return nil
}

func DocumentLink(ctx context.Context, documentURI protocol.URI, doc document.ComposeDocument) ([]protocol.DocumentLink, error) {
	url, err := url.Parse(string(documentURI))
	if err != nil {
		return nil, fmt.Errorf("LSP client sent invalid URI: %v", string(documentURI))
	}

	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	links := []protocol.DocumentLink{}
	for _, documentNode := range file.Docs {
		if mappingNode, ok := documentNode.Body.(*ast.MappingNode); ok {
			for _, node := range mappingNode.Values {
				links = append(links, scanForLinks(url, node)...)
			}
		} else if mappingNodeValue, ok := documentNode.Body.(*ast.MappingValueNode); ok {
			links = append(links, scanForLinks(url, mappingNodeValue)...)
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
