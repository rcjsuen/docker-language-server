package server_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func createDidChangeConfiguration(setting string) protocol.DidChangeConfigurationParams {
	return protocol.DidChangeConfigurationParams{Settings: []string{setting}}
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
	handler := &ConfigurationHandler{t: t, experimental: configuration.Experimental{VulnerabilityScanning: false}}
	conn := jsonrpc2.NewConn(context.Background(), clientStream, handler)
	initialize(t, conn, protocol.InitializeParams{})

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	// when a file is opened, verify that configuration is fetched
	didOpen := createDidOpenTextDocumentParams(homedir, "Dockerfile", "FROM scratch", protocol.DockerfileLanguage)
	handler.handled = false
	err = conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
	require.NoError(t, err)
	for !handler.handled || configuration.Get(didOpen.TextDocument.URI).Experimental.VulnerabilityScanning {
		time.Sleep(100 * time.Millisecond)
	}

	testCases := []struct {
		name           string
		changedSetting string
		experimental   configuration.Experimental
	}{
		{
			name:           "vulnerabilityScanning=true",
			changedSetting: "docker.lsp.experimental.vulnerabilityScanning",
			experimental: configuration.Experimental{
				VulnerabilityScanning: true,
				Scout: configuration.Scout{
					CriticalHighVulnerabilities: false,
					NotPinnedDigest:             false,
					RecommendedTag:              false,
					Vulnerabilites:              false,
				},
			},
		},
		{
			name:           "vulnerabilityScanning=true,criticalHighVulnerabilities=true",
			changedSetting: "docker.lsp.experimental.scout.criticalHighVulnerabilities",
			experimental: configuration.Experimental{
				VulnerabilityScanning: true,
				Scout: configuration.Scout{
					CriticalHighVulnerabilities: true,
					NotPinnedDigest:             false,
					RecommendedTag:              false,
					Vulnerabilites:              false,
				},
			},
		},
		{
			name:           "vulnerabilityScanning=true,notPinnedDigest=true",
			changedSetting: "docker.lsp.experimental.scout.notPinnedDigest",
			experimental: configuration.Experimental{
				VulnerabilityScanning: true,
				Scout: configuration.Scout{
					CriticalHighVulnerabilities: false,
					NotPinnedDigest:             true,
					RecommendedTag:              false,
					Vulnerabilites:              false,
				},
			},
		},
		{
			name:           "vulnerabilityScanning=true,recommendedTag=true",
			changedSetting: "docker.lsp.experimental.scout.recommendedTag",
			experimental: configuration.Experimental{
				VulnerabilityScanning: true,
				Scout: configuration.Scout{
					CriticalHighVulnerabilities: false,
					NotPinnedDigest:             false,
					RecommendedTag:              true,
					Vulnerabilites:              false,
				},
			},
		},
		{
			name:           "vulnerabilityScanning=true,vulnerabilities=true",
			changedSetting: "docker.lsp.experimental.scout.vulnerabilities",
			experimental: configuration.Experimental{
				VulnerabilityScanning: true,
				Scout: configuration.Scout{
					CriticalHighVulnerabilities: false,
					NotPinnedDigest:             false,
					RecommendedTag:              false,
					Vulnerabilites:              true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuration.Store(
				didOpen.TextDocument.URI,
				configuration.Configuration{
					Experimental: configuration.Experimental{VulnerabilityScanning: false},
				},
			)
			handler.handled = false
			handler.experimental = tc.experimental

			didChangeConfiguration := createDidChangeConfiguration(tc.changedSetting)
			err = conn.Notify(context.Background(), protocol.MethodWorkspaceDidChangeConfiguration, didChangeConfiguration)
			require.NoError(t, err)
			for !handler.handled || !configuration.Get(didOpen.TextDocument.URI).Experimental.VulnerabilityScanning {
				time.Sleep(100 * time.Millisecond)
			}
			require.Equal(t, tc.experimental, configuration.Get(didOpen.TextDocument.URI).Experimental)
		})
	}

	handler.handled = false
	didChangeConfiguration := createDidChangeConfiguration(configuration.ConfigTelemetry)
	err = conn.Notify(context.Background(), protocol.MethodWorkspaceDidChangeConfiguration, didChangeConfiguration)
	require.NoError(t, err)
	for !handler.handled {
		time.Sleep(100 * time.Millisecond)
	}
}
