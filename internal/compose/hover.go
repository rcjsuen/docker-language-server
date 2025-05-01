package compose

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/goccy/go-yaml/ast"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func Hover(ctx context.Context, params *protocol.HoverParams, doc document.ComposeDocument) (*protocol.Hover, error) {
	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	line := int(params.Position.Line) + 1
	character := int(params.Position.Character) + 1
	lines := strings.Split(string(doc.Input()), "\n")

	for _, documentNode := range file.Docs {
		if mappingNode, ok := documentNode.Body.(*ast.MappingNode); ok {
			m := constructNodePath([]ast.Node{}, mappingNode, int(params.Position.Line+1), int(params.Position.Character+1))
			hover := hover(composeSchema, m, line, character, len(lines[params.Position.Line])+1)
			if hover != nil {
				return hover, nil
			}
		}
	}
	return nil, nil
}

func hover(schema *jsonschema.Schema, nodes []ast.Node, line, column, lineLength int) *protocol.Hover {
	for _, match := range nodes {
		if schema.Ref != nil {
			schema = schema.Ref
		}

		if nested, ok := schema.Items.(*jsonschema.Schema); ok {
			for _, n := range nested.OneOf {
				if n.Types != nil && slices.Contains(n.Types.ToStrings(), "object") {
					if len(n.Properties) > 0 {
						if _, ok := n.Properties[match.GetToken().Value]; ok {
							schema = n
							break
						}
					}
				}
			}

			if _, ok := nested.Properties[match.GetToken().Value]; ok {
				schema = nested
			}
		}

		for _, nested := range schema.OneOf {
			if nested.Types != nil && slices.Contains(nested.Types.ToStrings(), "object") {
				schema = nested
				break
			}
		}

		if property, ok := schema.Properties[match.GetToken().Value]; ok {
			if property.Enum != nil {
				if match.GetToken().Position.Column <= column && column <= lineLength {
					var builder bytes.Buffer
					builder.WriteString("Allowed values:\n")
					enumValues := []string{}
					for _, value := range property.Enum.Values {
						enumValues = append(enumValues, fmt.Sprintf("%v", value))
					}
					slices.Sort(enumValues)
					for _, value := range enumValues {
						builder.WriteString(fmt.Sprintf("- `%v`\n", value))
					}
					return &protocol.Hover{
						Contents: protocol.MarkupContent{
							Kind:  protocol.MarkupKindMarkdown,
							Value: builder.String(),
						},
					}
				}
			}

			if match.GetToken().Position.Line == line && match.GetToken().Position.Column+len(match.GetToken().Value) >= column && property.Description != "" {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.MarkupKindPlainText,
						Value: property.Description,
					},
				}
			}
			schema = property
			continue
		}

		for regexp, property := range schema.PatternProperties {
			if regexp.MatchString(match.GetToken().Value) {
				if property.Ref == nil {
					schema = property
				} else {
					schema = property.Ref
				}
				break
			}
		}
	}
	return nil
}

func constructNodePath(matches []ast.Node, node ast.Node, line, col int) []ast.Node {
	switch n := node.(type) {
	case *ast.MappingValueNode:
		if keyNode, ok := n.Key.(*ast.StringNode); ok {
			if m := constructNodePath(matches, n.Key, line, col); m != nil {
				matches = append(matches, m...)
				return matches
			}
			if m := constructNodePath(matches, n.Value, line, col); m != nil {
				matches = append(matches, keyNode)
				matches = append(matches, m...)
				return matches
			}
		}
	case *ast.MappingNode:
		for _, kv := range n.Values {
			if m := constructNodePath(matches, kv, line, col); m != nil {
				matches = append(matches, m...)
				return matches
			}
		}
	case *ast.SequenceNode:
		for _, item := range n.Values {
			if m := constructNodePath(matches, item, line, col); m != nil {
				matches = append(matches, m...)
				return matches
			}
		}
	}

	token := node.GetToken()
	if token.Position.Line == line && token.Position.Column <= col && col <= token.Position.Column+len(token.Value) {
		return []ast.Node{node}
	}
	return nil
}
