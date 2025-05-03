package compose

import (
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func serviceReferences(node *ast.MappingValueNode, dependencyAttributeName string) []*token.Token {
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
						} else if dependentService, ok := attributeNode.Value.(*ast.StringNode); ok {
							tokens = append(tokens, dependentService.GetToken())
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
		for _, node := range mappingNode.Values {
			if s, ok := node.Key.(*ast.StringNode); ok {
				switch s.Value {
				case "services":
					refs := serviceReferences(node, "depends_on")
					decls := declarations(node, "services")
					highlights := highlightServiceReferences(refs, decls, line, character)
					if len(highlights) > 0 {
						return highlights, nil
					}
				}
			}
		}
	}
	return nil, nil
}

func highlightServiceReferences(refs, decls []*token.Token, line, character int) []protocol.DocumentHighlight {
	var match *token.Token
	for _, reference := range refs {
		if inToken(reference, line, character) {
			match = reference
			break
		}
	}

	if match == nil {
		for _, declaration := range decls {
			if inToken(declaration, line, character) {
				match = declaration
				break
			}
		}
	}

	if match != nil {
		highlights := []protocol.DocumentHighlight{}
		for _, reference := range refs {
			if reference.Value == match.Value {
				highlights = append(highlights, documentHighlightFromToken(reference, protocol.DocumentHighlightKindRead))
			}
		}

		for _, declaration := range decls {
			if declaration.Value == match.Value {
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
