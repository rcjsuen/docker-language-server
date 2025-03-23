package hcl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestCodeLens(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), "codeLensTests")
	bakeFilePath := filepath.Join(testsFolder, "docker-bake.hcl")
	uriString := fmt.Sprintf("file:///%v", filepath.ToSlash(bakeFilePath))

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
								"cwd":    testsFolder,
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
								"cwd":    testsFolder,
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
								"cwd":    testsFolder,
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
								"cwd":    testsFolder,
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
								"cwd":    testsFolder,
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
								"cwd":    testsFolder,
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
			doc := document.NewBakeHCLDocument(uri.URI(uriString), 1, []byte(tc.content))
			codeLens, err := CodeLens(context.Background(), uriString, doc)
			require.NoError(t, err)
			require.Equal(t, tc.codeLens, codeLens)
		})
	}
}
