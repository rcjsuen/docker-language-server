package server_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestInlayHint(t *testing.T) {
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
		dockerfileContent string
		rng               protocol.Range
		hints             []protocol.InlayHint
	}{
		{
			name:              "args lookup",
			content:           "target t1 {\n  args = {\n    undefined = \"test\"\n    empty = \"test\"\n    defined = \"test\"\n}\n}",
			dockerfileContent: "FROM scratch\nARG undefined\nARG empty=\nARG defined=value\n",
			rng: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 5, Character: 0},
			},
			hints: []protocol.InlayHint{
				{
					Label:       "(default value: value)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 4, Character: 20},
				},
			},
		},
	}

	temporaryDockerfile := fmt.Sprintf("file://%v", path.Join(os.TempDir(), "Dockerfile"))
	temporaryBakeFile := fmt.Sprintf("file://%v", path.Join(os.TempDir(), "docker-bake.hcl"))

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

			var hints []protocol.InlayHint
			err = conn.Call(context.Background(), protocol.MethodTextDocumentInlayHint, protocol.InlayHintParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: didOpenBakeFile.TextDocument.URI},
				Range:        tc.rng,
			}, &hints)
			require.NoError(t, err)
			require.Equal(t, tc.hints, hints)
		})
	}
}
