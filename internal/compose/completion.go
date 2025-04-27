package compose

import (
	"context"
	"slices"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
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
	if lspLine >= len(lines) {
		return nil, nil
	}

	if strings.HasPrefix(strings.TrimSpace(lines[lspLine]), "#") {
		return nil, nil
	}

	root := doc.RootNode()
	if len(root.Content) == 0 {
		return nil, nil
	}

	line := int(lspLine) + 1
	character := int(params.Position.Character) + 1
	topLevel, _, _ := NodeStructure(line, root.Content[0].Content)
	if len(topLevel) == 0 {
		return nil, nil
	} else if len(topLevel) == 1 {
		return nil, nil
	} else if topLevel[1].Column >= character {
		return nil, nil
	} else if len(topLevel) > 2 && topLevel[1].Column < character && character < topLevel[2].Column {
		topLevel = []*yaml.Node{topLevel[0], topLevel[1]}
	}

	if topLevel[0].Line == line {
		return nil, nil
	}

	items := []protocol.CompletionItem{}
	for attributeName, schema := range nodeProperties(topLevel, line, character) {
		items = append(items, protocol.CompletionItem{
			Detail: extractDetail(schema),
			Label:  attributeName,
		})
	}
	if len(items) == 0 {
		return nil, nil
	}
	slices.SortFunc(items, func(a, b protocol.CompletionItem) int {
		return strings.Compare(a.Label, b.Label)
	})
	return &protocol.CompletionList{Items: items}, nil
}

func NodeStructure(line int, rootNodes []*yaml.Node) ([]*yaml.Node, *yaml.Node, bool) {
	if len(rootNodes) == 0 {
		return nil, nil, false
	}

	var topLevel *yaml.Node
	var content *yaml.Node
	for i := 0; i < len(rootNodes); i += 2 {
		if rootNodes[i].Line < line {
			topLevel = rootNodes[i]
			content = rootNodes[i+1]
		} else if rootNodes[i].Line == line {
			return []*yaml.Node{rootNodes[i]}, rootNodes[i+1], true
		} else if line < rootNodes[i].Line {
			break
		}
	}
	nodes := []*yaml.Node{topLevel}
	candidates, subcontent := walkNodes(line, content.Content)
	nodes = append(nodes, candidates...)
	if subcontent != nil {
		content = subcontent
	}
	return nodes, content, false
}

func walkNodes(line int, nodes []*yaml.Node) ([]*yaml.Node, *yaml.Node) {
	var candidate *yaml.Node
	var candidateContent *yaml.Node
	for i := 0; i < len(nodes); i += 2 {
		if nodes[i].Line < line {
			candidate = nodes[i]
			if candidate.Kind == yaml.MappingNode {
				return walkNodes(line, candidate.Content)
			}
			if len(nodes) == i+1 {
				return []*yaml.Node{candidate}, nil
			}
			candidateContent = nodes[i+1]
		} else if nodes[i].Line == line {
			if nodes[i].Kind == yaml.MappingNode {
				return walkNodes(line, nodes[i].Content)
			}
			if len(nodes) == 1 {
				return []*yaml.Node{nodes[i]}, nil
			}
			return []*yaml.Node{nodes[i]}, nodes[i+1]
		} else if line < nodes[i].Line {
			break
		}
	}
	if candidateContent == nil {
		return []*yaml.Node{}, nil
	}
	walked, subcontent := walkNodes(line, candidateContent.Content)
	candidates := []*yaml.Node{candidate}
	candidates = append(candidates, walked...)
	if subcontent != nil {
		candidateContent = subcontent
	}
	return candidates, candidateContent
}

func extractDetail(schema *jsonschema.Schema) *string {
	if schema != nil {
		if schema.Types != nil {
			schemaTypes := schema.Types.ToStrings()
			return types.CreateStringPointer(strings.Join(schemaTypes, " or "))
		} else if schema.Ref != nil {
			if schema.Ref.Types != nil {
				schemaTypes := schema.Ref.Types.ToStrings()
				return types.CreateStringPointer(strings.Join(schemaTypes, " or "))
			}
			schema = schema.Ref
		}
		referencedTypes := []string{}
		for _, referenced := range schema.OneOf {
			if referenced.Types != nil {
				referencedTypes = append(referencedTypes, referenced.Types.ToStrings()[0])
			} else if referenced.Ref != nil {
				referencedTypes = append(referencedTypes, referenced.Ref.Types.ToStrings()[0])
			}
		}
		slices.Sort(referencedTypes)
		return types.CreateStringPointer(strings.Join(referencedTypes, " or "))
	}
	return nil
}
