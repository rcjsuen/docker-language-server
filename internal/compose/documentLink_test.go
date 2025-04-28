package compose

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

func TestDocumentLink_IncludedFiles(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), "composeDocumentLinkTests")
	composeFilePath := filepath.Join(testsFolder, "docker-compose.yml")
	referencedFilePath := filepath.Join(testsFolder, "file.yml")
	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(composeFilePath), "/"))
	referencedFileStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(referencedFilePath), "/"))

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
					Target:  types.CreateStringPointer(referencedFileStringURI),
					Tooltip: types.CreateStringPointer(referencedFilePath),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument("docker-compose.yml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}

func TestDocumentLink_ImageLinks(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		links   []protocol.DocumentLink
	}{
		{
			name: "image: alpine",
			content: `
services:
  test:
    image: alpine`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 17},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
				},
			},
		},
		{
			name: "image: alpine:1.23",
			content: `
services:
  test:
    image: alpine:1.23`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 17},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
				},
			},
		},
		{
			name: "image: grafana/grafana",
			content: `
services:
  test:
    image: grafana/grafana`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 26},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/r/grafana/grafana"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/r/grafana/grafana"),
				},
			},
		},
		{
			name: "image: grafana/grafana:1.23",
			content: `
services:
  test:
    image: grafana/grafana:1.23`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 26},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/r/grafana/grafana"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/r/grafana/grafana"),
				},
			},
		},
		{
			name: "image: ghcr.io/super-linter/super-linter",
			content: `
services:
  test:
    image: ghcr.io/super-linter/super-linter`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 44},
					},
					Target:  types.CreateStringPointer("https://ghcr.io/super-linter/super-linter"),
					Tooltip: types.CreateStringPointer("https://ghcr.io/super-linter/super-linter"),
				},
			},
		},
		{
			name: "image: ghcr.io/super-linter/super-linter:1.23",
			content: `
services:
  test:
    image: ghcr.io/super-linter/super-linter:1.23`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 44},
					},
					Target:  types.CreateStringPointer("https://ghcr.io/super-linter/super-linter"),
					Tooltip: types.CreateStringPointer("https://ghcr.io/super-linter/super-linter"),
				},
			},
		},
		{
			name: "image: mcr.microsoft.com/powershell",
			content: `
services:
  test:
    image: mcr.microsoft.com/powershell`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 39},
					},
					Target:  types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/powershell"),
					Tooltip: types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/powershell"),
				},
			},
		},
		{
			name: "image: mcr.microsoft.com/powershell:1.23",
			content: `
services:
  test:
    image: mcr.microsoft.com/powershell:1.23`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 39},
					},
					Target:  types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/powershell"),
					Tooltip: types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/powershell"),
				},
			},
		},
		{
			name: "image: mcr.microsoft.com/windows/servercore",
			content: `
services:
  test:
    image: mcr.microsoft.com/windows/servercore`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 47},
					},
					Target:  types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/windows/servercore"),
					Tooltip: types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/windows/servercore"),
				},
			},
		},
		{
			name: "image: mcr.microsoft.com/windows/servercore:1.23",
			content: `
services:
  test:
    image: mcr.microsoft.com/windows/servercore:1.23`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 47},
					},
					Target:  types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/windows/servercore"),
					Tooltip: types.CreateStringPointer("https://mcr.microsoft.com/artifact/mar/windows/servercore"),
				},
			},
		},
		{
			name: "invalid services",
			content: `
services:
  - `,
			links: []protocol.DocumentLink{},
		},
	}

	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument("docker-compose.yml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}
