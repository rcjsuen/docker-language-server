package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"gopkg.in/yaml.v3"
)

func DocumentSymbol(ctx context.Context, doc document.ComposeDocument) (result []any, err error) {
	root := doc.RootNode()
	if len(root.Content) > 0 {
		for i := range root.Content[0].Content {
			switch root.Content[0].Content[i].Value {
			case "services":
				symbols := createSymbol(root.Content[0].Content, i+1, protocol.SymbolKindClass)
				result = append(result, symbols...)
			case "networks":
				symbols := createSymbol(root.Content[0].Content, i+1, protocol.SymbolKindInterface)
				result = append(result, symbols...)
			case "volumes":
				symbols := createSymbol(root.Content[0].Content, i+1, protocol.SymbolKindFile)
				result = append(result, symbols...)
			case "configs":
				symbols := createSymbol(root.Content[0].Content, i+1, protocol.SymbolKindVariable)
				result = append(result, symbols...)
			case "secrets":
				symbols := createSymbol(root.Content[0].Content, i+1, protocol.SymbolKindKey)
				result = append(result, symbols...)
			case "include":
				for _, included := range root.Content[0].Content[i+1].Content {
					switch included.Kind {
					case yaml.MappingNode:
						// long syntax with an object
						for j := range included.Content {
							if included.Content[j].Value == "path" {
								switch included.Content[j+1].Kind {
								case yaml.SequenceNode:
									for _, path := range included.Content[j+1].Content {
										character := uint32(path.Column - 1)
										rng := protocol.Range{
											Start: protocol.Position{
												Line:      uint32(path.Line - 1),
												Character: character,
											},
											End: protocol.Position{
												Line:      uint32(path.Line - 1),
												Character: character + uint32(len(path.Value)),
											},
										}
										result = append(result, &protocol.DocumentSymbol{
											Name:           path.Value,
											Kind:           protocol.SymbolKindModule,
											Range:          rng,
											SelectionRange: rng,
										})
									}
								case yaml.ScalarNode:
									character := uint32(included.Content[j+1].Column - 1)
									rng := protocol.Range{
										Start: protocol.Position{
											Line:      uint32(included.Content[j+1].Line - 1),
											Character: character,
										},
										End: protocol.Position{
											Line:      uint32(included.Content[j+1].Line - 1),
											Character: character + uint32(len(included.Content[j+1].Value)),
										},
									}
									result = append(result, &protocol.DocumentSymbol{
										Name:           included.Content[j+1].Value,
										Kind:           protocol.SymbolKindModule,
										Range:          rng,
										SelectionRange: rng,
									})
								}
							}
						}
					case yaml.ScalarNode:
						// include:
						//   - abc.yml
						//   - def.yml
						character := uint32(included.Column - 1)
						rng := protocol.Range{
							Start: protocol.Position{
								Line:      uint32(included.Line - 1),
								Character: character,
							},
							End: protocol.Position{
								Line:      uint32(included.Line - 1),
								Character: character + uint32(len(included.Value)),
							},
						}
						result = append(result, &protocol.DocumentSymbol{
							Name:           included.Value,
							Kind:           protocol.SymbolKindModule,
							Range:          rng,
							SelectionRange: rng,
						})
					}
				}
			}
		}
	}
	return result, nil
}

func createSymbol(nodes []*yaml.Node, idx int, kind protocol.SymbolKind) (result []any) {
	for _, service := range nodes[idx].Content {
		if service.Value != "" {
			character := uint32(service.Column - 1)
			rng := protocol.Range{
				Start: protocol.Position{
					Line:      uint32(service.Line - 1),
					Character: character,
				},
				End: protocol.Position{
					Line:      uint32(service.Line - 1),
					Character: character + uint32(len(service.Value)),
				},
			}
			result = append(result, &protocol.DocumentSymbol{
				Name:           service.Value,
				Kind:           kind,
				Range:          rng,
				SelectionRange: rng,
			})
		}
	}
	return result
}
