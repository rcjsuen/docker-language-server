package hcl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestInlayHint(t *testing.T) {
	testCases := []struct {
		name              string
		content           string
		dockerfileContent string
		rng               protocol.Range
		items             []protocol.InlayHint
	}{
		{
			name:              "args lookup",
			content:           "target t1 {\n  args = {\n    undefined = \"test\"\n    empty = \"test\"\n    defined = \"test\"\n}\n}",
			dockerfileContent: "FROM scratch\nARG undefined\nARG empty=\nARG defined=value\n",
			rng: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 5, Character: 0},
			},
			items: []protocol.InlayHint{
				{
					Label:       "(default value: value)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 4, Character: 20},
				},
			},
		},
		{
			name:              "args lookup outside the range",
			content:           "target t1 {\n  args = {\n    undefined = \"test\"\n    empty = \"test\"\n    defined = \"test\"\n}\n}\n\n\n\n",
			dockerfileContent: "FROM scratch\nARG undefined\nARG empty=\nARG defined=value\n",
			rng: protocol.Range{
				Start: protocol.Position{Line: 8, Character: 0},
				End:   protocol.Position{Line: 8, Character: 0},
			},
			items: []protocol.InlayHint{},
		},
		{
			name:    "args lookup to a different context folder",
			content: "target \"backend\" {\n  context = \"./backend\"\n  args = {\n    BACKEND_VAR=\"changed\"\n  }\n}",
			rng: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 4, Character: 0},
			},
			items: []protocol.InlayHint{
				{
					Label:       "(default value: backend_value)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 3, Character: 25},
				},
			},
		},
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	inlayHintTestFolderPath := filepath.Join(projectRoot, "testdata", "inlayHint")
	dockerfilePath := filepath.Join(inlayHintTestFolderPath, "Dockerfile")
	bakeFilePath := filepath.Join(inlayHintTestFolderPath, "docker-bake.hcl")
	bakeFileURI := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(bakeFilePath), "/")))
	dockerfileURI := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(dockerfilePath), "/")))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			if len(tc.content) > 0 {
				changed, err := manager.Write(context.Background(), dockerfileURI, protocol.DockerfileLanguage, 1, []byte(tc.dockerfileContent))
				defer manager.Remove(dockerfileURI)
				require.NoError(t, err)
				require.True(t, changed)
			}
			bytes := []byte(tc.content)
			err := os.WriteFile(bakeFilePath, bytes, 0644)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := os.Remove(bakeFilePath)
				require.NoError(t, err)
			})

			doc := document.NewBakeHCLDocument(bakeFileURI, 1, []byte(tc.content))
			items, err := InlayHint(manager, doc, tc.rng)
			require.NoError(t, err)
			require.Equal(t, tc.items, items)
		})
	}
}
