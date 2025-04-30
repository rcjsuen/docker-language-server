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

func documentLinkTooltip(testsFolder, fileName string) *string {
	tooltip := filepath.Join(testsFolder, fileName)
	return &tooltip
}

func documentLinkTarget(testsFolder, fileName string) *string {
	path := filepath.Join(testsFolder, fileName)
	target := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(path), "/"))
	return &target
}

func TestDocumentLink_IncludedFiles(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), "composeDocumentLinkTests")
	composeFilePath := filepath.Join(testsFolder, "docker-compose.yml")
	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(composeFilePath), "/"))

	testCases := []struct {
		name    string
		content string
		links   []protocol.DocumentLink
	}{
		{
			name:    "empty file",
			content: "",
			links:   nil,
		},
		{
			name: "included files, string array",
			content: `include:
  - file.yml
  - "file2.yml"`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 4},
						End:   protocol.Position{Line: 1, Character: 12},
					},
					Target:  documentLinkTarget(testsFolder, "file.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file.yml"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 5},
						End:   protocol.Position{Line: 2, Character: 14},
					},
					Target:  documentLinkTarget(testsFolder, "file2.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file2.yml"),
				},
			},
		},
		{
			name: "included files, path as a string",
			content: `include:
  - path: file.yml
  - path: "file2.yml"`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 18},
					},
					Target:  documentLinkTarget(testsFolder, "file.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file.yml"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 11},
						End:   protocol.Position{Line: 2, Character: 20},
					},
					Target:  documentLinkTarget(testsFolder, "file2.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file2.yml"),
				},
			},
		},
		{
			name: "included files, mixed paths",
			content: `
include:
  - file.yml
  - path: file2.yml
  - path:
    - file3.yml`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 4},
						End:   protocol.Position{Line: 2, Character: 12},
					},
					Target:  documentLinkTarget(testsFolder, "file.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file.yml"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 10},
						End:   protocol.Position{Line: 3, Character: 19},
					},
					Target:  documentLinkTarget(testsFolder, "file2.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file2.yml"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 6},
						End:   protocol.Position{Line: 5, Character: 15},
					},
					Target:  documentLinkTarget(testsFolder, "file3.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file3.yml"),
				},
			},
		},
		{
			name:    "include declared with no content",
			content: "include:",
			links:   []protocol.DocumentLink{},
		},
		{
			name: "regular file",
			content: `
services:
  backend:

include:
  - file.yml
  - "file2.yml"`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 4},
						End:   protocol.Position{Line: 5, Character: 12},
					},
					Target:  documentLinkTarget(testsFolder, "file.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file.yml"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 6, Character: 5},
						End:   protocol.Position{Line: 6, Character: 14},
					},
					Target:  documentLinkTarget(testsFolder, "file2.yml"),
					Tooltip: documentLinkTooltip(testsFolder, "file2.yml"),
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
			name: "two services",
			content: `
services:
  test:
    image: alpine
  test2:
    image: postgres`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 17},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 11},
						End:   protocol.Position{Line: 5, Character: 19},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/postgres"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/postgres"),
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
			links: nil,
		},
		{
			name: "image: alpine",
			content: `
---
services:
  backend:
    image: alpine:3.20
---
services:
  backend2:
    image: alpine:3.21`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 4, Character: 11},
						End:   protocol.Position{Line: 4, Character: 17},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 8, Character: 11},
						End:   protocol.Position{Line: 8, Character: 17},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
				},
			},
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
