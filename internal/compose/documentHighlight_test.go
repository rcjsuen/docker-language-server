package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func documentHighlight(startLine, startCharacter, endLine, endCharacter protocol.UInteger, kind protocol.DocumentHighlightKind) protocol.DocumentHighlight {
	return protocol.DocumentHighlight{
		Kind: &kind,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      startLine,
				Character: startCharacter,
			},
			End: protocol.Position{
				Line:      endLine,
				Character: endCharacter,
			},
		},
	}
}

var serviceReferenceTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 6},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
		},
	},
	{
		name: "read highlight on an undefined quoted service array item",
		content: `
services:
  test:
    depends_on:
      - "test2"`,
		line:      4,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 9, 4, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 9},
								End:   protocol.Position{Line: 4, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 9},
			End:   protocol.Position{Line: 4, Character: 14},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 8},
								End:   protocol.Position{Line: 5, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 2},
								End:   protocol.Position{Line: 5, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 2},
			End:   protocol.Position{Line: 5, Character: 7},
		},
	},
	{
		name: "extends as a string attribute",
		content: `
services:
  test:
    image: alpine
  test2:
    extends: test`,
		line:      5,
		character: 15,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			documentHighlight(5, 13, 5, 17, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 13},
								End:   protocol.Position{Line: 5, Character: 17},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 13},
			End:   protocol.Position{Line: 5, Character: 17},
		},
	},
	{
		name: "extends as a quoted string attribute",
		content: `
services:
  test:
    image: alpine
  test2:
    extends: "test"`,
		line:      5,
		character: 15,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			documentHighlight(5, 14, 5, 18, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 14},
								End:   protocol.Position{Line: 5, Character: 18},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 14},
			End:   protocol.Position{Line: 5, Character: 18},
		},
	},
	{
		name: "extends as an object without a file attribute",
		content: `
services:
  test:
    image: alpine
  test2:
    extends:
      service: test`,
		line:      6,
		character: 17,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 2, 2, 6, protocol.DocumentHighlightKindWrite),
			documentHighlight(6, 15, 6, 19, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 15},
								End:   protocol.Position{Line: 6, Character: 19},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 15},
			End:   protocol.Position{Line: 6, Character: 19},
		},
	},
	{
		name: "extends as an object without a file attribute",
		content: `
services:
  test:
    image: alpine
  test2:
    extends:
      service: test
      file: non-existent.yaml`,
		line:      6,
		character: 17,
		ranges:    nil,
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
	},
}

func TestDocumentHighlight_Services(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range serviceReferenceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			slices.SortFunc(ranges, func(a, b protocol.DocumentHighlight) int {
				return int(a.Range.Start.Line) - int(b.Range.Start.Line)
			})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

var networkReferenceTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 6},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 8},
								End:   protocol.Position{Line: 5, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 2},
			End:   protocol.Position{Line: 6, Character: 7},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 2},
								End:   protocol.Position{Line: 7, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 2},
								End:   protocol.Position{Line: 7, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 7, Character: 2},
			End:   protocol.Position{Line: 7, Character: 7},
		},
	},
}

func TestDocumentHighlight_Networks(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range networkReferenceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

var volumeReferenceTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 6},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
		},
	},
	{
		name: "read highlight on an undefined volume array item with a mount path that is quoted",
		content: `
services:
  test:
    volumes:
      - "test2:/mount/path"`,
		line:      4,
		character: 11,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 9, 4, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 9},
								End:   protocol.Position{Line: 4, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 9},
			End:   protocol.Position{Line: 4, Character: 14},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 16},
								End:   protocol.Position{Line: 4, Character: 21},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 16},
			End:   protocol.Position{Line: 4, Character: 21},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 6},
								End:   protocol.Position{Line: 4, Character: 11},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 6},
			End:   protocol.Position{Line: 4, Character: 11},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 8},
								End:   protocol.Position{Line: 5, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 2},
			End:   protocol.Position{Line: 6, Character: 7},
		},
	},
}

func TestDocumentHighlight_Volumes(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range volumeReferenceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

var configReferenceTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 6},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 8},
								End:   protocol.Position{Line: 5, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 2},
			End:   protocol.Position{Line: 6, Character: 7},
		},
	},
}

