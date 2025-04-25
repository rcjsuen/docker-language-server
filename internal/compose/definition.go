package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"gopkg.in/yaml.v3"
)

func Definition(ctx context.Context, definitionLinkSupport bool, doc document.ComposeDocument, params *protocol.DefinitionParams) (any, error) {
	line := int(params.Position.Line) + 1
	character := int(params.Position.Character) + 1
	root := doc.RootNode()
	if len(root.Content) > 0 {
		for i := 0; i < len(root.Content[0].Content); i += 2 {
			switch root.Content[0].Content[i].Value {
			case "services":
				for _, service := range root.Content[0].Content[i+1].Content {
					for j := 0; j < len(service.Content); j += 2 {
						if service.Content[j].Value == "depends_on" {
							if service.Content[j+1].Kind == yaml.MappingNode {
								for k := 0; k < len(service.Content[j+1].Content); k += 2 {
									link := dependencyLink(root, definitionLinkSupport, params, service.Content[j+1].Content[k], line, character, "services")
									if link != nil {
										return link, nil
									}
								}
							}
							if service.Content[j+1].Kind == yaml.SequenceNode {
								for _, dependency := range service.Content[j+1].Content {
									link := dependencyLink(root, definitionLinkSupport, params, dependency, line, character, "services")
									if link != nil {
										return link, nil
									}
								}
							}
						}

						link := lookupDependencyLink(root, definitionLinkSupport, params, service, j, line, character, "configs")
						if link != nil {
							return link, nil
						}

						link = lookupDependencyLink(root, definitionLinkSupport, params, service, j, line, character, "networks")
						if link != nil {
							return link, nil
						}

						link = lookupDependencyLink(root, definitionLinkSupport, params, service, j, line, character, "secrets")
						if link != nil {
							return link, nil
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func lookupDependencyLink(root yaml.Node, definitionLinkSupport bool, params *protocol.DefinitionParams, service *yaml.Node, index, line, character int, nodeName string) any {
	if service.Content[index].Value == nodeName && service.Content[index+1].Kind == yaml.SequenceNode {
		for _, dependency := range service.Content[index+1].Content {
			link := dependencyLink(root, definitionLinkSupport, params, dependency, line, character, nodeName)
			if link != nil {
				return link
			}
		}
	}
	return nil
}

func dependencyLink(root yaml.Node, definitionLinkSupport bool, params *protocol.DefinitionParams, dependency *yaml.Node, line, character int, nodeName string) any {
	if dependency.Line == line && dependency.Column <= character && character <= dependency.Column+len(dependency.Value) {
		serviceRange := serviceDefinitionRange(root, nodeName, dependency.Value)
		if serviceRange == nil {
			return nil
		}

		return types.CreateDefinitionResult(
			definitionLinkSupport,
			*serviceRange,
			&protocol.Range{
				Start: protocol.Position{
					Line:      params.Position.Line,
					Character: protocol.UInteger(dependency.Column - 1),
				},
				End: protocol.Position{
					Line:      params.Position.Line,
					Character: protocol.UInteger(dependency.Column + len(dependency.Value) - 1),
				},
			},
			params.TextDocument.URI,
		)
	}
	return nil
}

func serviceDefinitionRange(root yaml.Node, nodeName, serviceName string) *protocol.Range {
	for i := 0; i < len(root.Content[0].Content); i += 2 {
		if root.Content[0].Content[i].Value == nodeName {
			for _, service := range root.Content[0].Content[i+1].Content {
				if service.Value == serviceName {
					return &protocol.Range{
						Start: protocol.Position{
							Line:      protocol.UInteger(service.Line) - 1,
							Character: protocol.UInteger(service.Column - 1),
						},
						End: protocol.Position{
							Line:      protocol.UInteger(service.Line) - 1,
							Character: protocol.UInteger(service.Column + len(serviceName) - 1),
						},
					}
				}
			}
		}
	}
	return nil
}
