package compose

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func Completion(ctx context.Context, params *protocol.CompletionParams, doc document.ComposeDocument) (*protocol.CompletionList, error) {
	if params.Position.Character == 0 {
		items := []protocol.CompletionItem{}
		for attributeName, schema := range schemaProperties() {
			item := protocol.CompletionItem{Label: attributeName}
			if schema.Description != "" {
				item.Documentation = schema.Description
			}
			items = append(items, item)
		}
		slices.SortFunc(items, func(a, b protocol.CompletionItem) int {
			return strings.Compare(a.Label, b.Label)
		})
		return &protocol.CompletionList{Items: items}, nil
	}

	lines := strings.Split(string(doc.Input()), "\n")
	lspLine := int(params.Position.Line)
	if strings.HasPrefix(strings.TrimSpace(lines[lspLine]), "#") {
		return nil, nil
	}

	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	line := int(lspLine) + 1
	character := int(params.Position.Character) + 1
	path := constructCompletionNodePath(file, line)
	if len(path) == 1 {
		return nil, nil
	} else if path[1].Key.GetToken().Position.Column >= character {
		return nil, nil
	}

	items := []protocol.CompletionItem{}
	nodeProps := nodeProperties(path, line, character)
	if schema, ok := nodeProps.(*jsonschema.Schema); ok {
		if schema.Enum != nil {
			for _, value := range schema.Enum.Values {
				item := protocol.CompletionItem{
					Detail: extractDetail(schema),
					Label:  value.(string),
				}
				items = append(items, item)
			}
		}
	} else if properties, ok := nodeProps.(map[string]*jsonschema.Schema); ok {
		sb := strings.Builder{}
		for i := range lines[lspLine] {
			if unicode.IsSpace(rune(lines[lspLine][i])) || lines[lspLine][i] == '-' {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("  ")
		for attributeName, schema := range properties {
			item := protocol.CompletionItem{
				Detail:         extractDetail(schema),
				Label:          attributeName,
				InsertText:     insertText(sb.String(), attributeName, schema),
				InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			}

			if schema.Enum != nil {
				options := []string{}
				for i := range schema.Enum.Values {
					options = append(options, schema.Enum.Values[i].(string))
				}
				slices.Sort(options)
				sb := strings.Builder{}
				sb.WriteString(attributeName)
				sb.WriteString(": ${1|")
				for i := range options {
					sb.WriteString(options[i])
					if i != len(schema.Enum.Values)-1 {
						sb.WriteString(",")
					}
				}
				sb.WriteString("|}")
				item.InsertText = types.CreateStringPointer(sb.String())
				item.InsertTextFormat = types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet)
			}
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, nil
	}
	slices.SortFunc(items, func(a, b protocol.CompletionItem) int {
		return strings.Compare(a.Label, b.Label)
	})
	return &protocol.CompletionList{Items: items}, nil
}

func constructCompletionNodePath(file *ast.File, line int) []*ast.MappingValueNode {
	for _, documentNode := range file.Docs {
		if mappingNode, ok := documentNode.Body.(*ast.MappingNode); ok {
			return NodeStructure(line, mappingNode.Values)
		}
	}
	return nil
}

func NodeStructure(line int, rootNodes []*ast.MappingValueNode) []*ast.MappingValueNode {
	if len(rootNodes) == 0 {
		return nil
	}

	var candidate *ast.MappingValueNode
	for _, node := range rootNodes {
		if node.GetToken().Position.Line < line {
			candidate = node
		} else if node.GetToken().Position.Line == line {
			return []*ast.MappingValueNode{node}
		} else {
			break
		}
	}
	nodes := []*ast.MappingValueNode{candidate}
	candidates := walkNodes(line, candidate)
	nodes = append(nodes, candidates...)
	return nodes
}

func walkNodes(line int, node *ast.MappingValueNode) []*ast.MappingValueNode {
	var candidate ast.Node
	value := node.Value
	if mappingNode, ok := value.(*ast.MappingNode); ok {
		for _, child := range mappingNode.Values {
			if child.GetToken().Position.Line < line {
				candidate = child
			} else if child.GetToken().Position.Line == line {
				candidate = child
				break
			}
		}
	} else if sequenceNode, ok := value.(*ast.SequenceNode); ok {
		for _, child := range sequenceNode.Values {
			if child.GetToken().Position.Line < line {
				if _, ok := child.(*ast.NullNode); ok {
					continue
				}
				candidate = child
			} else if child.GetToken().Position.Line == line {
				if _, ok := child.(*ast.NullNode); ok {
					break
				}
				candidate = child
				break
			}
		}
	}

	if mappingNode, ok := candidate.(*ast.MappingNode); ok {
		for _, child := range mappingNode.Values {
			if child.GetToken().Position.Line < line {
				candidate = child
			} else if child.GetToken().Position.Line == line {
				candidate = child
				break
			}
		}
	}

	if candidate == nil {
		return []*ast.MappingValueNode{}
	}

	if next, ok := candidate.(*ast.MappingValueNode); ok {
		nodes := []*ast.MappingValueNode{next}
		candidates := walkNodes(line, next)
		nodes = append(nodes, candidates...)
		return nodes
	}
	return []*ast.MappingValueNode{}
}

func referencedTypes(schema *jsonschema.Schema) []string {
	if schema.Types != nil {
		return schema.Types.ToStrings()
	} else if schema.Ref != nil {
		if schema.Ref.Types != nil {
			return schema.Ref.Types.ToStrings()
		}
		schema = schema.Ref
	}
	schemaTypes := []string{}
	for _, referenced := range schema.OneOf {
		if referenced.Types != nil {
			schemaTypes = append(schemaTypes, referenced.Types.ToStrings()[0])
		} else if referenced.Ref != nil {
			schemaTypes = append(schemaTypes, referenced.Ref.Types.ToStrings()[0])
		}
	}
	return schemaTypes
}

func extractDetail(schema *jsonschema.Schema) *string {
	schemaTypes := referencedTypes(schema)
	slices.Sort(schemaTypes)
	return types.CreateStringPointer(strings.Join(schemaTypes, " or "))
}

func insertText(spacing, attributeName string, schema *jsonschema.Schema) *string {
	schemaTypes := referencedTypes(schema)
	if slices.Contains(schemaTypes, "array") {
		if len(schemaTypes) == 1 {
			return types.CreateStringPointer(fmt.Sprintf("%v:\n%v- ", attributeName, spacing))
		} else if len(schemaTypes) == 2 && slices.Contains(schemaTypes, "object") {
			return types.CreateStringPointer(fmt.Sprintf("%v:\n%v", attributeName, spacing))
		}
		return nil
	}
	if slices.Contains(schemaTypes, "object") {
		if len(schemaTypes) == 1 {
			return types.CreateStringPointer(fmt.Sprintf("%v:\n%v", attributeName, spacing))
		}
		return types.CreateStringPointer(fmt.Sprintf("%v:", attributeName))
	}
	return types.CreateStringPointer(fmt.Sprintf("%v: ", attributeName))
}
