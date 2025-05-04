package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/pkg/buildkit"
	"github.com/docker/docker-language-server/internal/pkg/cli/metadata"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

type PublishDiagnosticsHandler struct {
	t               *testing.T
	responseChannel chan error
	diagnostics     protocol.PublishDiagnosticsParams
}

func (h *PublishDiagnosticsHandler) Handle(_ context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) {
	switch request.Method {
	case protocol.ServerTextDocumentPublishDiagnostics:
		if request.Notif && request.Params != nil {
			// always deserialize to a completely new struct
			h.diagnostics = protocol.PublishDiagnosticsParams{}
			require.NoError(h.t, json.Unmarshal(*request.Params, &h.diagnostics))
			h.responseChannel <- nil
		}
	case protocol.ServerWorkspaceConfiguration:
		if !request.Notif && request.Params != nil {
			HandleConfiguration(h.t, conn, request, true)
		}
	}
}

func TestPublishDiagnostics(t *testing.T) {
	// ensure the language server works without any workspace folders
	testPublishDiagnostics(t, protocol.InitializeParams{})

	// ensure the language server works without any workspace folders
	testPublishDiagnostics(t, protocol.InitializeParams{
		InitializationOptions: map[string]any{
			"dockerfileExperimental": map[string]bool{"removeOverlappingIssues": true},
		},
	})

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	testPublishDiagnostics(t, protocol.InitializeParams{
		WorkspaceFolders: []protocol.WorkspaceFolder{{Name: "home", URI: homedir}},
	})
}

func initialize(t *testing.T, conn *jsonrpc2.Conn, initializeParams protocol.InitializeParams) {
	if options, ok := initializeParams.InitializationOptions.(map[string]any); ok {
		options["telemetry"] = "off"
	} else {
		initializeParams.InitializationOptions = map[string]string{"telemetry": "off"}
	}
	var initializeResult *protocol.InitializeResult
	err := conn.Call(context.Background(), protocol.MethodInitialize, initializeParams, &initializeResult)
	require.NoError(t, err)

	syncKind := protocol.TextDocumentSyncKindFull
	expected := protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CodeActionProvider:         protocol.CodeActionOptions{},
			CompletionProvider:         &protocol.CompletionOptions{},
			DefinitionProvider:         protocol.DefinitionOptions{},
			DocumentFormattingProvider: protocol.DocumentFormattingOptions{},
			DocumentHighlightProvider:  &protocol.DocumentHighlightOptions{},
			DocumentLinkProvider:       &protocol.DocumentLinkOptions{},
			DocumentSymbolProvider:     protocol.DocumentSymbolOptions{},
			ExecuteCommandProvider: &protocol.ExecuteCommandOptions{
				Commands: []string{types.TelemetryCallbackCommandId},
			},
			HoverProvider:            protocol.HoverOptions{},
			InlayHintProvider:        protocol.InlayHintOptions{},
			InlineCompletionProvider: protocol.InlineCompletionOptions{},
			RenameProvider: protocol.RenameOptions{
				PrepareProvider: types.CreateBoolPointer(false),
			},
			SemanticTokensProvider: protocol.SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenModifiers: []string{},
					TokenTypes:     hcl.SemanticTokenTypes,
				},
				Full:  true,
				Range: false,
			},
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: &protocol.True,
				Change:    &syncKind,
			},
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    "docker-language-server",
			Version: &metadata.Version,
		},
	}
	requireJsonEqual(t, expected, initializeResult)
}

