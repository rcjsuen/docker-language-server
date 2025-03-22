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
	}

	dockerfilePath := filepath.ToSlash(filepath.Join(os.TempDir(), "Dockerfile"))
	dockerBakePath := filepath.ToSlash(filepath.Join(os.TempDir(), "docker-bake.hcl"))
	temporaryDockerfile := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerfilePath, "/")))
	temporaryBakeFile := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerBakePath, "/")))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			changed, err := manager.Write(context.Background(), temporaryDockerfile, protocol.DockerfileLanguage, 1, []byte(tc.dockerfileContent))
			require.NoError(t, err)
			require.True(t, changed)
			doc := document.NewBakeHCLDocument(temporaryBakeFile, 1, []byte(tc.content))
			items, err := InlayHint(manager, doc, tc.rng)
			require.NoError(t, err)
			require.Equal(t, tc.items, items)
		})
	}
}
