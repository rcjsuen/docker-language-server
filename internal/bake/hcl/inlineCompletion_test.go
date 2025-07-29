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
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestInlineCompletion(t *testing.T) {
	testCases := []struct {
		name              string
		content           string
		dockerfileContent string
		position          protocol.Position
		items             []protocol.InlineCompletionItem
	}{
		{
			name:              "suggest nothing inside a comment",
			content:           "// 123",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 1},
			items:             nil,
		},
		{
			name:              "suggest nothing inside a block",
			content:           "target t {}",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 2},
			items:             nil,
		},
		{
			name:              "suggest nothing inside an attribute",
			content:           "attribute = \"value\"",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 2},
			items:             nil,
		},
		{
			name:              "suggest nothing inside attribute whitespace",
			content:           "attribute   = \"value\"",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 10},
			items:             nil,
		},
		{
			name:              "suggest nothing if the line just contains a brace",
			content:           "target t {\n}",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 1, Character: 1},
			items:             nil,
		},
		{
			name:              "one build stage",
			content:           "",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "multiple build stages",
			content:           "",
			dockerfileContent: "FROM scratch AS simple\nFROM scratch as complex",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
				{
					InsertText: "target \"complex\" {\n  target = \"complex\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "skip pre-existing block with shared name",
			content:           "\ntarget \"simple\" {}",
			dockerfileContent: "FROM scratch AS simple\nFROM scratch as complex",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"complex\" {\n  target = \"complex\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "skip pre-existing block with different name but correct target",
			content:           "\ntarget \"t\" { target = \"simple\" }",
			dockerfileContent: "FROM scratch AS simple\nFROM scratch as complex",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"complex\" {\n  target = \"complex\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "arg without default value",
			content:           "",
			dockerfileContent: "ARG beforeFrom\nFROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n  args = {\n    beforeFrom = \"\"\n  }\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "arg with default value",
			content:           "",
			dockerfileContent: "ARG beforeFrom=123\nFROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n  args = {\n    beforeFrom = \"123\"\n  }\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "multiple args",
			content:           "",
			dockerfileContent: "ARG beforeFrom=123\nARG another\nFROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n  args = {\n    beforeFrom = \"123\"\n    another = \"\"\n  }\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "ARG after FROM is ignored",
			content:           "",
			dockerfileContent: "FROM scratch AS check\nARG ignored=true",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"check\" {\n  target = \"check\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
		{
			name:              "t prefix",
			content:           "t",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 1},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 1},
					},
				},
			},
		},
		{
			name:              "t prefix with leading whitespace",
			content:           " t",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 2},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 1},
						End:   protocol.Position{Line: 0, Character: 2},
					},
				},
			},
		},
	}

	dockerfilePath := filepath.ToSlash(filepath.Join(os.TempDir(), "Dockerfile"))
	dockerBakePath := filepath.ToSlash(filepath.Join(os.TempDir(), "docker-bake.hcl"))
	temporaryDockerfile := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerfilePath, "/")))
	temporaryBakeFile := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerBakePath, "/")))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			if tc.dockerfileContent != "" {
				changed, err := manager.Write(context.Background(), temporaryDockerfile, protocol.DockerfileLanguage, 1, []byte(tc.dockerfileContent))
				require.NoError(t, err)
				require.True(t, changed)
			}
			doc := document.NewBakeHCLDocument(temporaryBakeFile, 1, []byte(tc.content))
			items, err := InlineCompletion(context.Background(), &protocol.InlineCompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: string(temporaryBakeFile)},
					Position:     tc.position,
				},
			}, manager, doc)
			require.NoError(t, err)
			require.Equal(t, tc.items, items)
		})
	}
}
func TestInlineCompletion_WSL(t *testing.T) {
	testCases := []struct {
		name              string
		content           string
		dockerfileContent string
		position          protocol.Position
		items             []protocol.InlineCompletionItem
	}{
		{
			name:              "one build stage",
			content:           "",
			dockerfileContent: "FROM scratch AS simple",
			position:          protocol.Position{Line: 0, Character: 0},
			items: []protocol.InlineCompletionItem{
				{
					InsertText: "target \"simple\" {\n  target = \"simple\"\n}\n",
					Range: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			},
		},
	}

	bakeURI := "file://wsl%24/docker-desktop/tmp/docker-bake.hcl"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			changed, err := manager.Write(context.Background(), "file://wsl%24/docker-desktop/tmp/Dockerfile", protocol.DockerfileLanguage, 1, []byte(tc.dockerfileContent))
			require.NoError(t, err)
			require.True(t, changed)
			doc := document.NewBakeHCLDocument(uri.URI(bakeURI), 1, []byte(tc.content))
			items, err := InlineCompletion(context.Background(), &protocol.InlineCompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: bakeURI},
					Position:     tc.position,
				},
			}, manager, doc)
			require.NoError(t, err)
			require.Equal(t, tc.items, items)
		})
	}
}