func testPublishDiagnostics(t *testing.T, initializeParams protocol.InitializeParams) {
	s := startServer()

	client := bytes.NewBuffer(make([]byte, 0, 1024))
	server := bytes.NewBuffer(make([]byte, 0, 1024))
	serverStream := &TestStream{incoming: server, outgoing: client, closed: false}
	defer serverStream.Close()
	go s.ServeStream(serverStream)

	handler := &PublishDiagnosticsHandler{t: t, responseChannel: make(chan error)}
	clientStream := jsonrpc2.NewBufferedStream(&TestStream{incoming: client, outgoing: server, closed: false}, jsonrpc2.VSCodeObjectCodec{})
	defer clientStream.Close()
	conn := jsonrpc2.NewConn(
		context.Background(),
		clientStream,
		handler,
	)
	defer func() {
		buildkit.RemoveOverlappingIssues = false
	}()
	initialize(t, conn, initializeParams)

	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	testCases := []struct {
		name             string
		content          string
		included         []bool
		overlappingIssue bool
		diagnostics      []protocol.Diagnostic
	}{
		{
			name:             "no diagnostics",
			content:          "FROM scratch",
			included:         []bool{},
			overlappingIssue: false,
			diagnostics:      []protocol.Diagnostic{},
		},
		{
			name:             "MAINTAINER is deprecated",
			content:          "FROM scratch\nMAINTAINER x",
			included:         []bool{true},
			overlappingIssue: true,
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("docker-language-server"),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 12},
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []any{
						map[string]any{
							"edit":  "LABEL org.opencontainers.image.authors=\"x\"",
							"title": "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
						},
					},
				},
			},
		},
		{
			name:             "JSON args",
			content:          "FROM alpine:3.16.1\nCMD ls",
			included:         []bool{true, false, false, false},
			overlappingIssue: false,
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "JSON arguments recommended for ENTRYPOINT/CMD to prevent unintended behavior related to OS signals (JSON arguments recommended for CMD to prevent unintended behavior related to OS signals)",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "JSONArgsRecommended"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/json-args-recommended/",
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 6},
					},
				},
				{
					Message:  "The image can be pinned to a digest",
					Source:   types.CreateStringPointer("docker-language-server"),
					Code:     &protocol.IntegerOrString{Value: "not_pinned_digest"},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityHint),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
					Data: []any{
						map[string]any{
							"edit":  "FROM alpine:3.16.1@sha256:7580ece7963bfa863801466c0a488f11c86f85d9988051a9f9c68cb27f6b7872",
							"title": "Pin the base image digest",
						},
					},
				},
				{
					Message:  "The image contains 1 critical and 3 high vulnerabilities",
					Source:   types.CreateStringPointer("docker-language-server"),
					Code:     &protocol.IntegerOrString{Value: "critical_high_vulnerabilities"},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://hub.docker.com/layers/library/alpine/3.16.1/images/sha256-9b2a28eb47540823042a2ba401386845089bb7b62a9637d55816132c4c3c36eb",
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
				},
				{
					Message:  "Tag recommendations available",
					Source:   types.CreateStringPointer("docker-language-server"),
					Code:     &protocol.IntegerOrString{Value: "recommended_tag"},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityInformation),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
					Data: []any{
						map[string]any{
							"edit":  "FROM alpine:3.21.3",
							"title": "Update image to preferred tag (3.21.3)",
						},
						map[string]any{
							"edit":  "FROM alpine:3.20.6",
							"title": "Update image OS minor version (3.20.6)",
						},
						map[string]any{
							"edit":  "FROM alpine:3.18.12",
							"title": "Update image OS minor version (3.18.12)",
						},
					},
				},
			},
		},
	}

	removeOverlappingIssues := false
	if options, ok := initializeParams.InitializationOptions.(map[string]any); ok {
		if settings, ok := options["dockerfileExperimental"].(map[string]bool); ok {
			if value, ok := settings["removeOverlappingIssues"]; ok {
				removeOverlappingIssues = value
			}
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v (len(workspaceFolders) == %v, removeOverlappingIssues=%v)", tc.name, len(initializeParams.WorkspaceFolders), removeOverlappingIssues), func(t *testing.T) {
			didOpen := createDidOpenTextDocumentParams(homedir, t.Name(), tc.content, "dockerfile")
			err := conn.Notify(context.Background(), protocol.MethodTextDocumentDidOpen, didOpen)
			require.NoError(t, err)

			<-handler.responseChannel
			params := handler.diagnostics

			filteredDiagnostics := []protocol.Diagnostic{}
			if os.Getenv("DOCKER_NETWORK_NONE") == "true" {
				for i := range tc.included {
					if tc.included[i] {
						filteredDiagnostics = append(filteredDiagnostics, tc.diagnostics[i])
					}
				}
			} else {
				filteredDiagnostics = tc.diagnostics
			}

			if removeOverlappingIssues && tc.overlappingIssue {
				filteredDiagnostics = []protocol.Diagnostic{}
			}

			require.Equal(t, didOpen.TextDocument.URI, params.URI)
			require.Equal(t, filteredDiagnostics, params.Diagnostics)
			require.Equal(t, int32(1), *params.Version)
		})
	}
}
