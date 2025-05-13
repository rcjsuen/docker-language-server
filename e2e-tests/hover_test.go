package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func HandleConfiguration(t *testing.T, conn *jsonrpc2.Conn, request *jsonrpc2.Request, scanning bool) {
	var configurationParams protocol.ConfigurationParams
	err := json.Unmarshal(*request.Params, &configurationParams)
	require.NoError(t, err)
	configurations := []configuration.Configuration{}
	for range configurationParams.Items {
		configurations = append(
			configurations,
			configuration.Configuration{
				Experimental: configuration.Experimental{
					VulnerabilityScanning: scanning,
					Scout: configuration.Scout{
						CriticalHighVulnerabilities: true,
						NotPinnedDigest:             true,
						RecommendedTag:              true,
						Vulnerabilites:              true,
					},
				},
			},
		)
	}
	require.NoError(t, conn.Reply(context.Background(), request.ID, configurations))
}

type ConfigurationHandler struct {
	t        *testing.T
	scanning bool
}

func (h *ConfigurationHandler) Handle(_ context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) {
	switch request.Method {
	case protocol.ServerWorkspaceConfiguration:
		if !request.Notif && request.Params != nil {
			HandleConfiguration(h.t, conn, request, h.scanning)
		}
	}
}

func TestHover(t *testing.T) {
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
		languageID          protocol.LanguageIdentifier
		fileExtensionSuffix string
		name                string
		content             string
		position            protocol.Position
		result              *protocol.Hover
	}{
		{
			languageID:          protocol.DockerBakeLanguage,
			fileExtensionSuffix: ".hcl",
			name:                "hover over target block type",
			content:             "target \"api\" {}",
			position:            protocol.Position{Line: 0, Character: 3},
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "**target** _Block_\n\nA target reflects a single `docker build` invocation.",
				},
			},
		},
		{
			languageID:          protocol.DockerfileLanguage,
			fileExtensionSuffix: "",
			name:                "hover over alpine:3.16.1",
			content:             "FROM alpine:3.16.1",
			position:            protocol.Position{Line: 0, Character: 8},
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "Current image vulnerabilities:   1C   3H   9M   0L \r\n\r\nRecommended tags:\n\n<table>\n<tr><td><code>3.21.3</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n<tr><td><code>3.20.6</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n<tr><td><code>3.18.12</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n</table>\n",
				},
			},
		},
		{
			languageID:          protocol.DockerComposeLanguage,
			fileExtensionSuffix: ".yaml",
			name:                "version description",
			content:             "version: 1.2.3",
			position:            protocol.Position{Line: 0, Character: 4},
			result: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindPlainText,
					Value: "declared for backward compatibility, ignored. Please remove it.",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name()+tc.fileExtensionSuffix, tc.content, tc.languageID)
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			var hover *protocol.Hover
			err = conn.Call(context.Background(), protocol.MethodTextDocumentHover, protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
					Position:     protocol.Position{Line: 0, Character: 3},
				},
			}, &hover)
			require.NoError(t, err)
			require.Equal(t, tc.result, hover)
		})
	}

}
