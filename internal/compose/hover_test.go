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
					Kind:  protocol.MarkupKindPlainText,
					Value: "declared for backward compatibility, ignored.",
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
					Kind:  protocol.MarkupKindPlainText,
					Value: "define the Compose project name, until user defines one explicitly.",
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
					Kind:  protocol.MarkupKindPlainText,
					Value: "compose sub-projects to be included.",
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
					Value: "Allowed values:\n- `bind`\n- `cluster`\n- `image`\n- `npipe`\n- `tmpfs`\n- `volume`\n",
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
					Value: "Allowed values:\n- `Z`\n- `z`\n",
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
					Value: "Allowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n",
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
					Value: "Allowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n",
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
					Value: "Allowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n",
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
					Value: "Allowed values:\n- `disabled`\n- `enabled`\n- `readonly`\n- `writable`\n",
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
					Value: "Allowed values:\n- `host`\n- `private`\n",
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
					Value: "Allowed values:\n- `service_completed_successfully`\n- `service_healthy`\n- `service_started`\n",
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
					Value: "Allowed values:\n- `rebuild`\n- `restart`\n- `sync`\n- `sync+exec`\n- `sync+restart`\n",
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
					Value: "Allowed values:\n- `start-first`\n- `stop-first`\n",
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
					Value: "Allowed values:\n- `start-first`\n- `stop-first`\n",
				},
			},
		},
	}

	temporaryBakeFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(uri.URI(temporaryBakeFile), 1, []byte(tc.content))
			result, err := Hover(context.Background(), &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: temporaryBakeFile},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, doc)
			require.NoError(t, err)
			if tc.result == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Nil(t, result.Range)
				markupContent, ok := result.Contents.(protocol.MarkupContent)
				require.True(t, ok)
				require.Equal(t, tc.result.Contents, markupContent)
			}
		})
	}
}
