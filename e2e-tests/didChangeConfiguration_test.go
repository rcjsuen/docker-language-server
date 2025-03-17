package server_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func createDidChangeConfiguration() protocol.DidChangeConfigurationParams {
	return protocol.DidChangeConfigurationParams{
		Settings: []string{"docker.lsp.experimental.vulnerabilityScanning"},
	}
}

func TestDidChangeConfiguration(t *testing.T) {
	s := startServer()

	client := bytes.NewBuffer(make([]byte, 0, 1024))
	server := bytes.NewBuffer(make([]byte, 0, 1024))
	serverStream := &TestStream{incoming: server, outgoing: client, closed: false}
	defer serverStream.Close()
	go s.ServeStream(serverStream)

	clientStream := jsonrpc2.NewBufferedStream(&TestStream{incoming: client, outgoing: server, closed: false}, jsonrpc2.VSCodeObjectCodec{})
	defer clientStream.Close()
	handler := &ConfigurationHandler{t: t, scanning: false}
	conn := jsonrpc2.NewConn(context.Background(), clientStream, handler)
	initialize(t, conn, protocol.InitializeParams{})

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	didOpen := createDidOpenTextDocumentParams(homedir, "Dockerfile", "FROM scratch", protocol.DockerfileLanguage)
	err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
	require.NoError(t, err)

	for configuration.Get(didOpen.TextDocument.URI).Experimental.VulnerabilityScanning {
		fmt.Fprintln(os.Stderr, "Sleeping...")
		time.Sleep(100 * time.Millisecond)
	}

	handler.scanning = true
	didChangeConfiguration := createDidChangeConfiguration()
	err = conn.Notify(context.Background(), protocol.MethodWorkspaceDidChangeConfiguration, didChangeConfiguration)
	require.NoError(t, err)
	for !configuration.Get(didOpen.TextDocument.URI).Experimental.VulnerabilityScanning {
		time.Sleep(100 * time.Millisecond)
	}
	require.True(t, configuration.Get(didOpen.TextDocument.URI).Experimental.VulnerabilityScanning)
}
