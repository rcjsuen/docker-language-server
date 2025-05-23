package server

import (
	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

func (s *Server) WorkspaceDidChangeConfiguration(ctx *glsp.Context, params *protocol.DidChangeConfigurationParams) error {
	changedSettings, _ := params.Settings.([]any)
	scoutConfigurationChanged := false
	for _, setting := range changedSettings {
		config := setting.(string)
		switch config {
		case configuration.ConfigTelemetry:
			go s.FetchUnscopedConfiguration()
		case configuration.ConfigExperimentalVulnerabilityScanning:
			fallthrough
		case configuration.ConfigExperimentalScoutCriticalHighVulnerabilities:
			fallthrough
		case configuration.ConfigExperimentalScoutNotPinnedDigest:
			fallthrough
		case configuration.ConfigExperimentalScoutRecommendedTag:
			fallthrough
		case configuration.ConfigExperimentalScoutVulnerabilities:
			scoutConfigurationChanged = true
		}
	}

	if scoutConfigurationChanged {
		scopes := configuration.Documents()
		if len(scopes) > 0 {
			go func() {
				defer s.handlePanic("WorkspaceDidChangeConfiguration")

				s.FetchConfigurations(scopes)
				s.recomputeDiagnostics()
			}()
		}
	}
	return nil
}
