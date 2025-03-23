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
)

func TestDocumentLink(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), "documentLinkTests")
	userFolder := filepath.Join(testsFolder, "user")
	bakeFilePath := filepath.Join(userFolder, "docker-bake.hcl")
	bakeFileStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(bakeFilePath, "/"))

	testCases := []struct {
		name      string
		content   string
		path      string
		linkRange protocol.Range
		links     func(path string) []protocol.DocumentLink
	}{
		{
			name:    "dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"Dockerfile.api\"\n}",
			path:    filepath.Join(userFolder, "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 30},
			},
		},
		{
			name:    "./dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"./Dockerfile.api\"\n}",
			path:    filepath.Join(userFolder, "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 32},
			},
		},
		{
			name:    "../dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../Dockerfile.api\"\n}",
			path:    filepath.Join(testsFolder, "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 33},
			},
		},
		{
			name:    "folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"folder/Dockerfile.api\"\n}",
			path:    filepath.Join(filepath.Join(userFolder, "folder"), "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 37},
			},
		},
		{
			name:    "../folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../folder/Dockerfile.api\"\n}",
			path:    filepath.Join(filepath.Join(testsFolder, "folder"), "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 40},
			},
		},
		{
			name:    "dockerfile attribute points to undefined variable",
			content: "target \"api\" {\n  dockerfile = undefined\n}",
			path:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("docker-bake.hcl", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), bakeFileStringURI, doc)
			require.NoError(t, err)

			if tc.path == "" {
				require.Equal(t, []protocol.DocumentLink{}, links)
			} else {
				link := protocol.DocumentLink{
					Range:   tc.linkRange,
					Target:  types.CreateStringPointer(fmt.Sprintf("file:///%v", strings.TrimPrefix(tc.path, "/"))),
					Tooltip: types.CreateStringPointer(tc.path),
				}
				require.Equal(t, []protocol.DocumentLink{link}, links)
			}
		})
	}
}
