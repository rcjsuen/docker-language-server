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
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestHover(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		result    *protocol.Hover
	}{
		{
			name:      "version description",
			content:   "version: 1.2.3",
			line:      0,
			character: 4,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "declared for backward compatibility, ignored. Please remove it.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/version-and-name/)",
				},
			},
		},
		{
			name:      "name description",
			content:   "name: customName",
			line:      0,
			character: 4,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "define the Compose project name, until user defines one explicitly.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/version-and-name/)",
				},
			},
		},
		{
			name:      "name but in the whitespace",
			content:   "name: customName",
			line:      0,
			character: 5,
			result:    nil,
		},
		{
			name:      "name but in the attribute value",
			content:   "name: customName",
			line:      0,
			character: 12,
			result:    nil,
		},
		{
			name:      "include description",
			content:   "include:",
			line:      0,
			character: 4,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "compose sub-projects to be included.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/include/)",
				},
			},
		},
		{
			name: "include's project_directory attribute",
			content: `include:
- project_directory: folder`,
			line:      1,
			character: 7,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Path to resolve relative paths set in the Compose file\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/include/#project_directory)",
				},
			},
		},
		{
			name:      "incomplete node",
			content:   "version",
			line:      0,
			character: 2,
			result:    nil,
		},
		{
			name: "type (of volumes) enum values when hovering over the attribute's name",
			content: `
services:
  test:
    volumes:
      - type:`,
			line:      4,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "The mount type: bind for mounting host directories, volume for named volumes, tmpfs for temporary filesystems, cluster for cluster volumes, npipe for named pipes, or image for mounting from an image.\n\nAllowed values:\n- `bind`\n- `cluster`\n- `image`\n- `npipe`\n- `tmpfs`\n- `volume`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "selinux enum values when hovering over the attribute's name",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          selinux: `,
			line:      6,
			character: 13,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "SELinux relabeling options: 'z' for shared content, 'Z' for private unshared content.\n\nAllowed values:\n- `Z`\n- `z`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "recursive enum values when hovering over the attribute's name",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          recursive: enabled`,
			line:      6,
			character: 17,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Recursively mount the source directory.\n\nAllowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "recursive enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          recursive: enabled`,
			line:      6,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Recursively mount the source directory.\n\nAllowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "recursive enum values when hovering over the attribute's value",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          recursive: `,
			line:      6,
			character: 13,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Recursively mount the source directory.\n\nAllowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "recursive enum values when hovering over the attribute's value at the end",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          recursive: enabled`,
			line:      6,
			character: 28,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Recursively mount the source directory.\n\nAllowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#volumes)",
				},
			},
		},
		{
			name: "cgroup enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    cgroup:`,
			line:      3,
			character: 7,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Specify the cgroup namespace to join. Use 'host' to use the host's cgroup namespace, or 'private' to use a private cgroup namespace.\n\nAllowed values:\n- `host`\n- `private`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#cgroup)",
				},
			},
		},
		{
			name: "condition enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    depends_on:
      test2:
        condition:`,
			line:      5,
			character: 14,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Condition to wait for. 'service_started' waits until the service has started, 'service_healthy' waits until the service is healthy (as defined by its healthcheck), 'service_completed_successfully' waits until the service has completed successfully.\n\nAllowed values:\n- `service_completed_successfully`\n- `service_healthy`\n- `service_started`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#depends_on)",
				},
			},
		},
		{
			name: "action enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    develop:
      watch:
        - path: "./"
          action: rebuild`,
			line:      6,
			character: 13,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Action to take when a change is detected: rebuild the container, sync files, restart the container, sync and restart, or sync and execute a command.\n\nAllowed values:\n- `rebuild`\n- `restart`\n- `sync`\n- `sync+exec`\n- `sync+restart`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#develop)",
				},
			},
		},
		{
			name: "rollback_config enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    deploy:
      rollback_config:
        order: start-first`,
			line:      5,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Order of operations during rollbacks: 'stop-first' (default) or 'start-first'.\n\nAllowed values:\n- `start-first`\n- `stop-first`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#deploy)",
				},
			},
		},
		{
			name: "update_config enum values when hovering over the attribute's name in the front",
			content: `
services:
  test:
    deploy:
      update_config:
        order: start-first`,
			line:      5,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Order of operations during updates: 'stop-first' (default) or 'start-first'.\n\nAllowed values:\n- `start-first`\n- `stop-first`\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#deploy)",
				},
			},
		},
		{
			name: "hovering over an extends service as a string",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    image: alpine:3.21
    extends: test`,
			line:      6,
			character: 15,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.21` +
						"\n```",
				},
			},
		},
		{
			name: "hovering over an extends service as a string with a comment",
			content: `
services:
  test:
    # comment
    image: alpine:3.21
  test2:
    image: alpine:3.21
    extends: test`,
			line:      7,
			character: 15,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  # comment
  image: alpine:3.21` +
						"\n```",
				},
			},
		},
		{
			name: "hovering over an extends object",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    extends:
      service: test
    image: alpine:3.21`,
			line:      6,
			character: 17,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.21` +
						"\n```",
				},
			},
		},
		{
			name: "hovering over an invalid extends object with invalid attribute",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    image: alpine:3.21
    extends:
      test:`,
			line:      7,
			character: 8,
			result:    nil,
		},
		{
			name: "hovering over an invalid extends object with a service attribute",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    image: alpine:3.21
    extends:
      service:
        test:`,
			line:      8,
			character: 10,
			result:    nil,
		},
		{
			name: "hovering over an extends service with whitespace",
			content: `
services:
  test:
    image: alpine:3.21

    attach: true
  test2:
    image: alpine:3.21
    extends: test`,
			line:      8,
			character: 15,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.21

  attach: true` +
						"\n```",
				},
			},
		},
		{
			name: "hovering over a depends_on array item",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    depends_on:
      - test`,
			line:      6,
			character: 10,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.21` +
						"\n```",
				},
			},
		},
		{
			name: "hovering over a depends_on array item that is not a string",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    depends_on:
      - test:`,
			line:      6,
			character: 10,
			result:    nil,
		},
		{
			name: "hovering over a depends_on object item",
			content: `
services:
  test:
    image: alpine:3.21
  test2:
    depends_on:
      test:`,
			line:      6,
			character: 8,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.21` +
						"\n```",
				},
			},
		},
		{
			name: "container_name attribute on services",
			content: `
services:
  test:
    container_name: abc`,
			line:      3,
			character: 7,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Specify a custom container name, rather than a generated default name.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#container_name)",
				},
			},
		},
		{
			name:      "container_name attribute with a comment before it",
			content:   "services:\n  test:\n#\n    container_name: abc",
			line:      3,
			character: 7,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Specify a custom container name, rather than a generated default name.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#container_name)",
				},
			},
		},
		{
			name:      "container_name attribute with a comment before it (CRLF)",
			content:   "services:\r\n  test:\r\n#\r\n    container_name: abc",
			line:      3,
			character: 7,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Specify a custom container name, rather than a generated default name.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#container_name)",
				},
			},
		},
		{
			name: "develop attribute attribute from the services object",
			content: `
services:
  testService:
    develop:`,
			line:      3,
			character: 8,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Development configuration for the service, used for development workflows.\n\nSchema: [compose-spec.json](https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json)\n\n[Online documentation](https://docs.docker.com/reference/compose-file/services/#develop)",
				},
			},
		},
		{
			name: "service hover with whitespace around it",
			content: `
services:

  backend:
    image: hello

  frontend:
    depends_on:
      - backend`,
			line:      8,
			character: 12,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `backend:
  image: hello` +
						"\n```",
				},
			},
		},
		{
			name: "service hover with multiple lines of whitespace around it",
			content: `
services:



  backend:
    image: hello

  frontend:
    depends_on:
      - backend`,
			line:      10,
			character: 12,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `backend:
  image: hello` +
						"\n```",
				},
			},
		},
	}

	composeFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), uri.URI(composeFile), 1, []byte(tc.content))
			result, err := Hover(context.Background(), &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: composeFile},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, doc)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
		})
	}
}

func TestHover_InterFileSupport(t *testing.T) {
	testCases := []struct {
		name         string
		content      string
		otherContent string
		line         uint32
		character    uint32
		result       *protocol.Hover
	}{
		{
			name: "hovering over an extends service as a string",
			content: `
include:
  - compose.other.yaml
services:
  test2:
    image: alpine:3.21
    extends: test`,
			otherContent: `
services:
  test:
    image: alpine:3.20`,
			line:      6,
			character: 15,
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind: protocol.MarkupKindMarkdown,
					Value: "```YAML\n" + `test:
  image: alpine:3.20` +
						"\n```",
				},
			},
		},
	}

	composeFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	composeOtherFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.other.yaml")), "/"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := document.NewDocumentManager()
			changed, err := mgr.Write(context.Background(), uri.URI(composeOtherFile), protocol.DockerComposeLanguage, 1, []byte(tc.otherContent))
			require.NoError(t, err)
			require.True(t, changed)
			doc := document.NewComposeDocument(mgr, uri.URI(composeFile), 1, []byte(tc.content))
			result, err := Hover(context.Background(), &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: composeFile},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, doc)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
		})
	}
}
