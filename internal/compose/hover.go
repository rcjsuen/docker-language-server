package compose

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
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
			nodePath := constructNodePath([]ast.Node{}, mappingNode, int(params.Position.Line+1), int(params.Position.Character+1))
			result := serviceHover(doc, mappingNode, nodePath)
			if result != nil {
				return result, nil
			}
			result = hover(composeSchema, nodePath, line, character, len(lines[params.Position.Line])+1)
			if result != nil {
				return result, nil
			}
		}
	}
	return nil, nil
}

func createYamlHover(node *ast.MappingValueNode) *protocol.Hover {
	split := strings.Split(node.String(), "\n")
	skip := -1
	for i := range len(split) {
		if skip == -1 {
			for j := range len(split[i]) {
				if !unicode.IsSpace(rune(split[i][j])) {
					skip = j
					break
				}
			}
		}
		// extra check in case there are empty lines
		if len(split[i]) > skip {
			split[i] = split[i][skip:]
		}
	}
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: fmt.Sprintf("```YAML\n%v\n```", strings.Join(split, "\n")),
		},
	}
}

func createServiceHover(doc document.ComposeDocument, mappingNode *ast.MappingNode, serviceName string) *protocol.Hover {
	for _, node := range mappingNode.Values {
		if s, ok := node.Key.(*ast.StringNode); ok && s.Value == "services" {
			for _, service := range node.Value.(*ast.MappingNode).Values {
				if service.Key.GetToken().Value == serviceName {
					return createYamlHover(service)
				}
			}
		}
	}

	node, _ := dependencyLookup(doc, "services", serviceName)
	if node != nil {
		return createYamlHover(node)
	}
	return nil
}

func serviceHover(doc document.ComposeDocument, mappingNode *ast.MappingNode, nodePath []ast.Node) *protocol.Hover {
	if (len(nodePath) == 4 || len(nodePath) == 5) && nodePath[0].GetToken().Value == "services" {
		if nodePath[2].GetToken().Value == "extends" {
			serviceName := nodePath[3].GetToken().Value
			if len(nodePath) == 5 && nodePath[3].GetToken().Value == "service" {
				if _, ok := nodePath[4].(*ast.StringNode); ok {
					if nodePath[4].GetToken().Next == nil || nodePath[4].GetToken().Next.Type != token.MappingValueType {
						serviceName = nodePath[4].GetToken().Value
					}
				} else {
					return nil
				}
			} else if nodePath[3].GetToken().Next != nil && nodePath[3].GetToken().Next.Type == token.MappingValueType {
				return nil
			}
			result := createServiceHover(doc, mappingNode, serviceName)
			if result != nil {
				return result
			}
		}

		if nodePath[2].GetToken().Value == "depends_on" {
			if nodePath[3].GetToken().Next != nil &&
				nodePath[3].GetToken().Next.Type == token.MappingValueType &&
				nodePath[3].GetToken().Prev.Type == token.SequenceEntryType {
				return nil
			}
			serviceName := nodePath[3].GetToken().Value
			result := createServiceHover(doc, mappingNode, serviceName)
			if result != nil {
				return result
			}
		}
	}
	return nil
}

func hover(schema *jsonschema.Schema, nodes []ast.Node, line, column, lineLength int) *protocol.Hover {
	for _, match := range nodes {
		if schema.Ref != nil {
			schema = schema.Ref
		}

		if nested, ok := schema.Items.(*jsonschema.Schema); ok {
			if nested.Ref != nil {
				nested = nested.Ref
			}
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
					var builder strings.Builder
					if property.Description != "" {
						builder.WriteString(property.Description)
						builder.WriteString("\n\n")
					}
					builder.WriteString("Allowed values:\n")
					enumValues := []string{}
					for _, value := range property.Enum.Values {
						enumValues = append(enumValues, fmt.Sprintf("%v", value))
					}
					slices.Sort(enumValues)
					for _, value := range enumValues {
						builder.WriteString(fmt.Sprintf("- `%v`\n", value))
					}
					builder.WriteString("\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)")
					builder.WriteString(fmt.Sprintf(
						"\n\n[Online documentation](https://docs.docker.com/reference/compose-file/%v/#%v)",
						nodes[0].GetToken().Value,
						nodes[2].GetToken().Value,
					))
					return &protocol.Hover{
						Contents: protocol.MarkupContent{
							Kind:  protocol.MarkupKindMarkdown,
							Value: builder.String(),
						},
					}
				}
			}

			if match.GetToken().Position.Line == line && match.GetToken().Position.Column+len(match.GetToken().Value) >= column && property.Description != "" {
				var builder strings.Builder
				builder.WriteString(property.Description)
				builder.WriteString("\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)")
				switch nodes[0].GetToken().Value {
				case "name":
					builder.WriteString("\n\n[Online documentation](https://docs.docker.com/reference/compose-file/version-and-name/)")
				case "version":
					builder.WriteString("\n\n[Online documentation](https://docs.docker.com/reference/compose-file/version-and-name/)")
				case "include":
					if len(nodes) == 1 {
						builder.WriteString(fmt.Sprintf(
							"\n\n[Online documentation](https://docs.docker.com/reference/compose-file/%v/)",
							nodes[0].GetToken().Value,
						))
					} else {
						builder.WriteString(fmt.Sprintf(
							"\n\n[Online documentation](https://docs.docker.com/reference/compose-file/%v/#%v)",
							nodes[0].GetToken().Value,
							nodes[1].GetToken().Value,
						))
					}
				default:
					if len(nodes) == 1 {
						builder.WriteString(fmt.Sprintf(
							"\n\n[Online documentation](https://docs.docker.com/reference/compose-file/%v/)",
							nodes[0].GetToken().Value,
						))
					} else {
						builder.WriteString(fmt.Sprintf(
							"\n\n[Online documentation](https://docs.docker.com/reference/compose-file/%v/#%v)",
							nodes[0].GetToken().Value,
							nodes[2].GetToken().Value,
						))
					}
				}
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: builder.String(),
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
