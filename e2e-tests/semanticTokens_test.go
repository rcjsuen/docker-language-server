package server_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestSemanticTokensFull(t *testing.T) {
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
		name    string
		content string
		result  []uint32
	}{
		{
			name:    "target {}",
			content: "target {}",
			result:  []uint32{0, 0, 6, hcl.SemanticTokenTypeIndex(hcl.TokenType_Type), 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+".hcl", tc.content, "dockerbake")
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var result protocol.SemanticTokens
			var unset protocol.SemanticTokens
			err = conn.Call(context.Background(), protocol.MethodTextDocumentSemanticTokensFull, protocol.SemanticTokensParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
			}, &result)
			require.NoError(t, err)
			require.Equal(t, unset.ResultID, result.ResultID)
			require.Equal(t, tc.result, result.Data)
		})
	}
}
