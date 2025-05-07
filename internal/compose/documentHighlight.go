package compose

import (
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func serviceDependencyReferences(node *ast.MappingValueNode, dependencyAttributeName string, arrayOnly bool) []*token.Token {
	if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
		tokens := []*token.Token{}
		for _, serviceNode := range servicesNode.Values {
			if serviceAttributes, ok := serviceNode.Value.(*ast.MappingNode); ok {
				for _, attributeNode := range serviceAttributes.Values {
					if attributeNode.Key.GetToken().Value == dependencyAttributeName {
						if sequenceNode, ok := attributeNode.Value.(*ast.SequenceNode); ok {
							for _, service := range sequenceNode.Values {
								tokens = append(tokens, service.GetToken())
							}
						} else if !arrayOnly {
							if mappingNode, ok := attributeNode.Value.(*ast.MappingNode); ok {
								for _, dependentService := range mappingNode.Values {
									tokens = append(tokens, dependentService.Key.GetToken())
								}
							}
						}
					}
				}
			}
		}
		return tokens
	}
	return nil
}

func extendedServiceReferences(node *ast.MappingValueNode) []*token.Token {
	if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
		tokens := []*token.Token{}
		for _, serviceNode := range servicesNode.Values {
			if serviceAttributes, ok := serviceNode.Value.(*ast.MappingNode); ok {
				for _, attributeNode := range serviceAttributes.Values {
					if attributeNode.Key.GetToken().Value == "extends" {
						if extendedValue, ok := attributeNode.Value.(*ast.StringNode); ok {
							tokens = append(tokens, extendedValue.GetToken())
						} else if mappingNode, ok := attributeNode.Value.(*ast.MappingNode); ok {
							localService := true
							for _, extendsObjectAttribute := range mappingNode.Values {
								if extendsObjectAttribute.Key.GetToken().Value == "file" {
									localService = false
									break
								}
							}

							if localService {
								for _, extendsObjectAttribute := range mappingNode.Values {
									if extendsObjectAttribute.Key.GetToken().Value == "service" {
										tokens = append(tokens, extendsObjectAttribute.Value.GetToken())
									}
								}
							}
						}
					}
				}
			}
		}
		return tokens
	}
	return nil
}

func volumeToken(t *token.Token) *token.Token {
	idx := strings.Index(t.Value, ":")
	if idx != -1 {
		return &token.Token{
			Value:    t.Value[0:idx],
			Position: t.Position,
		}
	}
	return t
}

func volumeReferences(node *ast.MappingValueNode) []*token.Token {
	if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
		tokens := []*token.Token{}
		for _, serviceNode := range servicesNode.Values {
			if serviceAttributes, ok := serviceNode.Value.(*ast.MappingNode); ok {
				for _, attributeNode := range serviceAttributes.Values {
					if attributeNode.Key.GetToken().Value == "volumes" {
						if sequenceNode, ok := attributeNode.Value.(*ast.SequenceNode); ok {
							for _, service := range sequenceNode.Values {
								if volumeObjectNode, ok := service.(*ast.MappingNode); ok {
									for _, volumeAttribute := range volumeObjectNode.Values {
										if volumeAttribute.Key.GetToken().Value == "source" {
											tokens = append(tokens, volumeAttribute.Value.GetToken())
										}
									}
								} else {
									tokens = append(tokens, volumeToken(service.GetToken()))
								}
							}
						} else if mappingNode, ok := attributeNode.Value.(*ast.MappingNode); ok {
							for _, dependentService := range mappingNode.Values {
								tokens = append(tokens, dependentService.Key.GetToken())
							}
						}
					}
				}
			}
		}
		return tokens
	}
	return nil
}

func declarations(node *ast.MappingValueNode, dependencyType string) []*token.Token {
	if s, ok := node.Key.(*ast.StringNode); ok && s.Value == dependencyType {
		if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
			tokens := []*token.Token{}
			for _, serviceNode := range servicesNode.Values {
				tokens = append(tokens, serviceNode.Key.GetToken())
			}
			return tokens
		}
	}
	return nil
}

