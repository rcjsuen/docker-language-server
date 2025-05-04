package compose

import (
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

func Rename(doc document.ComposeDocument, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	highlights, err := DocumentHighlight(doc, params.Position)
	if err != nil || len(highlights) == 0 {
		return nil, err
	}

	edits := []protocol.TextEdit{}
	for _, highlight := range highlights {
		edits = append(edits, protocol.TextEdit{
			NewText: params.NewName,
			Range:   highlight.Range,
		})
	}
	return &protocol.WorkspaceEdit{
		Changes: map[protocol.DocumentUri][]protocol.TextEdit{
			params.TextDocument.URI: edits,
		},
	}, nil
}
