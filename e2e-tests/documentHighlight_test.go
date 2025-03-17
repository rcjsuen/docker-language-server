package server_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestDocumentHighlight(t *testing.T) {
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
		ranges   []*protocol.DocumentHighlight
	}{
		{
			name:     "cursor in group's block targets attribute pointing at unquoted target",
			content:  "group g { targets = [\"build\"] }\ntarget build {}\ntarget irrelevant {}",
			position: protocol.Position{Line: 0, Character: 25},
			ranges: []*protocol.DocumentHighlight{
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 22},
						End:   protocol.Position{Line: 0, Character: 27},
					},
				},
				{
					Kind: types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 7},
						End:   protocol.Position{Line: 1, Character: 12},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+".hcl", tc.content, "dockerbake")
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var ranges []*protocol.DocumentHighlight
			err = conn.Call(context.Background(), protocol.MethodTextDocumentDocumentHighlight, protocol.DocumentHighlightParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
					Position:     tc.position,
				},
			}, &ranges)
			require.NoError(t, err)
			require.Equal(t, tc.ranges, ranges)
		})
	}
}
