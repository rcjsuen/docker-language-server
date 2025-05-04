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

func TestPrepareRename(t *testing.T) {
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
		name     string
		content  string
		position protocol.Position
		result   *protocol.Range
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
			result: &protocol.Range{
				Start: protocol.Position{Line: 4, Character: 8},
				End:   protocol.Position{Line: 4, Character: 13},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+".yaml", tc.content, "dockercompose")
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var result *protocol.Range
			err = conn.Call(context.Background(), protocol.MethodTextDocumentPrepareRename, protocol.PrepareRenameParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
					Position:     tc.position,
				},
			}, &result)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
		})
	}
}
