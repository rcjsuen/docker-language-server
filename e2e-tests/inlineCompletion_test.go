package server_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestInlineCompletion(t *testing.T) {
	s := startServer()

	client := bytes.NewBuffer(make([]byte, 0, 1024))
	server := bytes.NewBuffer(make([]byte, 0, 1024))
	serverStream := &TestStream{incoming: server, outgoing: client, closed: false}
	defer serverStream.Close()
	go s.ServeStream(serverStream)

	clientStream := jsonrpc2.NewBufferedStream(&TestStream{incoming: client, outgoing: server, closed: false}, jsonrpc2.VSCodeObjectCodec{})
	defer clientStream.Close()
	conn := jsonrpc2.NewConn(context.Background(), clientStream, &ConfigurationHandler{t: t})
	initialize(t, conn, protocol.InitializeParams{})

	testCases := []struct {
		name              string
		content           string
		line              protocol.UInteger
		character         protocol.UInteger
		dockerfileContent string
		items             []protocol.InlineCompletionItem
	}{
		{
			name:              "one build stage",
			content:           "",
			line:              0,
			character:         0,
			dockerfileContent: "FROM scratch AS simple",
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
			name:              "outside document bounds",
			content:           "",
			line:              1,
			character:         0,
			dockerfileContent: "FROM scratch AS simple",
			items:             nil,
		},
	}

	temporaryDockerfile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "Dockerfile")), "/"))
	temporaryBakeFile := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "docker-bake.hcl")), "/"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpenDockerfile := protocol.DidOpenTextDocumentParams{
				TextDocument: protocol.TextDocumentItem{
					URI:        temporaryDockerfile,
					Text:       tc.dockerfileContent,
					LanguageID: "dockerfile",
					Version:    1,
				},
			}
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpenDockerfile)
			require.NoError(t, err)

			didOpenBakeFile := protocol.DidOpenTextDocumentParams{
				TextDocument: protocol.TextDocumentItem{
					URI:        temporaryBakeFile,
					Text:       tc.content,
					LanguageID: "dockerbake",
					Version:    1,
				},
			}
			err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpenBakeFile)
			require.NoError(t, err)

			var result []protocol.InlineCompletionItem
			err = conn.Call(context.Background(), protocol.MethodTextDocumentInlineCompletion, protocol.InlineCompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpenBakeFile.TextDocument.URI},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, &result)
			require.NoError(t, err)
			require.Equal(t, tc.items, result)
		})
	}
}