func DocumentHighlight(doc document.ComposeDocument, position protocol.Position) ([]protocol.DocumentHighlight, error) {
	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	line := int(position.Line) + 1
	character := int(position.Character) + 1
	if mappingNode, ok := file.Docs[0].Body.(*ast.MappingNode); ok {
		var networkRefs []*token.Token
		var volumeRefs []*token.Token
		var configRefs []*token.Token
		var secretRefs []*token.Token
		var networkDeclarations []*token.Token
		var volumeDeclarations []*token.Token
		var configDeclarations []*token.Token
		var secretDeclarations []*token.Token
		for _, node := range mappingNode.Values {
			if s, ok := node.Key.(*ast.StringNode); ok {
				switch s.Value {
				case "services":
					refs := serviceDependencyReferences(node, "depends_on", false)
					refs = append(refs, extendedServiceReferences(node)...)
					decls := declarations(node, "services")
					highlights := highlightReferences(refs, decls, line, character)
					if len(highlights) > 0 {
						return highlights, nil
					}
					networkRefs = serviceDependencyReferences(node, "networks", false)
					configRefs = serviceDependencyReferences(node, "configs", true)
					secretRefs = serviceDependencyReferences(node, "secrets", true)
					volumeRefs = volumeReferences(node)
				case "networks":
					networkDeclarations = declarations(node, "networks")
				case "volumes":
					volumeDeclarations = declarations(node, "volumes")
				case "configs":
					configDeclarations = declarations(node, "configs")
				case "secrets":
					secretDeclarations = declarations(node, "secrets")
				}
			}
		}
		highlights := highlightReferences(networkRefs, networkDeclarations, line, character)
		if len(highlights) > 0 {
			return highlights, nil
		}
		highlights = highlightReferences(volumeRefs, volumeDeclarations, line, character)
		if len(highlights) > 0 {
			return highlights, nil
		}
		highlights = highlightReferences(configRefs, configDeclarations, line, character)
		if len(highlights) > 0 {
			return highlights, nil
		}
		highlights = highlightReferences(secretRefs, secretDeclarations, line, character)
		if len(highlights) > 0 {
			return highlights, nil
		}
		return highlights, nil
	}
	return nil, nil
}

func highlightReferences(refs, decls []*token.Token, line, character int) []protocol.DocumentHighlight {
	var highlightedName *string
	for _, reference := range refs {
		if inToken(reference, line, character) {
			highlightedName = &reference.Value
			break
		}
	}

	if highlightedName == nil {
		for _, declaration := range decls {
			if inToken(declaration, line, character) {
				highlightedName = &declaration.Value
				break
			}
		}
	}

	if highlightedName != nil {
		highlights := []protocol.DocumentHighlight{}
		for _, reference := range refs {
			if reference.Value == *highlightedName {
				highlights = append(highlights, documentHighlightFromToken(reference, protocol.DocumentHighlightKindRead))
			}
		}

		for _, declaration := range decls {
			if declaration.Value == *highlightedName {
				highlights = append(highlights, documentHighlightFromToken(declaration, protocol.DocumentHighlightKindWrite))
			}
		}
		return highlights
	}
	return nil
}

func documentHighlightFromToken(t *token.Token, kind protocol.DocumentHighlightKind) protocol.DocumentHighlight {
	return documentHighlight(
		protocol.UInteger(t.Position.Line)-1,
		protocol.UInteger(t.Position.Column)-1,
		protocol.UInteger(t.Position.Line)-1,
		protocol.UInteger(t.Position.Column+len(t.Value))-1,
		kind,
	)
}

func documentHighlight(startLine, startCharacter, endLine, endCharacter protocol.UInteger, kind protocol.DocumentHighlightKind) protocol.DocumentHighlight {
	return protocol.DocumentHighlight{
		Kind: &kind,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      startLine,
				Character: startCharacter,
			},
			End: protocol.Position{
				Line:      endLine,
				Character: endCharacter,
			},
		},
	}
}

func inToken(t *token.Token, line, character int) bool {
	return t.Position.Line == line && t.Position.Column <= character && character <= t.Position.Column+len(t.Value)
}
