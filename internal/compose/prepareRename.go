package compose

import (
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

func PrepareRename(doc document.ComposeDocument, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	highlights, err := DocumentHighlight(doc, params.Position)
	if err != nil || len(highlights) == 0 {
		return nil, err
	}

	for _, highlight := range highlights {
		if insideRange(highlight.Range, params.Position.Line, params.Position.Character) {
			return &highlight.Range, nil
		}
	}
	return nil, nil
}
