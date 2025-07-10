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

func createRange(t *token.Token, length int) protocol.Range {
	offset := 0
	if t.Type == token.DoubleQuoteType {
		offset = 1
	}
	return protocol.Range{
		Start: protocol.Position{
			Line:      protocol.UInteger(t.Position.Line - 1),
			Character: protocol.UInteger(t.Position.Column + offset - 1),
		},
		End: protocol.Position{
			Line:      protocol.UInteger(t.Position.Line - 1),
			Character: protocol.UInteger(t.Position.Column + offset + length - 1),
		},
	}
}

func createLink(u *url.URL, node *token.Token) *protocol.DocumentLink {
	file := node.Value
	abs, err := types.AbsolutePath(u, file)
	if err == nil {
		abs = filepath.ToSlash(abs)
		return &protocol.DocumentLink{
			Range:   createRange(node, len(file)),
			Target:  types.CreateStringPointer(protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(abs, "/")))),
			Tooltip: types.CreateStringPointer(abs),
		}
	}
	return nil
}

func createFileLink(u *url.URL, serviceNode *ast.MappingValueNode) *protocol.DocumentLink {
	attributeValue := stringNode(serviceNode.Value)
	if attributeValue != nil {
		return createLink(u, attributeValue.GetToken())
	}
	return nil
}

func stringNode(value ast.Node) *ast.StringNode {
	if s, ok := resolveAnchor(value).(*ast.StringNode); ok {
		return s
	}
	return nil
}

func createdNestedLink(u *url.URL, serviceNode *ast.MappingValueNode, parent, child string) *protocol.DocumentLink {
	if resolveAnchor(serviceNode.Key).GetToken().Value == parent {
		if mappingNode, ok := resolveAnchor(serviceNode.Value).(*ast.MappingNode); ok {
			for _, buildAttribute := range mappingNode.Values {
				if resolveAnchor(buildAttribute.Key).GetToken().Value == child {
					return createFileLink(u, buildAttribute)
				}
			}
		}
	}
	return nil
}

func createImageLink(serviceNode *ast.MappingValueNode) *protocol.DocumentLink {
	if resolveAnchor(serviceNode.Key).GetToken().Value == "image" {
		service := stringNode(serviceNode.Value)
		if service != nil {
			linkedText, link := extractImageLink(service.Value)
			if linkedText != "" {
				return &protocol.DocumentLink{
					Range:   createRange(service.GetToken(), len(linkedText)),
					Target:  types.CreateStringPointer(link),
					Tooltip: types.CreateStringPointer(link),
				}
			}
		}
	}
	return nil
}

func createObjectFileLink(u *url.URL, serviceNode *ast.MappingValueNode) *protocol.DocumentLink {
	if resolveAnchor(serviceNode.Key).GetToken().Value == "file" {
		return createFileLink(u, serviceNode)
	}
	return nil
}

func includedFiles(nodes []ast.Node) []*token.Token {
	tokens := []*token.Token{}
	for _, entry := range nodes {
		if mappingNode, ok := resolveAnchor(entry).(*ast.MappingNode); ok {
			for _, value := range mappingNode.Values {
				if resolveAnchor(value.Key).GetToken().Value == "path" {
					if paths, ok := resolveAnchor(value.Value).(*ast.SequenceNode); ok {
						// include:
						//   - path:
						//     - ../commons/compose.yaml
						//     - ./commons-override.yaml
						for _, path := range paths.Values {
							tokens = append(tokens, resolveAnchor(path).GetToken())
						}
					} else {
						// include:
						// - path: ../commons/compose.yaml
						//   project_directory: ..
						//   env_file: ../another/.env
						tokens = append(tokens, resolveAnchor(value.Value).GetToken())
					}
				}
			}
		} else {
			// include:
			//   - abc.yml
			//   - def.yml
			stringNode := stringNode(entry)
			if stringNode != nil {
				tokens = append(tokens, stringNode.GetToken())
			}

		}
	}
	return tokens
}

func scanForLinks(u *url.URL, n *ast.MappingValueNode) []protocol.DocumentLink {
	if s, ok := n.Key.(*ast.StringNode); ok {
		links := []protocol.DocumentLink{}
		switch s.Value {
		case "include":
			if sequence, ok := resolveAnchor(n.Value).(*ast.SequenceNode); ok {
				for _, token := range includedFiles(sequence.Values) {
					link := createLink(u, token)
					if link != nil {
						links = append(links, *link)
					}
				}
			}
		case "services":
			if mappingNode, ok := resolveAnchor(n.Value).(*ast.MappingNode); ok {
				for _, node := range mappingNode.Values {
					if serviceAttributes, ok := resolveAnchor(node.Value).(*ast.MappingNode); ok {
						for _, serviceAttribute := range serviceAttributes.Values {
							link := createImageLink(serviceAttribute)
							if link != nil {
								links = append(links, *link)
							}

							link = createdNestedLink(u, serviceAttribute, "build", "dockerfile")
							if link != nil {
								links = append(links, *link)
							}

							link = createdNestedLink(u, serviceAttribute, "credential_spec", "file")
							if link != nil {
								links = append(links, *link)
							}
						}
					}
				}
			}
		case "configs":
			if mappingNode, ok := resolveAnchor(n.Value).(*ast.MappingNode); ok {
				for _, node := range mappingNode.Values {
					if configAttributes, ok := resolveAnchor(node.Value).(*ast.MappingNode); ok {
						for _, configAttribute := range configAttributes.Values {
							link := createObjectFileLink(u, configAttribute)
							if link != nil {
								links = append(links, *link)
							}
						}
					}
				}
			}
		case "secrets":
			if mappingNode, ok := resolveAnchor(n.Value).(*ast.MappingNode); ok {
				for _, node := range mappingNode.Values {
					if configAttributes, ok := resolveAnchor(node.Value).(*ast.MappingNode); ok {
						for _, configAttribute := range configAttributes.Values {
							link := createObjectFileLink(u, configAttribute)
							if link != nil {
								links = append(links, *link)
							}
						}
					}
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
		}
	}
	return links, nil
}

func extractImageLink(nodeValue string) (string, string) {
	if strings.HasPrefix(nodeValue, "ghcr.io") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			if len(nodeValue) <= 8 {
				return "", ""
			}
			return nodeValue, fmt.Sprintf("https://%v", nodeValue)
		}
		return nodeValue[0:idx], fmt.Sprintf("https://%v", nodeValue[0:idx])
	}

	if strings.HasPrefix(nodeValue, "mcr.microsoft.com") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			if len(nodeValue) <= 18 {
				return "", ""
			}
			return nodeValue, fmt.Sprintf("https://mcr.microsoft.com/artifact/mar/%v", nodeValue[18:])
		}
		return nodeValue[0:idx], fmt.Sprintf("https://mcr.microsoft.com/artifact/mar/%v", nodeValue[18:idx])
	}

	if strings.HasPrefix(nodeValue, "quay.io") {
		idx := strings.LastIndex(nodeValue, ":")
		if idx == -1 {
			if len(nodeValue) <= 8 {
				return "", ""
			}
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
