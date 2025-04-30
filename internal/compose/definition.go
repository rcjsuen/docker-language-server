package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func findSequenceDependencyToken(attributeNode *ast.MappingValueNode, attributeName string, line, column int) (string, *token.Token) {
	if dependencies, ok := attributeNode.Value.(*ast.SequenceNode); ok {
		for _, dependency := range dependencies.Values {
			dependencyToken := dependency.GetToken()
			if dependencyToken.Position.Line == line && dependencyToken.Position.Column <= column && column <= dependencyToken.Position.Column+len(dependencyToken.Value) {
				return attributeName, dependencyToken
			}
		}
	}
	return "", nil
}

func findDependencyToken(attributeNode *ast.MappingValueNode, attributeName string, line, column int) (string, *token.Token) {
	if attributeNode.Key.GetToken().Value == attributeName {
		return findSequenceDependencyToken(attributeNode, attributeName, line, column)
	}
	return "", nil
}

func lookupReference(serviceNode *ast.MappingValueNode, line, column int) (string, *token.Token) {
	if serviceAttributes, ok := serviceNode.Value.(*ast.MappingNode); ok {
		for _, attributeNode := range serviceAttributes.Values {
			if attributeNode.Key.GetToken().Value == "depends_on" {
				if _, ok := attributeNode.Value.(*ast.SequenceNode); ok {
					reference, dependency := findSequenceDependencyToken(attributeNode, "services", line, column)
					if dependency != nil {
						return reference, dependency
					}
				} else if serviceAttributes, ok := attributeNode.Value.(*ast.MappingNode); ok {
					for _, dependency := range serviceAttributes.Values {
						dependencyToken := dependency.Key.GetToken()
						if dependencyToken.Position.Line == line && dependencyToken.Position.Column <= column && column <= dependencyToken.Position.Column+len(dependencyToken.Value) {
							return "services", dependencyToken
						}
					}
				}
			}

			reference, dependency := findDependencyToken(attributeNode, "configs", line, column)
			if dependency != nil {
				return reference, dependency
			}
			reference, dependency = findDependencyToken(attributeNode, "networks", line, column)
			if dependency != nil {
				return reference, dependency
			}
			reference, dependency = findDependencyToken(attributeNode, "secrets", line, column)
			if dependency != nil {
				return reference, dependency
			}
		}
	}
	return "", nil
}

func lookupDependency(node *ast.MappingValueNode, line, column int) (string, *token.Token) {
	if s, ok := node.Key.(*ast.StringNode); ok && s.Value == "services" {
		if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
			for _, serviceNode := range servicesNode.Values {
				reference, dependency := lookupReference(serviceNode, line, column)
				if dependency != nil {
					return reference, dependency
				}
			}
		} else if valueNode, ok := node.Value.(*ast.MappingValueNode); ok {
			return lookupReference(valueNode, line, column)
		}
	}
	return "", nil
}

func findDefinition(node *ast.MappingValueNode, referenceType, referenceName string) *token.Token {
	if s, ok := node.Key.(*ast.StringNode); ok && s.Value == referenceType {
		if servicesNode, ok := node.Value.(*ast.MappingNode); ok {
			for _, serviceNode := range servicesNode.Values {
				if serviceNode.Key.GetToken().Value == referenceName {
					return serviceNode.Key.GetToken()
				}
			}
		} else if networks, ok := node.Value.(*ast.MappingValueNode); ok {
			if networks.Key.GetToken().Value == referenceName {
				return networks.Key.GetToken()
			}
		}
	}
	return nil
}

func Definition(ctx context.Context, definitionLinkSupport bool, doc document.ComposeDocument, params *protocol.DefinitionParams) (any, error) {
	line := int(params.Position.Line) + 1
	character := int(params.Position.Character) + 1

	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	if mappingNode, ok := file.Docs[0].Body.(*ast.MappingNode); ok {
		for _, node := range mappingNode.Values {
			reference, dependency := lookupDependency(node, line, character)
			if dependency != nil {
				for _, node := range mappingNode.Values {
					referenced := findDefinition(node, reference, dependency.Value)
					if referenced != nil {
						return dependencyLink(definitionLinkSupport, params, referenced, dependency), nil
					}
				}
				return nil, nil
			}
		}
	} else if mappingNodeValue, ok := file.Docs[0].Body.(*ast.MappingValueNode); ok {
		reference, dependency := lookupDependency(mappingNodeValue, line, character)
		if dependency != nil {
			referenced := findDefinition(mappingNodeValue, reference, dependency.Value)
			if referenced != nil {
				return dependencyLink(definitionLinkSupport, params, referenced, dependency), nil
			}
		}
	}
	return nil, nil
}

func dependencyLink(definitionLinkSupport bool, params *protocol.DefinitionParams, referenced, dependency *token.Token) any {
	return types.CreateDefinitionResult(
		definitionLinkSupport,
		protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(referenced.Position.Line - 1),
				Character: protocol.UInteger(referenced.Position.Column - 1),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(referenced.Position.Line - 1),
				Character: protocol.UInteger(referenced.Position.Column + len(referenced.Value) - 1),
			},
		},
		&protocol.Range{
			Start: protocol.Position{
				Line:      params.Position.Line,
				Character: protocol.UInteger(dependency.Position.Column - 1),
			},
			End: protocol.Position{
				Line:      params.Position.Line,
				Character: protocol.UInteger(dependency.Position.Column + len(dependency.Value) - 1),
			},
		},
		params.TextDocument.URI,
	)
}
