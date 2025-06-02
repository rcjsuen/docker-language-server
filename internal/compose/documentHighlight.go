package compose

import (
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

type dependencyReference struct {
	dependencyType     string
	documentHighlights []protocol.DocumentHighlight
}

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
			Type:     t.Type,
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

func findFragments(mappingNode *ast.MappingNode, anchors []*ast.AnchorNode, aliases []*ast.AliasNode) ([]*ast.AnchorNode, []*ast.AliasNode) {
	for _, node := range mappingNode.Values {
		if anchor, ok := node.Value.(*ast.AnchorNode); ok {
			anchors = append(anchors, anchor)
		} else if alias, ok := node.Value.(*ast.AliasNode); ok {
			aliases = append(aliases, alias)
		} else if m, ok := node.Value.(*ast.MappingNode); ok {
			otherAnchors, otherAliases := findFragments(m, []*ast.AnchorNode{}, []*ast.AliasNode{})
			anchors = append(anchors, otherAnchors...)
			aliases = append(aliases, otherAliases...)
		} else if s, ok := node.Value.(*ast.SequenceNode); ok {
			for _, item := range s.Values {
				if anchor, ok := item.(*ast.AnchorNode); ok {
					anchors = append(anchors, anchor)
				} else if alias, ok := item.(*ast.AliasNode); ok {
					aliases = append(aliases, alias)
				} else if m, ok := item.(*ast.MappingNode); ok {
					otherAnchors, otherAliases := findFragments(m, []*ast.AnchorNode{}, []*ast.AliasNode{})
					anchors = append(anchors, otherAnchors...)
					aliases = append(aliases, otherAliases...)
				}
			}
		}
	}
	return anchors, aliases
}

func fragmentName(anchors []*ast.AnchorNode, aliases []*ast.AliasNode, line, character int) *string {
	for i := range anchors {
		if inToken(anchors[i].Name.GetToken(), line, character) {
			return &anchors[i].Name.GetToken().Value
		}
	}
	for i := range aliases {
		if inToken(aliases[i].Value.GetToken(), line, character) {
			return &aliases[i].Value.GetToken().Value
		}
	}
	return nil
}

func fragmentRange(anchors []*ast.AnchorNode, anchorName string, line, character int) (*token.Position, *token.Position) {
	var start *token.Position
	for i := range anchors {
		if anchors[i].Name.GetToken().Value == anchorName {
			p := anchors[i].GetToken().Position
			if p.Line < line || (p.Line == line && p.Column < character) {
				start = p
			} else {
				return start, &token.Position{
					Line:   p.Line,
					Column: p.Column - 1,
				}
			}
		}
	}
	return start, nil
}

func fragmentReference(mappingNode *ast.MappingNode, line, character int) (*ast.AnchorNode, []*ast.AliasNode) {
	anchors, aliases := findFragments(mappingNode, []*ast.AnchorNode{}, []*ast.AliasNode{})
	anchorName := fragmentName(anchors, aliases, line, character)
	if anchorName != nil {
		var anchor *ast.AnchorNode
		matchingAliases := []*ast.AliasNode{}
		startLine, endLine := fragmentRange(anchors, *anchorName, line, character)
		if startLine != nil {
			for i := range anchors {
				p := anchors[i].GetToken().Position
				if p.Line == startLine.Line && p.Column <= startLine.Column {
					// keep iterating so the closest match is always being updated and assigned
					anchor = anchors[i]
				}
			}
			for i := range aliases {
				if aliases[i].Value.GetToken().Value != *anchorName {
					continue
				}
				if endLine == nil {
					p := aliases[i].GetToken().Position
					if (startLine.Line == p.Line && startLine.Column < p.Column) || startLine.Line < p.Line {
						matchingAliases = append(matchingAliases, aliases[i])
					}
				} else {
					p := aliases[i].GetToken().Position
					if startLine.Line < p.Line {
						if p.Line < endLine.Line {
							matchingAliases = append(matchingAliases, aliases[i])
						} else if p.Line == endLine.Line && p.Column < endLine.Column {
							matchingAliases = append(matchingAliases, aliases[i])
						}
					} else if startLine.Line == p.Line {
						if startLine.Column < p.Column {
							if p.Line == endLine.Line && p.Column < endLine.Column {
								matchingAliases = append(matchingAliases, aliases[i])
							} else if p.Line < endLine.Line {
								matchingAliases = append(matchingAliases, aliases[i])
							}
						}
					}
				}
			}
		} else if endLine == nil {
			// anchor not defined anywhere, add all the aliases with the matching name
			for i := range aliases {
				if aliases[i].Value.GetToken().Value == *anchorName {
					matchingAliases = append(matchingAliases, aliases[i])
				}
			}
		} else {
			// valid anchor found but defined after the alias that the cursor is on
			for i := range aliases {
				t := aliases[i].Value.GetToken()
				if t.Value == *anchorName {
					if t.Position.Line < endLine.Line {
						matchingAliases = append(matchingAliases, aliases[i])
					} else if t.Position.Line == endLine.Line && t.Position.Column < endLine.Column {
						matchingAliases = append(matchingAliases, aliases[i])
					}
				}
			}
		}
		return anchor, matchingAliases
	}
	return nil, nil
}

func DocumentHighlight(doc document.ComposeDocument, position protocol.Position) ([]protocol.DocumentHighlight, error) {
	_, references := DocumentHighlights(doc, position)
	if len(references.documentHighlights) == 0 {
		return nil, nil
	}
	return references.documentHighlights, nil
}

func DocumentHighlights(doc document.ComposeDocument, position protocol.Position) (string, dependencyReference) {
	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return "", dependencyReference{documentHighlights: nil}
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
					name, highlights := highlightReferences("services", refs, decls, line, character)
					if len(highlights.documentHighlights) > 0 {
						return name, highlights
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
		name, highlights := highlightReferences("networks", networkRefs, networkDeclarations, line, character)
		if len(highlights.documentHighlights) > 0 {
			return name, highlights
		}
		name, highlights = highlightReferences("volumes", volumeRefs, volumeDeclarations, line, character)
		if len(highlights.documentHighlights) > 0 {
			return name, highlights
		}
		name, highlights = highlightReferences("configs", configRefs, configDeclarations, line, character)
		if len(highlights.documentHighlights) > 0 {
			return name, highlights
		}
		name, highlights = highlightReferences("secrets", secretRefs, secretDeclarations, line, character)
		if len(highlights.documentHighlights) > 0 {
			return name, highlights
		}

		fragments := []protocol.DocumentHighlight{}
		anchor, aliases := fragmentReference(mappingNode, line, character)
		if anchor != nil {
			fragments = append(fragments, documentHighlightFromToken(anchor.Name.GetToken(), protocol.DocumentHighlightKindWrite))
		}
		for i := range aliases {
			fragments = append(fragments, documentHighlightFromToken(aliases[i].Value.GetToken(), protocol.DocumentHighlightKindRead))
		}
		return "", dependencyReference{documentHighlights: fragments}
	}
	return "", dependencyReference{documentHighlights: nil}
}

func highlightReferences(dependencyType string, refs, decls []*token.Token, line, character int) (string, dependencyReference) {
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
		return *highlightedName, dependencyReference{dependencyType: dependencyType, documentHighlights: highlights}
	}
	return "", dependencyReference{documentHighlights: nil}
}

func documentHighlightFromToken(t *token.Token, kind protocol.DocumentHighlightKind) protocol.DocumentHighlight {
	return protocol.DocumentHighlight{
		Kind:  &kind,
		Range: createRange(t, len(t.Value)),
	}
}

func inToken(t *token.Token, line, character int) bool {
	return t.Position.Line == line && t.Position.Column <= character && character <= t.Position.Column+len(t.Value)
}
