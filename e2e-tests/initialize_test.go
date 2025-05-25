package server_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/pkg/server"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
)

func init() {
	os.Setenv("DOCKER_LANGUAGE_SERVER_TELEMETRY", "false")
}

func createDidOpenTextDocumentParams(homedir, testName, text string, languageID protocol.LanguageIdentifier) protocol.DidOpenTextDocumentParams {
	return protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(homedir, testName)), "/"))),
			Text:       text,
			LanguageID: languageID,
			Version:    1,
		},
	}
}

func createDidChangeTextDocumentParams(homedir, testName, text string, version int32) protocol.DidChangeTextDocumentParams {
	return protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(homedir, testName)), "/"))),
			},
			Version: version,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEvent{
				Text: text,
			},
		},
	}
}

func TestInitializeFunctionIsolatedCall(t *testing.T) {
	testCases := []struct {
		name             string
		params           protocol.InitializeParams
		workspaceFolders []string
	}{
		{
			name: "one workspace folder",
			params: protocol.InitializeParams{
				WorkspaceFolders: []protocol.WorkspaceFolder{{Name: "abc", URI: "uri:///a/b/c"}},
			},
			workspaceFolders: []string{"/a/b/c"},
		},
		{
			name: "two workspace folders",
			params: protocol.InitializeParams{
				WorkspaceFolders: []protocol.WorkspaceFolder{
					{Name: "abc", URI: "uri:///a/b/c"},
					{Name: "def", URI: "uri:///d/e/f"},
				},
			},
			workspaceFolders: []string{"/a/b/c", "/d/e/f"},
		},
		{
			name:             "rootUri",
			params:           protocol.InitializeParams{RootURI: types.CreateStringPointer("uri:///a/b/c")},
			workspaceFolders: []string{"/a/b/c"},
		},
		{
			name:             "rootPath",
			params:           protocol.InitializeParams{RootPath: types.CreateStringPointer("/a/b/c")},
			workspaceFolders: []string{"/a/b/c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testInitialize(t, tc.params)
			require.Equal(t, tc.workspaceFolders, s.WorkspaceFolders())
		})
	}
}

func testInitialize(t *testing.T, initializeParams protocol.InitializeParams) *server.Server {
	s := startServer()
	_, err := s.Initialize(&glsp.Context{}, &initializeParams)
	require.NoError(t, err)
	return s
}

type TestStream struct {
	incoming *bytes.Buffer
	outgoing *bytes.Buffer
	closed   bool
}

func (ts *TestStream) Read(b []byte) (int, error) {
	for {
		if ts.closed {
			return 0, io.EOF
		}

		r, err := ts.incoming.Read(b)
		if r > 0 {
			return r, err
		} else if err != io.EOF {
			return r, err
		}
		time.Sleep(1 * time.Second)
	}
}

func (ts *TestStream) Write(n []byte) (int, error) {
	return ts.outgoing.Write(n)
}

func (ts *TestStream) Close() error {
	ts.closed = true
	return nil
}

func startServer() *server.Server {
	docManager := document.NewDocumentManager()
	s := server.NewServer(docManager)
	return s
}
