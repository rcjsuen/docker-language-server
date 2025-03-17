package hcl

import (
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
)

func TestDocumentHighlight(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		position protocol.Position
		ranges   []protocol.DocumentHighlight
	}{
		{
			name:     "cursor before group's block targets attribute quotation",
			content:  "group g { targets = [\"build\"] }\ntarget build {}",
			position: protocol.Position{Line: 0, Character: 21},
			ranges:   nil,
		},
		{
			name:     "cursor after group's block targets attribute quotation",
			content:  "group g { targets = [\"build\"] }\ntarget build {}",
			position: protocol.Position{Line: 0, Character: 28},
			ranges:   nil,
		},
		{
			name:     "cursor in group's block targets attribute pointing at unquoted target",
			content:  "group g { targets = [\"build\"] }\ntarget build {}\ntarget irrelevant {}",
			position: protocol.Position{Line: 0, Character: 25},
			ranges: []protocol.DocumentHighlight{
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 22},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 7},
						End:   protocol.Position{Line: 1, Character: 12},
					},
				},
			},
		},
		{
			name:     "cursor in group's block targets attribute pointing at quoted target",
			content:  "group g { targets = [\"build\"] }\ntarget \"build\" {}\ntarget irrelevant {}",
			position: protocol.Position{Line: 0, Character: 25},
			ranges: []protocol.DocumentHighlight{
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 22},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 8},
						End:   protocol.Position{Line: 1, Character: 13},
					},
				},
			},
		},
		{
			name:     "cursor before target block's quoted label",
			content:  "group g { targets = [\"build\"] }\ntarget \"build\" {}",
			position: protocol.Position{Line: 1, Character: 7},
			ranges:   nil,
		},
		{
			name:     "cursor after target block's quoted label",
			content:  "group g { targets = [\"build\"] }\ntarget \"build\" {}",
			position: protocol.Position{Line: 1, Character: 14},
			ranges:   nil,
		},
		{
			name:     "cursor in target block's unquoted label",
			content:  "group g { targets = [\"build\"] }\ntarget build {}\ntarget irrelevant {}",
			position: protocol.Position{Line: 1, Character: 12},
			ranges: []protocol.DocumentHighlight{
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 22},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 7},
						End:   protocol.Position{Line: 1, Character: 12},
					},
				},
			},
		},
		{
			name:     "cursor in target block's quoted label",
			content:  "group g { targets = [\"build\"] }\ntarget \"build\" {}\ntarget irrelevant {}",
			position: protocol.Position{Line: 1, Character: 12},
			ranges: []protocol.DocumentHighlight{
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 22},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 8},
						End:   protocol.Position{Line: 1, Character: 13},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, tc.position)
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}
