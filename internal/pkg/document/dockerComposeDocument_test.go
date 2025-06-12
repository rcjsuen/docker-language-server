package document

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestIncludedPaths(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		paths   []string
	}{
		{
			name: "handle all path types",
			content: `
include:
  - arrayItem.yaml
  - path: pathObject.yaml
  - path:
      - pathArrayItem.yaml`,
			paths: []string{"arrayItem.yaml", "pathObject.yaml", "pathArrayItem.yaml"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := WithReadDocumentFunc(func(uri.URI) ([]byte, error) {
				return []byte(tc.content), nil
			})
			mgr := NewDocumentManager(opts)
			w, err := mgr.Write(context.Background(), "", protocol.DockerComposeLanguage, 1, []byte(tc.content))
			require.NoError(t, err)
			require.True(t, w)
			doc := mgr.Get(context.Background(), "")
			require.Equal(t, tc.paths, doc.(*composeDocument).includedPaths())
		})
	}
}

func TestIncludedFiles(t *testing.T) {
	testCases := []struct {
		name            string
		content         string
		resolved        bool
		externalContent map[string]string
	}{
		{
			name: "ignore unrecognized URIs",
			content: `
include:
  - arrayItem.yaml`,
			resolved: true,
			externalContent: map[string]string{
				"arrayItem.yaml": `name: arrayItem`,
			},
		},
		{
			name: "ignore unrecognized URIs",
			content: `
include:
  - oci://www.docker.com/
  - https://www.docker.com/`,
			resolved:        true,
			externalContent: map[string]string{},
		},
		{
			name: "recurse for more files",
			content: `
include:
  - first.yaml`,
			resolved: true,
			externalContent: map[string]string{
				"first.yaml": `
include:
  - second.yaml`,
				"second.yaml": "name: second",
			},
		},
		{
			name: "two-way self recursion",
			content: `
include:
  - first.yaml`,
			resolved: false,
			externalContent: map[string]string{
				"first.yaml": `
include:
  - compose.yaml`,
			},
		},
		{
			name: "three-way self recursion",
			content: `
include:
  - first.yaml`,
			resolved: false,
			externalContent: map[string]string{
				"first.yaml": `
include:
  - second.yaml`,
				"second.yaml": `
include:
  - compose.yaml`,
			},
		},
		{
			name:            "include node is invalid",
			content:         "include:",
			resolved:        true,
			externalContent: map[string]string{},
		},
	}

	folder := os.TempDir()
	temporaryComposeFileURI := fileURI(folder, "compose.yaml")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for path, content := range tc.externalContent {
				err := os.WriteFile(filepath.Join(folder, path), []byte(content), 0644)
				require.NoError(t, err)
			}
			mgr := NewDocumentManager()
			u := uri.URI(temporaryComposeFileURI)
			w, err := mgr.Write(context.Background(), u, protocol.DockerComposeLanguage, 1, []byte(tc.content))
			require.NoError(t, err)
			require.True(t, w)
			doc := mgr.Get(context.Background(), u)
			includedFiles, resolved := doc.(ComposeDocument).IncludedFiles()
			require.Equal(t, tc.resolved, resolved)
			if resolved {
				require.Len(t, includedFiles, len(tc.externalContent))
			} else {
				require.Len(t, includedFiles, 0)
			}
		})
	}
}

func fileURI(folder, name string) string {
	return fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(folder, name)), "/"))
}
