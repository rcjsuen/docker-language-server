package compose

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func Hover(ctx context.Context, params *protocol.HoverParams, doc document.ComposeDocument) (*protocol.Hover, error) {
	line := int(params.Position.Line) + 1
	root := doc.RootNode()
	if len(root.Content) > 0 {
		lines := strings.Split(string(doc.Input()), "\n")
		character := int(params.Position.Character) + 1
		topLevel, _, _ := NodeStructure(line, root.Content[0].Content)
		return hoverLookup(composeSchema, topLevel, character, len(lines[params.Position.Line])+1), nil
	}
	return nil, nil
}

func hoverLookup(schema *jsonschema.Schema, nodes []*yaml.Node, column, lineLength int) *protocol.Hover {
	for _, node := range nodes {
		if schema.Ref != nil {
			schema = schema.Ref
		}

		if nested, ok := schema.Items.(*jsonschema.Schema); ok {
			for _, n := range nested.OneOf {
				if n.Types != nil && slices.Contains(n.Types.ToStrings(), "object") {
					if len(n.Properties) > 0 {
						if _, ok := n.Properties[node.Value]; ok {
							schema = n
							break
						}
					}
				}
			}

			if _, ok := nested.Properties[node.Value]; ok {
				schema = nested
			}
		}

		if property, ok := schema.Properties[node.Value]; ok {
			if property.Enum != nil {
				if node.Column <= column && column <= lineLength {
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

			if node.Column+len(node.Value) >= column && property.Description != "" {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.MarkupKindPlainText,
						Value: property.Description,
					},
				}
			}
			schema = property
		}

		for regexp, property := range schema.PatternProperties {
			if regexp.MatchString(node.Value) {
				if property.Ref == nil {
					schema = property
				} else {
					schema = property.Ref
				}
				break
			}
		}

		for _, nested := range schema.OneOf {
			if nested.Types != nil && slices.Contains(nested.Types.ToStrings(), "object") {
				schema = nested
				break
			}
		}
	}
	return nil
}
