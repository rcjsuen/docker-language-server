package compose

import (
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

func TestDocumentHighlight_Services(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a service",
			content: `
services:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined service array item",
			content: `
services:
  test:
    depends_on:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined service object with no properties",
			content: `
services:
  test:
    depends_on:
      test2:`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined service object with properties",
			content: `
services:
  test:
    depends_on:
      test2:
        condition: service_started`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "cursor not on anything meaningful",
			content: `
services:
  test:
    depends_on:
      - test2
      - test2`,
			line:      3,
			character: 9,
			ranges:    nil,
		},
		{
			name: "read highlight on an undefined service array item, duplicated",
			content: `
services:
  test:
    depends_on:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a service array item",
			content: `
services:
  test:
    depends_on:
      - test2
  test2:`,
			line:      5,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 2, 5, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

func TestDocumentHighlight_Networks(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a network",
			content: `
networks:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined network array item",
			content: `
services:
  test:
    networks:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined network object with no properties",
			content: `
services:
  test:
    networks:
      test2:`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined network object with properties",
			content: `
services:
  test:
    networks:
      test2:
        priority: 0`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined networks array item, duplicated",
			content: `
services:
  test:
    networks:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a network array item (cursor on read)",
			content: `
services:
  test:
    networks:
      - test2
networks:
  test2:`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a network array item (cursor on write)",
			content: `
services:
  test:
    networks:
      - test2
networks:
  test2:`,
			line:      6,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a network object (read)",
			content: `
services:
  test:
    networks:
      test2:
        priority: 0
networks:
  test2:`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
				documentHighlight(7, 2, 7, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a network object (write)",
			content: `
services:
  test:
    networks:
      test2:
        priority: 0
networks:
  test2:`,
			line:      7,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
				documentHighlight(7, 2, 7, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

func TestDocumentHighlight_Volumes(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a volumes",
			content: `
volumes:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined volume array item",
			content: `
services:
  test:
    volumes:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined volume array item with a mount path",
			content: `
services:
  test:
    volumes:
      - test2:/mount/path`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined volume array item's mount path",
			content: `
services:
  test:
    volumes:
      - test2:/mount/path`,
			line:      4,
			character: 18,
			ranges:    nil,
		},
		{
			name: "read highlight on an invalid volume array item object",
			content: `
services:
  test:
    volumes:
      - source:`,
			line:      4,
			character: 10,
			ranges:    nil,
		},
		{
			name: "read highlight on an volume array item object's source",
			content: `
services:
  test:
    volumes:
      - source: test2`,
			line:      4,
			character: 18,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 16, 4, 21, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an volume array item object's target which is invalid",
			content: `
services:
  test:
    volumes:
      - target: test2`,
			line:      4,
			character: 18,
			ranges:    nil,
		},
		{
			name: "read highlight on an invalid volume object",
			content: `
services:
  test:
    volumes:
      test2:`,
			line:      4,
			character: 9,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 6, 4, 11, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an undefined volumes array item, duplicated",
			content: `
services:
  test:
    volumes:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a volume array item (cursor on read)",
			content: `
services:
  test:
    volumes:
      - test2
volumes:
  test2:`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a volume array item (cursor on write)",
			content: `
services:
  test:
    volumes:
      - test2
volumes:
  test2:`,
			line:      6,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

func TestDocumentHighlight_Configs(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a configs",
			content: `
configs:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined config array item",
			content: `
services:
  test:
    configs:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an invalid config object",
			content: `
services:
  test:
    configs:
      test2:`,
			line:      4,
			character: 9,
			ranges:    nil,
		},
		{
			name: "read highlight on an undefined configs array item, duplicated",
			content: `
services:
  test:
    configs:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a config array item (cursor on read)",
			content: `
services:
  test:
    configs:
      - test2
configs:
  test2:`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a config array item (cursor on write)",
			content: `
services:
  test:
    configs:
      - test2
configs:
  test2:`,
			line:      6,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

func TestDocumentHighlight_Secrets(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      protocol.UInteger
		character protocol.UInteger
		ranges    []protocol.DocumentHighlight
	}{
		{
			name: "write highlight on a secrets",
			content: `
secrets:
  test:`,
			line:      2,
			character: 4,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read highlight on an undefined secret array item",
			content: `
services:
  test:
    secrets:
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read highlight on an invalid secret object",
			content: `
services:
  test:
    secrets:
      test2:`,
			line:      4,
			character: 9,
			ranges:    nil,
		},
		{
			name: "read highlight on an undefined secrets array item, duplicated",
			content: `
services:
  test:
    secrets:
      - test2
      - test2`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(5, 8, 5, 13, protocol.DocumentHighlightKindRead),
			},
		},
		{
			name: "read/write highlight on a secret array item (cursor on read)",
			content: `
services:
  test:
    secrets:
      - test2
secrets:
  test2:`,
			line:      4,
			character: 10,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
		{
			name: "read/write highlight on a secret array item (cursor on write)",
			content: `
services:
  test:
    secrets:
      - test2
secrets:
  test2:`,
			line:      6,
			character: 5,
			ranges: []protocol.DocumentHighlight{
				documentHighlight(4, 8, 4, 13, protocol.DocumentHighlightKindRead),
				documentHighlight(6, 2, 6, 7, protocol.DocumentHighlightKindWrite),
			},
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}
