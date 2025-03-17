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

func TestCompletion(t *testing.T) {
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
		name      string
		content   string
		line      uint32
		character uint32
		items     []protocol.CompletionItem
	}{
		{
			name:      "empty file",
			content:   "",
			line:      0,
			character: 0,
			items: []protocol.CompletionItem{
				{
					Detail:           types.CreateStringPointer("Block"),
					Kind:             types.CreateCompletionItemKindPointer(protocol.CompletionItemKindClass),
					Label:            "function",
					InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					TextEdit: protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{Line: 0, Character: 0},
							End:   protocol.Position{Line: 0, Character: 0},
						},
						NewText: "function \"${1:functionName}\" {\n  ${2}\n}",
					},
				},
				{
					Detail:           types.CreateStringPointer("Block"),
					Kind:             types.CreateCompletionItemKindPointer(protocol.CompletionItemKindClass),
					Label:            "group",
					InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					TextEdit: protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{Line: 0, Character: 0},
							End:   protocol.Position{Line: 0, Character: 0},
						},
						NewText: "group \"${1:groupName}\" {\n  ${2}\n}",
					},
				},
				{
					Detail:           types.CreateStringPointer("Block"),
					Kind:             types.CreateCompletionItemKindPointer(protocol.CompletionItemKindClass),
					Label:            "target",
					InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					TextEdit: protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{Line: 0, Character: 0},
							End:   protocol.Position{Line: 0, Character: 0},
						},
						NewText: "target \"${1:targetName}\" {\n  ${2}\n}",
					},
				},
				{
					Detail:           types.CreateStringPointer("Block"),
					Kind:             types.CreateCompletionItemKindPointer(protocol.CompletionItemKindClass),
					Label:            "variable",
					InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					TextEdit: protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{Line: 0, Character: 0},
							End:   protocol.Position{Line: 0, Character: 0},
						},
						NewText: "variable \"${1:variableName}\" {\n  ${2}\n}",
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

			var result protocol.CompletionList
			err = conn.Call(context.Background(), protocol.MethodTextDocumentCompletion, protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
					Position:     protocol.Position{Line: 0, Character: 0},
				},
			}, &result)
			require.NoError(t, err)
			require.False(t, result.IsIncomplete)
			require.Equal(t, tc.items, result.Items)
		})
	}
}
