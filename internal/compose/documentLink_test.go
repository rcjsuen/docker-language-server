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
			links:   []protocol.DocumentLink{},
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
			name: "included files, attribute misspelt",
			content: `include:
  - path2: file.yml`,
			links: []protocol.DocumentLink{},
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
		{
			name: "anchors and aliases",
			content: `
include:
  - &link compose.other.yaml
  - *link`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 10},
						End:   protocol.Position{Line: 2, Character: 28},
					},
					Target:  documentLinkTarget(testsFolder, "compose.other.yaml"),
					Tooltip: documentLinkTooltip(testsFolder, "compose.other.yaml"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), "docker-compose.yml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}

func TestDocumentLink_ServiceImageLinks(t *testing.T) {
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
			name: "quoted string",
			content: `
services:
  test:
    image: "alpine"`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 12},
						End:   protocol.Position{Line: 3, Character: 18},
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
			name: "image: ghcr.io",
			content: `
services:
  test:
    image: ghcr.io`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: ghcr.io/",
			content: `
services:
  test:
    image: ghcr.io/`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: ghcr.io:",
			content: `
services:
  test:
    image: ghcr.io:`,
			links: nil,
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
			name: "image: mcr.microsoft.com",
			content: `
services:
  test:
    image: mcr.microsoft.com`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: mcr.microsoft.com/",
			content: `
services:
  test:
    image: mcr.microsoft.com/`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: mcr.microsoft.com:",
			content: `
services:
  test:
    image: mcr.microsoft.com:`,
			links: nil,
		},
		{
			name: "image: quay.io/prometheus/node-exporter",
			content: `
services:
  test:
    image: quay.io/prometheus/node-exporter`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 43},
					},
					Target:  types.CreateStringPointer("https://quay.io/repository/prometheus/node-exporter"),
					Tooltip: types.CreateStringPointer("https://quay.io/repository/prometheus/node-exporter"),
				},
			},
		},
		{
			name: "image: quay.io/prometheus/node-exporter:v1.9.1",
			content: `
services:
  test:
    image: quay.io/prometheus/node-exporter:v1.9.1`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 11},
						End:   protocol.Position{Line: 3, Character: 43},
					},
					Target:  types.CreateStringPointer("https://quay.io/repository/prometheus/node-exporter"),
					Tooltip: types.CreateStringPointer("https://quay.io/repository/prometheus/node-exporter"),
				},
			},
		},
		{
			name: "image: quay.io",
			content: `
services:
  test:
    image: quay.io`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: quay.io/",
			content: `
services:
  test:
    image: quay.io/`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "image: quay.io:",
			content: `
services:
  test:
    image: quay.io:`,
			links: nil,
		},
		{
			name: "anchors and aliases to nothing",
			content: `
services:
  backend:
    image: &alpine
  backend2:
    image: *alpine`,
			links: []protocol.DocumentLink{},
		},
		{
			name: "anchor has string content",
			content: `
services:
  backend:
    image: &alpine alpine:3.21`,
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 19},
						End:   protocol.Position{Line: 3, Character: 25},
					},
					Target:  types.CreateStringPointer("https://hub.docker.com/_/alpine"),
					Tooltip: types.CreateStringPointer("https://hub.docker.com/_/alpine"),
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
		{
			name: "two documents",
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
			mgr := document.NewDocumentManager()
			doc := document.NewComposeDocument(mgr, "docker-compose.yml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}

func TestDocumentLink_ServiceDockerfileLinks(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), t.Name())
	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(testsFolder, "compose.yaml")), "/"))

	testCases := []struct {
		name      string
		content   string
		path      string
		linkRange protocol.Range
	}{
		{
			name: "./Dockerfile2",
			content: `
services:
  test:
    build:
      dockerfile: ./Dockerfile2`,
			path: filepath.Join(testsFolder, "Dockerfile2"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 18},
				End:   protocol.Position{Line: 4, Character: 31},
			},
		},
		{
			name: `"./Dockerfile2"`,
			content: `
services:
  test:
    build:
      dockerfile: "./Dockerfile2"`,
			path: filepath.Join(testsFolder, "Dockerfile2"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 19},
				End:   protocol.Position{Line: 4, Character: 32},
			},
		},
		{
			name: "anchors and aliases to nothing",
			content: `
services:
  test:
    build:
      dockerfile: &file
  test2:
    build:
      dockerfile: *file`,
		},
		{
			name: "anchor has string content",
			content: `
services:
  test:
    build:
      dockerfile: &file ./Dockerfile2
  test2:
    build:
      dockerfile: *file`,
			path: filepath.Join(testsFolder, "Dockerfile2"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 24},
				End:   protocol.Position{Line: 4, Character: 37},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := document.NewDocumentManager()
			doc := document.NewComposeDocument(mgr, "compose.yaml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			if tc.path == "" {
				require.Equal(t, []protocol.DocumentLink{}, links)
			} else {
				link := protocol.DocumentLink{
					Range:   tc.linkRange,
					Target:  types.CreateStringPointer(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(tc.path), "/"))),
					Tooltip: types.CreateStringPointer(tc.path),
				}
				require.Equal(t, []protocol.DocumentLink{link}, links)
			}
		})
	}
}

