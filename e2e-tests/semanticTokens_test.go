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
		name               string
		languageIdentifier protocol.LanguageIdentifier
		content            string
		result             *protocol.SemanticTokens
	}{
		{
			name:               "target {}",
			languageIdentifier: protocol.DockerBakeLanguage,
			content:            "target {}",
			result:             &protocol.SemanticTokens{Data: []uint32{0, 0, 6, hcl.SemanticTokenTypeIndex(hcl.TokenType_Type), 0}},
		},
		{
			name:               "single line comment after content with no newlines after it",
			languageIdentifier: protocol.DockerBakeLanguage,
			content:            "variable \"port\" {default = true} # hello",
			result: &protocol.SemanticTokens{
				Data: []uint32{
					0, 0, 8, hcl.SemanticTokenTypeIndex(hcl.TokenType_Type), 0,
					0, 9, 6, hcl.SemanticTokenTypeIndex(hcl.TokenType_Class), 0,
					0, 8, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Property), 0,
					0, 10, 4, hcl.SemanticTokenTypeIndex(hcl.TokenType_Keyword), 0,
					0, 6, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Comment), 0,
				},
			},
		},
		{
			name:               "single line comment after content followed by LF",
			languageIdentifier: protocol.DockerBakeLanguage,
			content:            "variable \"port\" {default = true} # hello\n",
			result: &protocol.SemanticTokens{
				Data: []uint32{
					0, 0, 8, hcl.SemanticTokenTypeIndex(hcl.TokenType_Type), 0,
					0, 9, 6, hcl.SemanticTokenTypeIndex(hcl.TokenType_Class), 0,
					0, 8, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Property), 0,
					0, 10, 4, hcl.SemanticTokenTypeIndex(hcl.TokenType_Keyword), 0,
					0, 6, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Comment), 0,
				},
			},
		},
		{
			name:               "single line comment after content followed by CRLF",
			languageIdentifier: protocol.DockerBakeLanguage,
			content:            "variable \"port\" {default = true} # hello\r\n",
			result: &protocol.SemanticTokens{
				Data: []uint32{
					0, 0, 8, hcl.SemanticTokenTypeIndex(hcl.TokenType_Type), 0,
					0, 9, 6, hcl.SemanticTokenTypeIndex(hcl.TokenType_Class), 0,
					0, 8, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Property), 0,
					0, 10, 4, hcl.SemanticTokenTypeIndex(hcl.TokenType_Keyword), 0,
					0, 6, 7, hcl.SemanticTokenTypeIndex(hcl.TokenType_Comment), 0,
				},
			},
		},
		{
			name:               "open dockerfile.hcl (issue 84)",
			languageIdentifier: protocol.DockerfileLanguage,
			content:            "FROM scratch",
			result:             nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+".hcl", tc.content, tc.languageIdentifier)
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var result *protocol.SemanticTokens
			err = conn.Call(context.Background(), protocol.MethodTextDocumentSemanticTokensFull, protocol.SemanticTokensParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
			}, &result)
			require.NoError(t, err)
			require.Equal(t, tc.result, result)
		})
	}
}
