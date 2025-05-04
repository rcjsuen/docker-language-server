package server

import (
	"encoding/json"
	"net/url"
	"slices"

	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/pkg/buildkit"
	"github.com/docker/docker-language-server/internal/pkg/cli/metadata"
	"github.com/docker/docker-language-server/internal/telemetry"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/go-git/go-git/v5"
)

func (s *Server) Initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	s.client = &LanguageClient{call: ctx.Call, notify: ctx.Notify}

	workspaceFolders := []string{}
	for _, workspaceFolder := range params.WorkspaceFolders {
		parsed, _ := url.Parse(workspaceFolder.URI)
		if parsed != nil {
			workspaceFolders = append(workspaceFolders, parsed.Path)
		}
	}

	if clientConfig, ok := params.InitializationOptions.(map[string]any); ok {
		if settings, ok := clientConfig["dockerfileExperimental"].(map[string]any); ok {
			if value, ok := settings["removeOverlappingIssues"].(bool); ok {
				buildkit.RemoveOverlappingIssues = value
			}
		}

		if settings, ok := clientConfig["dockercomposeExperimental"].(map[string]any); ok {
			if composeCompletion, ok := settings["composeCompletion"].(bool); ok {
				s.composeCompletion = composeCompletion
			}
		}

		if value, ok := clientConfig["telemetry"].(string); ok {
			s.updateTelemetrySetting(value)
		}
	}

	if len(workspaceFolders) > 0 {
		s.workspaceFolders = workspaceFolders
	} else if params.RootURI != nil {
		parsed, _ := url.Parse(*params.RootURI)
		s.workspaceFolders = []string{parsed.Path}
	} else if params.RootPath != nil {
		s.workspaceFolders = []string{*params.RootPath}
	}

	if len(s.workspaceFolders) > 0 {
		for i := range s.workspaceFolders {
			r, err := git.PlainOpen(s.workspaceFolders[i])
			if err == nil {
				remote, err := r.Remote("origin")
				if err == nil {
					config := remote.Config()
					if config != nil && len(config.URLs) > 0 {
						s.gitRemotes[s.workspaceFolders[i]] = types.GitRepository(config.URLs[0])
					}
				}
			}
		}
	}

	var codeLensProvider *protocol.CodeLensOptions
	bytes, err := json.Marshal(params.Capabilities.Experimental)
	if err == nil {
		_ = json.Unmarshal(bytes, &s.capabilities)
		if s.capabilities != nil && slices.Contains(s.capabilities.Capabilities.Commands, types.BakeBuildCommandId) {
			codeLensProvider = &protocol.CodeLensOptions{}
		}
	}

	s.toggleSupportedFeatures(params)
	if params.ClientInfo != nil {
		s.sessionTelemetryProperties["lsp_client_info_name"] = params.ClientInfo.Name
		if params.ClientInfo.Version != nil {
			s.sessionTelemetryProperties["lsp_client_info_version"] = *params.ClientInfo.Version
		}
		if s.capabilities != nil {
			if s.capabilities.Capabilities.ClientInfoExtras.ClientName != "" {
				s.sessionTelemetryProperties["client_name"] = s.capabilities.Capabilities.ClientInfoExtras.ClientName
			}
			if s.capabilities.Capabilities.ClientInfoExtras.ClientSession != "" {
				s.sessionTelemetryProperties["client_session"] = s.capabilities.Capabilities.ClientInfoExtras.ClientSession
			}
		}
	}
	s.Enqueue(telemetry.EventServerHeartbeat, map[string]any{
		"type": telemetry.ServerHeartbeatTypeInitialized,
	})

	syncKind := protocol.TextDocumentSyncKindFull
	result := protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CodeActionProvider:        protocol.CodeActionOptions{},
			CodeLensProvider:          codeLensProvider,
			CompletionProvider:        &protocol.CompletionOptions{},
			DefinitionProvider:        protocol.DefinitionOptions{},
			DocumentHighlightProvider: protocol.DocumentHighlightOptions{},
			DocumentLinkProvider:      &protocol.DocumentLinkOptions{},
			DocumentSymbolProvider:    protocol.DocumentSymbolOptions{},
			ExecuteCommandProvider: &protocol.ExecuteCommandOptions{
				Commands: []string{types.TelemetryCallbackCommandId},
			},
			HoverProvider:            protocol.HoverOptions{},
			InlayHintProvider:        protocol.InlayHintOptions{},
			InlineCompletionProvider: protocol.InlineCompletionOptions{},
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
	if params.Capabilities.TextDocument != nil {
		if params.Capabilities.TextDocument.Formatting != nil &&
			params.Capabilities.TextDocument.Formatting.DynamicRegistration != nil &&
			*params.Capabilities.TextDocument.Formatting.DynamicRegistration {
			s.registerFormattingCapability()
		}
		if params.Capabilities.TextDocument.Rename != nil &&
			params.Capabilities.TextDocument.Rename.DynamicRegistration != nil &&
			*params.Capabilities.TextDocument.Rename.DynamicRegistration {
			s.registerRenameCapability()
		}
	} else {
		result.Capabilities.DocumentFormattingProvider = protocol.DocumentFormattingOptions{}
		result.Capabilities.RenameProvider = protocol.RenameOptions{
			PrepareProvider: types.CreateBoolPointer(true),
		}
	}
	return result, nil
}

func (s *Server) toggleSupportedFeatures(params *protocol.InitializeParams) {
	if params.Capabilities.TextDocument != nil {
		if params.Capabilities.TextDocument.Definition != nil {
			if params.Capabilities.TextDocument.Definition.LinkSupport != nil {
				s.definitionLinkSupport = *params.Capabilities.TextDocument.Definition.LinkSupport
			}
		}
	}

	if params.Capabilities.Window != nil {
		if params.Capabilities.Window.ShowDocument != nil && params.Capabilities.Window.ShowDocument.Support {
			s.showDocumentSupport = true
		}
		if params.Capabilities.Window.ShowMessage != nil {
			s.showMessageRequestSupport = true
		}
	}
}

type ClientInfoExtras struct {
	ClientName    string `json:"client_name"`
	ClientSession string `json:"client_session"`
}

type DockerLanguageServerCapabilities struct {
	Commands         []string         `json:"commands"`
	ClientInfoExtras ClientInfoExtras `json:"clientInfoExtras"`
}

type ExperimentalCapabilities struct {
	Capabilities DockerLanguageServerCapabilities `json:"dockerLanguageServerCapabilities"`
}
