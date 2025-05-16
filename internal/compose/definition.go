package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
)

func insideRange(rng protocol.Range, line, character protocol.UInteger) bool {
	return rng.Start.Line == line && rng.Start.Character <= character && character <= rng.End.Character
}

func Definition(ctx context.Context, definitionLinkSupport bool, doc document.ComposeDocument, params *protocol.DefinitionParams) (any, error) {
	name, dependency := DocumentHighlights(doc, params.Position)
	if len(dependency.documentHighlights) == 0 {
		return nil, nil
	}

	targetURI := params.TextDocument.URI
	var sourceRange *protocol.Range
	var definitionRange *protocol.Range
	for _, highlight := range dependency.documentHighlights {
		if *highlight.Kind == protocol.DocumentHighlightKindWrite {
			definitionRange = &highlight.Range
			if insideRange(highlight.Range, params.Position.Line, params.Position.Character) {
				sourceRange = &highlight.Range
				break
			}
		}
	}

	if definitionRange == nil {
		node, u := dependencyLookup(doc, dependency.dependencyType, name)
		if node != nil {
			definitionRange = rangeFromToken(node.Key.GetToken())
			targetURI = u
		}
	}

	for _, highlight := range dependency.documentHighlights {
		if *highlight.Kind == protocol.DocumentHighlightKindRead {
			if insideRange(highlight.Range, params.Position.Line, params.Position.Character) {
				sourceRange = &highlight.Range
				break
			}
		}
	}

	if definitionRange != nil {
		return types.CreateDefinitionResult(
			definitionLinkSupport,
			*definitionRange,
			sourceRange,
			targetURI,
		), nil
	}
	return nil, nil
}

func dependencyLookup(doc document.ComposeDocument, dependencyType, name string) (*ast.MappingValueNode, string) {
	files, _ := doc.IncludedFiles()
	for u, file := range files {
		for _, doc := range file.Docs {
			if mappingNode, ok := doc.Body.(*ast.MappingNode); ok {
				for _, node := range mappingNode.Values {
					if s, ok := node.Key.(*ast.StringNode); ok && s.Value == dependencyType {
						for _, service := range node.Value.(*ast.MappingNode).Values {
							if s, ok := service.Key.(*ast.StringNode); ok && s.Value == name {
								return service, u
							}
						}
					}
				}
			}
		}
	}
	return nil, ""
}
