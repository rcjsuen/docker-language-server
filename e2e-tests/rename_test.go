package server_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
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

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		content       string
		position      protocol.Position
		workspaceEdit func(protocol.DocumentUri) *protocol.WorkspaceEdit
	}{
		{
			name: "rename dependent service",
			content: `
services:
  test:
    depends_on:
      - test2
  test2:`,
			position: protocol.Position{Line: 4, Character: 11},
			workspaceEdit: func(u protocol.DocumentUri) *protocol.WorkspaceEdit {
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+".yaml", tc.content, "dockercompose")
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var workspaceEdit *protocol.WorkspaceEdit
			err = conn.Call(context.Background(), protocol.MethodTextDocumentRename, protocol.RenameParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
					Position:     tc.position,
				},
				NewName: "newName",
			}, &workspaceEdit)
			require.NoError(t, err)
			require.Equal(t, tc.workspaceEdit(didOpen.TextDocument.URI), workspaceEdit)
		})
	}
}
