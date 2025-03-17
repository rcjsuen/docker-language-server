package compose

import (
	"context"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
)

func TestDocumentLink(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		links   []protocol.DocumentLink
	}{
		{
			name: "included files, short syntax",
			content: `include:
  - file.yml`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 12},
					},
					Target:  types.CreateStringPointer("file:///home/user/file.yml"),
					Tooltip: types.CreateStringPointer("/home/user/file.yml"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument("docker-compose.yml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), "file:///home/user/docker-compose.yml.hcl", doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}
