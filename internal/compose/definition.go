package compose

import (
	"context"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
)

func insideRange(rng protocol.Range, line, character protocol.UInteger) bool {
	return rng.Start.Line == line && rng.Start.Character <= character && character <= rng.End.Character
}

func Definition(ctx context.Context, definitionLinkSupport bool, doc document.ComposeDocument, params *protocol.DefinitionParams) (any, error) {
	highlights, err := DocumentHighlight(doc, params.Position)
	if err != nil {
		return nil, err
	}

	var sourceRange *protocol.Range
	var definitionRange *protocol.Range
	for _, highlight := range highlights {
		if *highlight.Kind == protocol.DocumentHighlightKindWrite {
			definitionRange = &highlight.Range
			if insideRange(highlight.Range, params.Position.Line, params.Position.Character) {
				sourceRange = &highlight.Range
				break
			}
		}
	}

	for _, highlight := range highlights {
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
			params.TextDocument.URI,
		), nil
	}
	return nil, nil
}
