package hcl

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
			name:    "dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 30},
					},
					Target:  types.CreateStringPointer("file:///home/user/Dockerfile.api"),
					Tooltip: types.CreateStringPointer("/home/user/Dockerfile.api"),
				},
			},
		},
		{
			name:    "./dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"./Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 32},
					},
					Target:  types.CreateStringPointer("file:///home/user/Dockerfile.api"),
					Tooltip: types.CreateStringPointer("/home/user/Dockerfile.api"),
				},
			},
		},
		{
			name:    "../dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 33},
					},
					Target:  types.CreateStringPointer("file:///home/Dockerfile.api"),
					Tooltip: types.CreateStringPointer("/home/Dockerfile.api"),
				},
			},
		},
		{
			name:    "folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"folder/Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 37},
					},
					Target:  types.CreateStringPointer("file:///home/user/folder/Dockerfile.api"),
					Tooltip: types.CreateStringPointer("/home/user/folder/Dockerfile.api"),
				},
			},
		},
		{
			name:    "../folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../folder/Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 40},
					},
					Target:  types.CreateStringPointer("file:///home/folder/Dockerfile.api"),
					Tooltip: types.CreateStringPointer("/home/folder/Dockerfile.api"),
				},
			},
		},
		{
			name:    "dockerfile attribugte points to undefined variable",
			content: "target \"api\" {\n  dockerfile = undefined\n}",
			links:   []protocol.DocumentLink{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), "file:///home/user/docker-bake.hcl", doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}
