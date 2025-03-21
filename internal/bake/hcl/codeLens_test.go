package hcl

import (
	"context"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
)

func TestCodeLens(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		codeLens []protocol.CodeLens
	}{
		{
			name:     "empty file",
			content:  "",
			codeLens: []protocol.CodeLens{},
		},
		{
			name:     "target block with no label",
			content:  "target {}",
			codeLens: []protocol.CodeLens{},
		},
		{
			name:    "target block",
			content: "target \"first\" {  target = \"abc\"\n}",
			codeLens: []protocol.CodeLens{
				{
					Command: &protocol.Command{
						Title:   "Build",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "build",
								"target": "first",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
				{
					Command: &protocol.Command{
						Title:   "Check",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "check",
								"target": "first",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
				{
					Command: &protocol.Command{
						Title:   "Print",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "print",
								"target": "first",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
			},
		},
		{
			name:    "group block",
			content: "group \"g1\" {}",
			codeLens: []protocol.CodeLens{
				{
					Command: &protocol.Command{
						Title:   "Build",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "build",
								"target": "g1",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
				{
					Command: &protocol.Command{
						Title:   "Check",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "check",
								"target": "g1",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
				{
					Command: &protocol.Command{
						Title:   "Print",
						Command: types.BakeBuildCommandId,
						Arguments: []any{
							map[string]string{
								"call":   "print",
								"target": "g1",
								"cwd":    "/tmp",
							},
						},
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0},
						End:   protocol.Position{Line: 0},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("file:///tmp/docker-bake.hcl", 1, []byte(tc.content))
			codeLens, err := CodeLens(context.Background(), "file:///tmp/docker-bake.hcl", doc)
			require.NoError(t, err)
			require.Equal(t, tc.codeLens, codeLens)
		})
	}
}
