package scout

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestCalculateDiagnostics(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		diagnostics []protocol.Diagnostic
	}{
		{
			name:    "outdated alpine:3.16.1",
			content: "FROM alpine:3.16.1",
			diagnostics: []protocol.Diagnostic{
				{
					Message: "The image can be pinned to a digest",
					Source:  types.CreateStringPointer("scout-testing-source"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityHint),
				},
				{
					Message: "The image contains 1 critical and 3 high vulnerabilities",
					Source:  types.CreateStringPointer("scout-testing-source"),
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://hub.docker.com/layers/library/alpine/3.16.1/images/sha256-9b2a28eb47540823042a2ba401386845089bb7b62a9637d55816132c4c3c36eb",
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
				},
				{
					Message: "Tag recommendations available",
					Source:  types.CreateStringPointer("scout-testing-source"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 18},
					},
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityInformation),
				},
			},
		},
	}

	c := NewService()
	for _, tc := range testCases {
		uri := uri.URI("uri:///Dockerfile")
		doc := document.NewDocument(uri, protocol.DockerfileLanguage, 1, []byte(tc.content))
		testConfigs := []struct {
			description   string
			shouldScan    bool
			setupFunction func()
		}{
			{
				description: "removed",
				shouldScan:  true,
				setupFunction: func() {
					configuration.Remove(protocol.DocumentUri(uri))
				},
			},
			{
				description: "scan=true",
				shouldScan:  true,
				setupFunction: func() {
					configuration.Store(
						protocol.DocumentUri(uri),
						configuration.Configuration{Experimental: configuration.Experimental{VulnerabilityScanning: true}},
					)
				},
			},
			{
				description: "scan=false",
				shouldScan:  false,
				setupFunction: func() {
					configuration.Store(
						protocol.DocumentUri(uri),
						configuration.Configuration{Experimental: configuration.Experimental{VulnerabilityScanning: false}},
					)
				},
			},
		}

		for _, testConfig := range testConfigs {
			t.Run(fmt.Sprintf("%v (%v)", tc.name, testConfig.description), func(t *testing.T) {
				defer configuration.Remove(protocol.DocumentUri(uri))
				testConfig.setupFunction()
				diagnostics := c.CollectDiagnostics("scout-testing-source", "", doc, tc.content)
				if os.Getenv("DOCKER_NETWORK_NONE") == "true" || !testConfig.shouldScan {
					require.Len(t, diagnostics, 0)
					return
				}

				for _, diagnostic := range diagnostics {
					found := false
					for _, expectedDiagnostic := range tc.diagnostics {
						if diagnostic.Message == expectedDiagnostic.Message {
							require.Equal(t, expectedDiagnostic.Range, diagnostic.Range)
							require.Equal(t, expectedDiagnostic.Severity, diagnostic.Severity)
							require.Equal(t, expectedDiagnostic.Source, diagnostic.Source)
							require.Equal(t, expectedDiagnostic.CodeDescription, diagnostic.CodeDescription)
							found = true
							break
						}
					}

					if !found {
						t.Errorf("expected diagnostic could not be found: %v", diagnostic.Message)
					}
				}
				require.Equal(t, len(tc.diagnostics), len(diagnostics))
			})
		}
	}
}

func TestGetHovers(t *testing.T) {
	testCases := []struct {
		x     string
		name  string
		image string
		value string
	}{
		{
			name:  "hovers combined",
			image: "alpine:3.16.1",
			value: "Current image vulnerabilities:   1C   3H   9M   0L \r\n\r\nRecommended tags:\n\n<table>\n<tr><td><code>3.21.3</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n<tr><td><code>3.20.6</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n<tr><td><code>3.18.12</code></td><td align=\"right\">  0C</td><td align=\"right\">  0H</td><td align=\"right\">  0M</td><td align=\"right\">  0L</td><td align=\"right\"></td></tr>\n</table>\n",
		},
	}

	s := NewService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hover, err := s.Hover(context.Background(), tc.image)
			if os.Getenv("DOCKER_NETWORK_NONE") == "true" {
				var dns *net.DNSError
				require.True(t, errors.As(err, &dns))
				return
			}

			require.Nil(t, err)
			markupContent, ok := hover.Contents.(protocol.MarkupContent)
			require.True(t, ok)
			require.Equal(t, protocol.MarkupKindMarkdown, markupContent.Kind)
			require.Equal(t, tc.value, markupContent.Value)
		})
	}
}