func TestDocumentLink_ConfigFileLinks(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), t.Name())
	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(testsFolder, "compose.yaml")), "/"))

	testCases := []struct {
		name      string
		content   string
		path      string
		linkRange protocol.Range
	}{
		{
			name: "./httpd.conf",
			content: `
configs:
  test:
    file: ./httpd.conf`,
			path: filepath.Join(testsFolder, "httpd.conf"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 10},
				End:   protocol.Position{Line: 3, Character: 22},
			},
		},
		{
			name: `"./httpd.conf"`,
			content: `
configs:
  test:
    file: "./httpd.conf"`,
			path: filepath.Join(testsFolder, "httpd.conf"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 11},
				End:   protocol.Position{Line: 3, Character: 23},
			},
		},
		{
			name: "anchors and aliases to nothing",
			content: `
configs:
  test:
    file: &configFile
  test2:
    file: *configFile`,
		},
		{
			name: "anchor has string content",
			content: `
configs:
  test:
    file: &configFile ./httpd.conf
  test2:
    file: *configFile`,
			path: filepath.Join(testsFolder, "httpd.conf"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 22},
				End:   protocol.Position{Line: 3, Character: 34},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := document.NewDocumentManager()
			doc := document.NewComposeDocument(mgr, "compose.yaml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			if tc.path == "" {
				require.Equal(t, []protocol.DocumentLink{}, links)
			} else {
				link := protocol.DocumentLink{
					Range:   tc.linkRange,
					Target:  types.CreateStringPointer(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(tc.path), "/"))),
					Tooltip: types.CreateStringPointer(tc.path),
				}
				require.Equal(t, []protocol.DocumentLink{link}, links)
			}
		})
	}
}

func TestDocumentLink_SecretFileLinks(t *testing.T) {
	testsFolder := filepath.Join(os.TempDir(), t.Name())
	composeStringURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(testsFolder, "compose.yaml")), "/"))

	testCases := []struct {
		name      string
		content   string
		path      string
		linkRange protocol.Range
	}{
		{
			name: "./server.cert",
			content: `
secrets:
  test:
    file: ./server.cert`,
			path: filepath.Join(testsFolder, "server.cert"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 10},
				End:   protocol.Position{Line: 3, Character: 23},
			},
		},
		{
			name: `"./server.cert"`,
			content: `
secrets:
  test:
    file: "./server.cert"`,
			path: filepath.Join(testsFolder, "server.cert"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 11},
				End:   protocol.Position{Line: 3, Character: 24},
			},
		},
		{
			name: "anchors and aliases to nothing",
			content: `
secrets:
  test:
    file: &configFile
  test2:
    file: *configFile`,
		},
		{
			name: "anchor has string content",
			content: `
secrets:
  test:
    file: &configFile ./server.cert
  test2:
    file: *configFile`,
			path: filepath.Join(testsFolder, "server.cert"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 22},
				End:   protocol.Position{Line: 3, Character: 35},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := document.NewDocumentManager()
			doc := document.NewComposeDocument(mgr, "compose.yaml", 1, []byte(tc.content))
			links, err := DocumentLink(context.Background(), composeStringURI, doc)
			require.NoError(t, err)
			if tc.path == "" {
				require.Equal(t, []protocol.DocumentLink{}, links)
			} else {
				link := protocol.DocumentLink{
					Range:   tc.linkRange,
					Target:  types.CreateStringPointer(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(tc.path), "/"))),
					Tooltip: types.CreateStringPointer(tc.path),
				}
				require.Equal(t, []protocol.DocumentLink{link}, links)
			}
		})
	}
}
