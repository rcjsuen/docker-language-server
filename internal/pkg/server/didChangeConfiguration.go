package server

import (
	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

func (s *Server) WorkspaceDidChangeConfiguration(ctx *glsp.Context, params *protocol.DidChangeConfigurationParams) error {
	changedSettings, _ := params.Settings.([]any)
	for _, setting := range changedSettings {
		config := setting.(string)
		if config == configuration.ConfigTelemetry {
			go s.FetchUnscopedConfiguration()
		}

		if config == configuration.ConfigExperimentalVulnerabilityScanning {
			scopes := configuration.Documents()
			if len(scopes) > 0 {
				go func() {
					defer s.handlePanic("WorkspaceDidChangeConfiguration")

					s.FetchConfigurations(scopes)
					s.recomputeDiagnostics()
				}()
			}
		}
	}
	return nil
}
