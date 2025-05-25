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
	testFolder := filepath.Join(homedir, t.Name())

	testCases := []struct {
		name      string
		content   string
		linkRange protocol.Range
		path      string
		links     []protocol.DocumentLink
	}{
		{
			name:    "dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"Dockerfile.api\"\n}",
			path:    filepath.Join(homedir, "TestDocumentLink", "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 30},
			},
		},
		{
			name:    "./dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"./Dockerfile.api\"\n}",
			path:    filepath.Join(homedir, "TestDocumentLink", "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 32},
			},
		},
		{
			name:    "../dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../Dockerfile.api\"\n}",
			path:    filepath.Join(homedir, "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 33},
			},
		},
		{
			name:    "folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"folder/Dockerfile.api\"\n}",
			path:    filepath.Join(homedir, "TestDocumentLink", "folder", "Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 37},
			},
		},
		{
			name:    "../folder/dockerfile attribute in targets block",
			content: "target \"api\" {\n  dockerfile = \"../folder/Dockerfile.api\"\n}",
			path:    filepath.Join(homedir, "folder/Dockerfile.api"),
			linkRange: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 16},
				End:   protocol.Position{Line: 1, Character: 40},
			},
		},
	}

	didOpen := createDidOpenTextDocumentParams(testFolder, "DocumentLink.hcl", "", "dockerbake")
	err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
	require.NoError(t, err)

	version := int32(2)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			didChange := createDidChangeTextDocumentParams(testFolder, "DocumentLink.hcl", tc.content, version)
			version++
			err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidChange, didChange)
			require.NoError(t, err)

			var result []protocol.DocumentLink
			err = conn.Call(context.Background(), protocol.MethodTextDocumentDocumentLink, protocol.DocumentLinkParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: didOpen.TextDocument.URI},
			}, &result)
			require.NoError(t, err)
			links := []protocol.DocumentLink{
				{
					Range:   tc.linkRange,
					Target:  types.CreateStringPointer(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(tc.path), "/"))),
					Tooltip: types.CreateStringPointer(tc.path),
				},
			}
			require.Equal(t, links, result)
		})
	}
}
