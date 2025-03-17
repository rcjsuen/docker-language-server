package server_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestDocumentLink(t *testing.T) {
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

	testFolder := path.Join(homedir, t.Name())
	parentFolder, err := filepath.Abs(path.Join(testFolder, ".."))
	require.NoError(t, err)

	testCases := []struct {
		name    string
		content string
		links   []protocol.DocumentLink
	}{
		{
			name:    "dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 30},
					},
					Target:  types.CreateStringPointer(fmt.Sprintf("file://%v", path.Join(homedir, "TestDocumentLink", "Dockerfile.api"))),
					Tooltip: types.CreateStringPointer(path.Join(homedir, "TestDocumentLink", "Dockerfile.api")),
				},
			},
		},
		{
			name:    "./dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"./Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 32},
					},
					Target:  types.CreateStringPointer(fmt.Sprintf("file://%v", path.Join(homedir, "TestDocumentLink", "Dockerfile.api"))),
					Tooltip: types.CreateStringPointer(path.Join(homedir, "TestDocumentLink", "Dockerfile.api")),
				},
			},
		},
		{
			name:    "../dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 33},
					},
					Target:  types.CreateStringPointer(fmt.Sprintf("file://%v", path.Join(parentFolder, "Dockerfile.api"))),
					Tooltip: types.CreateStringPointer(path.Join(parentFolder, "Dockerfile.api")),
				},
			},
		},
		{
			name:    "folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"folder/Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 37},
					},
					Target:  types.CreateStringPointer(fmt.Sprintf("file://%v", path.Join(homedir, "TestDocumentLink", "folder", "Dockerfile.api"))),
					Tooltip: types.CreateStringPointer(path.Join(homedir, "TestDocumentLink", "folder", "Dockerfile.api")),
				},
			},
		},
		{
			name:    "../folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../folder/Dockerfile.api\"\n}",
			links: []protocol.DocumentLink{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 16},
						End:   protocol.Position{Line: 1, Character: 40},
					},
					Target:  types.CreateStringPointer(fmt.Sprintf("file://%v", path.Join(parentFolder, "folder/Dockerfile.api"))),
					Tooltip: types.CreateStringPointer(path.Join(parentFolder, "folder/Dockerfile.api")),
				},
			},
		},
	}

	didOpen := createDidOpenTextDocumentParams(testFolder, "DocumentLink.hcl", "", "dockerbake")
	err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didChange := createDidChangeTextDocumentParams(testFolder, "DocumentLink.hcl", tc.content)
			err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidChange, didChange)
			require.NoError(t, err)

			var result []protocol.DocumentLink
			err = conn.Call(context.Background(), protocol.MethodTextDocumentDocumentLink, protocol.DocumentLinkParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
			}, &result)
			require.NoError(t, err)
			require.Equal(t, tc.links, result)
		})
	}
}