func TestDocumentHighlight_Configs(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range configReferenceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

var secretReferenceTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 2},
								End:   protocol.Position{Line: 2, Character: 6},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 6},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 8},
								End:   protocol.Position{Line: 5, Character: 13},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 8},
			End:   protocol.Position{Line: 4, Character: 13},
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
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 8},
								End:   protocol.Position{Line: 4, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 2},
								End:   protocol.Position{Line: 6, Character: 7},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 2},
			End:   protocol.Position{Line: 6, Character: 7},
		},
	},
}

func TestDocumentHighlight_Secrets(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range secretReferenceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}

var fragmentTestCases = []struct {
	name          string
	content       string
	line          protocol.UInteger
	character     protocol.UInteger
	ranges        []protocol.DocumentHighlight
	renameEdits   func(protocol.DocumentUri) *protocol.WorkspaceEdit
	prepareRename *protocol.Range
}{
	{
		name: "anchor with no alias",
		content: `
volumes:
  db-data: &default-volume
    driver: custom`,
		line:      2,
		character: 17,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 12, 2, 26, protocol.DocumentHighlightKindWrite),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 12},
								End:   protocol.Position{Line: 2, Character: 26},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 12},
			End:   protocol.Position{Line: 2, Character: 26},
		},
	},
	{
		name: "anchor with alias pointing at the anchor",
		content: `
volumes:
  db-data: &default-volume
    driver: default
  metrics: *default-volume`,
		line:      2,
		character: 17,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 12, 2, 26, protocol.DocumentHighlightKindWrite),
			documentHighlight(4, 12, 4, 26, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 12},
								End:   protocol.Position{Line: 2, Character: 26},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 12},
								End:   protocol.Position{Line: 4, Character: 26},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 12},
			End:   protocol.Position{Line: 2, Character: 26},
		},
	},
	{
		name: "anchor with alias pointing at the second alias",
		content: `
volumes:
  db-data: &default-volume
    driver: default
  metrics: *default-volume
  another: *default-volume`,
		line:      5,
		character: 17,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 12, 2, 26, protocol.DocumentHighlightKindWrite),
			documentHighlight(4, 12, 4, 26, protocol.DocumentHighlightKindRead),
			documentHighlight(5, 12, 5, 26, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 12},
								End:   protocol.Position{Line: 2, Character: 26},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 12},
								End:   protocol.Position{Line: 4, Character: 26},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 12},
								End:   protocol.Position{Line: 5, Character: 26},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 12},
			End:   protocol.Position{Line: 5, Character: 26},
		},
	},
	{
		name: "hovering over nothing",
		content: `
volumes:
  db-data: &default-volume
    driver: default
  metrics: *default-volume`,
		line:      4,
		character: 0,
		ranges:    nil,
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return nil
		},
		prepareRename: nil,
	},
	{
		name: "read reference on the first duplicated alias",
		content: `
services:
  serviceA:
    image: &redis redis:8-alpine
  serviceB:
    image: *redis
  serviceC:
    image: &redis redis:7-alpine
  serviceD:
    image: *redis`,
		line:      5,
		character: 14,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(3, 12, 3, 17, protocol.DocumentHighlightKindWrite),
			documentHighlight(5, 12, 5, 17, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 3, Character: 12},
								End:   protocol.Position{Line: 3, Character: 17},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 12},
								End:   protocol.Position{Line: 5, Character: 17},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 5, Character: 12},
			End:   protocol.Position{Line: 5, Character: 17},
		},
	},
	{
		name: "read reference on the second duplicated alias",
		content: `
services:
  serviceA:
    image: &redis redis:8-alpine
  serviceB:
    image: *redis
  serviceC:
    image: &redis redis:7-alpine
  serviceD:
    image: *redis`,
		line:      9,
		character: 14,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(7, 12, 7, 17, protocol.DocumentHighlightKindWrite),
			documentHighlight(9, 12, 9, 17, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 12},
								End:   protocol.Position{Line: 7, Character: 17},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 9, Character: 12},
								End:   protocol.Position{Line: 9, Character: 17},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 9, Character: 12},
			End:   protocol.Position{Line: 9, Character: 17},
		},
	},
	{
		name: "multiple anchors",
		content: `
services:
  serviceA:
    image: &redis8 redis:8-alpine
  serviceB:
    image: *redis8
  serviceC:
    image: &redis7 redis:7-alpine
  serviceD:
    image: *redis7`,
		line:      3,
		character: 14,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(3, 12, 3, 18, protocol.DocumentHighlightKindWrite),
			documentHighlight(5, 12, 5, 18, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 3, Character: 12},
								End:   protocol.Position{Line: 3, Character: 18},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 5, Character: 12},
								End:   protocol.Position{Line: 5, Character: 18},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 3, Character: 12},
			End:   protocol.Position{Line: 3, Character: 18},
		},
	},
	{
		name: "anchor in an array, write reference",
		content: `
services:
  serviceA:
    labels:
      - &label a.b.c=value
  serviceB:
    labels:
      - *label`,
		line:      4,
		character: 11,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 9, 4, 14, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 9, 7, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 9},
								End:   protocol.Position{Line: 4, Character: 14},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 9},
								End:   protocol.Position{Line: 7, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 9},
			End:   protocol.Position{Line: 4, Character: 14},
		},
	},
	{
		name: "anchor in an array, read reference",
		content: `
services:
  serviceA:
    labels:
      - &label a.b.c=value
  serviceB:
    labels:
      - *label`,
		line:      7,
		character: 11,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 9, 4, 14, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 9, 7, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 9},
								End:   protocol.Position{Line: 4, Character: 14},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 9},
								End:   protocol.Position{Line: 7, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 7, Character: 9},
			End:   protocol.Position{Line: 7, Character: 14},
		},
	},
	{
		name: "anchor in an object inside an array, read reference",
		content: `
services:
  backend:
    volumes:
      - type: &volumeType bind
        source: vol
        target: /data
      - type: *volumeType
        source: vol
        target: /data2`,
		line:      7,
		character: 20,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 15, 4, 25, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 15, 7, 25, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 15},
								End:   protocol.Position{Line: 4, Character: 25},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 15},
								End:   protocol.Position{Line: 7, Character: 25},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 7, Character: 15},
			End:   protocol.Position{Line: 7, Character: 25},
		},
	},
	{
		name: "anchor/alias references all on the same line",
		content: `
services:
  frontend:
    build:
      tags: [&keys aa, *keys, &keys bb, *keys, &keys cc, *keys]`,
		line:      4,
		character: 43,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 31, 4, 35, protocol.DocumentHighlightKindWrite),
			documentHighlight(4, 41, 4, 45, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 31},
								End:   protocol.Position{Line: 4, Character: 35},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 41},
								End:   protocol.Position{Line: 4, Character: 45},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 41},
			End:   protocol.Position{Line: 4, Character: 45},
		},
	},
	{
		name: "anchor/alias references are staggered",
		content: `
services:
  frontend:
    build:
      tags: [&keys aa]
  backend:
    build:
      tags: [*keys, &keys bb, *keys]`,
		line:      7,
		character: 16,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 14, 4, 18, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 14, 7, 18, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 14},
								End:   protocol.Position{Line: 4, Character: 18},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 14},
								End:   protocol.Position{Line: 7, Character: 18},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 7, Character: 14},
			End:   protocol.Position{Line: 7, Character: 18},
		},
	},
	{
		name: "duplicated anchor/alias references all on the same line",
		content: `
services:
  frontend:
    build:
      tags: [&keys ab, *keys]
  backend:
    build:
      tags: [&keys ab, *keys]`,
		line:      4,
		character: 16,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 14, 4, 18, protocol.DocumentHighlightKindWrite),
			documentHighlight(4, 24, 4, 28, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 14},
								End:   protocol.Position{Line: 4, Character: 18},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 24},
								End:   protocol.Position{Line: 4, Character: 28},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 14},
			End:   protocol.Position{Line: 4, Character: 18},
		},
	},
	{
		name: "interweaving fragments on the first anchor",
		content: `
services:
  test: &test
    image: alpine:3.22
  test2: &testAgain
    image: alpine:3.21
  test3: *test
  test4: *testAgain`,
		line:      2,
		character: 11,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 9, 2, 13, protocol.DocumentHighlightKindWrite),
			documentHighlight(6, 10, 6, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 9},
								End:   protocol.Position{Line: 2, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 10},
								End:   protocol.Position{Line: 6, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 13},
		},
	},
	{
		name: "interweaving fragments on the first alias",
		content: `
services:
  test: &test
    image: alpine:3.22
  test2: &testAgain
    image: alpine:3.21
  test3: *test
  test4: *testAgain`,
		line:      6,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 9, 2, 13, protocol.DocumentHighlightKindWrite),
			documentHighlight(6, 10, 6, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 9},
								End:   protocol.Position{Line: 2, Character: 13},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 6, Character: 10},
								End:   protocol.Position{Line: 6, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 6, Character: 10},
			End:   protocol.Position{Line: 6, Character: 14},
		},
	},
	{
		name: "interweaving fragments on the second anchor",
		content: `
services:
  test: &test
    image: alpine:3.22
  test2: &testAgain
    image: alpine:3.21
  test3: *test
  test4: *testAgain`,
		line:      4,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 10, 4, 19, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 10, 7, 19, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 10},
								End:   protocol.Position{Line: 4, Character: 19},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 10},
								End:   protocol.Position{Line: 7, Character: 19},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 10},
			End:   protocol.Position{Line: 4, Character: 19},
		},
	},
	{
		name: "interweaving fragments on the second alias",
		content: `
services:
  test: &test
    image: alpine:3.22
  test2: &testAgain
    image: alpine:3.21
  test3: *test
  test4: *testAgain`,
		line:      7,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 10, 4, 19, protocol.DocumentHighlightKindWrite),
			documentHighlight(7, 10, 7, 19, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 10},
								End:   protocol.Position{Line: 4, Character: 19},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 7, Character: 10},
								End:   protocol.Position{Line: 7, Character: 19},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 7, Character: 10},
			End:   protocol.Position{Line: 7, Character: 19},
		},
	},
	{
		name: "aliases with no matching anchor",
		content: `
services:
  test1: *test
  test2: *test
  test3: *test`,
		line:      2,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 10, 2, 14, protocol.DocumentHighlightKindRead),
			documentHighlight(3, 10, 3, 14, protocol.DocumentHighlightKindRead),
			documentHighlight(4, 10, 4, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 10},
								End:   protocol.Position{Line: 2, Character: 14},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 3, Character: 10},
								End:   protocol.Position{Line: 3, Character: 14},
							},
						},
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 10},
								End:   protocol.Position{Line: 4, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 10},
			End:   protocol.Position{Line: 2, Character: 14},
		},
	},
	{
		name: "alias before the actual anchor",
		content: `
services:
  test1: *test
  test2: &test
  test3: *test`,
		line:      2,
		character: 12,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(2, 10, 2, 14, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 2, Character: 10},
								End:   protocol.Position{Line: 2, Character: 14},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 2, Character: 10},
			End:   protocol.Position{Line: 2, Character: 14},
		},
	},
	{
		name: "alias before the actual anchor on the same line",
		content: `
services:
  test:
    build:
      tags: [*test, &test t1, *test]`,
		line:      4,
		character: 16,
		ranges: []protocol.DocumentHighlight{
			documentHighlight(4, 14, 4, 18, protocol.DocumentHighlightKindRead),
		},
		renameEdits: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
			return &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					u: {
						{
							NewText: "newName",
							Range: protocol.Range{
								Start: protocol.Position{Line: 4, Character: 14},
								End:   protocol.Position{Line: 4, Character: 18},
							},
						},
					},
				},
			}
		},
		prepareRename: &protocol.Range{
			Start: protocol.Position{Line: 4, Character: 14},
			End:   protocol.Position{Line: 4, Character: 18},
		},
	},
}

func TestDocumentHighlight_Fragments(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))
	u := uri.URI(composeFileURI)
	for _, tc := range fragmentTestCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(document.NewDocumentManager(), u, 1, []byte(tc.content))
			ranges, err := DocumentHighlight(doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}
