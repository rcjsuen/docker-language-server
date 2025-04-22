package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
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
							if service.Content[j+1].Kind == yaml.SequenceNode {
								for _, dependency := range service.Content[j+1].Content {
									link := serviceDependencyLink(root, definitionLinkSupport, params, dependency, line, character)
									if link != nil {
										return link, nil
									}
								}
							} else if service.Content[j+1].Kind == yaml.MappingNode {
								for k := 0; k < len(service.Content[j+1].Content); k += 2 {
									link := serviceDependencyLink(root, definitionLinkSupport, params, service.Content[j+1].Content[k], line, character)
									if link != nil {
										return link, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func serviceDependencyLink(root yaml.Node, definitionLinkSupport bool, params *protocol.DefinitionParams, dependency *yaml.Node, line, character int) any {
	if dependency.Line == line && dependency.Column <= character && character <= dependency.Column+len(dependency.Value) {
		serviceRange := serviceDefinitionRange(root, dependency.Value)
		if serviceRange == nil {
			return nil
		}

		return createDefinitionResult(
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

func serviceDefinitionRange(root yaml.Node, serviceName string) *protocol.Range {
	for i := 0; i < len(root.Content[0].Content); i += 2 {
		switch root.Content[0].Content[i].Value {
		case "services":
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

func createDefinitionResult(definitionLinkSupport bool, targetRange protocol.Range, originSelectionRange *protocol.Range, linkURI protocol.URI) any {
	if !definitionLinkSupport {
		return []protocol.Location{
			{
				Range: targetRange,
				URI:   linkURI,
			},
		}
	}

	return []protocol.LocationLink{
		{
			OriginSelectionRange: originSelectionRange,
			TargetRange:          targetRange,
			TargetSelectionRange: targetRange,
			TargetURI:            linkURI,
		},
	}
}
