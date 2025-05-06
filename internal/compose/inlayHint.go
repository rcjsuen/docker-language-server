package compose

import (
	"fmt"
	"slices"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func allServiceProperties(node ast.Node) map[string]map[string]ast.Node {
	if servicesNode, ok := node.(*ast.MappingNode); ok {
		services := map[string]map[string]ast.Node{}
		for _, serviceNode := range servicesNode.Values {
			if properties, ok := serviceNode.Value.(*ast.MappingNode); ok {
				serviceProperties := map[string]ast.Node{}
				for _, property := range properties.Values {
					serviceProperties[property.Key.GetToken().Value] = property.Value
				}
				services[serviceNode.Key.GetToken().Value] = serviceProperties
			}
		}
		return services
	}
	return nil
}

func hierarchyProperties(service string, serviceProps map[string]map[string]ast.Node, chain []map[string]ast.Node) []map[string]ast.Node {
	if extends, ok := serviceProps[service]["extends"]; ok {
		if s, ok := extends.(*ast.StringNode); ok {
			// block self-referencing recursions
			if s.Value != service {
				chain = append(chain, hierarchyProperties(s.Value, serviceProps, chain)...)
			}
		} else if mappingNode, ok := extends.(*ast.MappingNode); ok {
			external := false
			for _, value := range mappingNode.Values {
				if value.Key.GetToken().Value == "file" {
					external = true
					break
				}
			}

			if !external {
				for _, value := range mappingNode.Values {
					if value.Key.GetToken().Value == "service" {
						chain = append(chain, hierarchyProperties(value.Value.GetToken().Value, serviceProps, chain)...)
					}
				}
			}
		}
	}
	chain = append(chain, serviceProps[service])
	return chain
}

func InlayHint(doc document.ComposeDocument, rng protocol.Range) ([]protocol.InlayHint, error) {
	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	hints := []protocol.InlayHint{}
	for _, docNode := range file.Docs {
		if mappingNode, ok := docNode.Body.(*ast.MappingNode); ok {
			for _, node := range mappingNode.Values {
				if s, ok := node.Key.(*ast.StringNode); ok && s.Value == "services" {
					serviceProps := allServiceProperties(node.Value)
					for service, props := range serviceProps {
						chain := hierarchyProperties(service, serviceProps, []map[string]ast.Node{})
						if len(chain) == 1 {
							continue
						}
						slices.Reverse(chain)
						chain = chain[1:]
						for name, value := range props {
							if name == "extends" {
								continue
							}
							// skip object attributes for now
							if _, ok := value.(*ast.MappingNode); ok {
								continue
							}

							for _, parentProps := range chain {
								if parentProp, ok := parentProps[name]; ok {
									if _, ok := parentProp.(*ast.MappingNode); !ok {
										length := len(value.GetToken().Value)
										if value.GetToken().Type == token.DoubleQuoteType {
											length += 2
										}
										hints = append(hints, protocol.InlayHint{
											Label:       fmt.Sprintf("(parent value: %v)", parentProp.GetToken().Value),
											PaddingLeft: types.CreateBoolPointer(true),
											Position: protocol.Position{
												Line:      uint32(value.GetToken().Position.Line) - 1,
												Character: uint32(value.GetToken().Position.Column + length - 1),
											},
										})
										break
									}
								}
							}
						}
					}
					break
				}
			}
		}
	}
	return hints, nil
}
