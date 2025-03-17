package hcl

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/scout"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestCollectDiagnostics(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		diagnostics []protocol.Diagnostic
	}{
		{
			name:    "target block with no name",
			content: "target {\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "Missing name for target (All target blocks must have 1 labels (name).)",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 8},
					},
				},
			},
		},
		{
			name:    "target block with missing attribute value",
			content: "target {\n  dockerfile = \n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "Invalid expression (Expected the start of an expression, but found an invalid expression token.)",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 15},
					},
				},
			},
		},
		{
			name:    "target block with network attribute empty string",
			content: "target \"t1\" {\n  tags = [ \"alpine:3.17.0\" ]\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "The image contains 1 critical and 7 high vulnerabilities",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "critical_high_vulnerabilities"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://hub.docker.com/layers/library/alpine/3.17.0/images/sha256-c0d488a800e4127c334ad20d61d7bc21b4097540327217dfab52262adc02380c",
					},
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 12},
						End:   protocol.Position{Line: 1, Character: 25},
					},
				},
			},
		},
		{
			name:    "target block with network attribute empty string",
			content: "target \"t1\" {\n  network = \"\"\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "network attribute must be one of: default, host, or none",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 12},
						End:   protocol.Position{Line: 1, Character: 14},
					},
				},
			},
		},
		{
			name:    "target block with entitlements attribute empty string",
			content: "target \"t1\" {\n  entitlements = [ \"\" ]\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "entitlements attribute must be one of: network.host or security.insecure",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 19},
						End:   protocol.Position{Line: 1, Character: 21},
					},
				},
			},
		},
		{
			name:        "args can be found in Dockerfile (unquoted)",
			content:     "target \"t1\" {\n  args = {\n    valid = \"value\"\n  }\n}",
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:        "args can be found in Dockerfile (quoted)",
			content:     "target \"t1\" {\n  args = {\n    \"valid\" = \"value\"\n  }\n}",
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:    "args cannot be found in Dockerfile",
			content: "target \"t1\" {\n  args = {\n    missing = \"value\"\n  }\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "'missing' not defined as an ARG in the Dockerfile",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 4},
						End:   protocol.Position{Line: 2, Character: 11},
					},
				},
			},
		},
		{
			name:        "args references built-in args",
			content:     "target \"t1\" {\n  args = {\n    HTTP_PROXY = \"\"\n    HTTPS_PROXY = \"\"\n    FTP_PROXY = \"\"\n    NO_PROXY = \"\"\n    ALL_PROXY = \"\"\n  }\n}",
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:    "target cannot be found in Dockerfile",
			content: "target \"t1\" {\n  target = \"nonexistent\"\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "target could not be found in the Dockerfile",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 11},
						End:   protocol.Position{Line: 1, Character: 24},
					},
				},
			},
		},
		{
			name:        "ignore args if dockerfile-inline is used",
			content:     "target \"t1\" {\n  dockerfile-inline = \"\"\n  args = {\n    missing = \"value\"\n  }\n}",
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:    "dockerfile attribute flagged as unnecessary if dockerfile-inline exists",
			content: "target \"t1\" {\n  dockerfile = \"./Dockerfile\"\n  dockerfile-inline = \"FROM scratch\"\n}",
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "dockerfile attribute is ignored if dockerfile-inline is defined",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Tags:     []protocol.DiagnosticTag{protocol.DiagnosticTagUnnecessary},
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 2},
						End:   protocol.Position{Line: 1, Character: 29},
					},
					Data: []types.NamedEdit{
						{
							Title: "Remove unnecessary dockerfile attribute",
							Edit:  "",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 1},
								End:   protocol.Position{Line: 2},
							},
						},
					},
				},
			},
		},
		{
			name: "target inheritance checked when looking at the target attribute",
			content: `target "lint" {
  target = "build"
  dockerfile = "./Dockerfile2"
}

target "lint2" {
  inherits = ["lint"]
  target = "build"
}`,
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:        "target inheritance references non-existing parent target",
			content:     "target \"child\" {\n  inherits = [\"parent\"]\n  target = \"build\"\n}",
			diagnostics: []protocol.Diagnostic{},
		},
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := path.Join(wd, "..", "..", "..")
	diagnosticsTestFolderPath := path.Join(projectRoot, "testdata", "diagnostics")
	bakeFilePath := path.Join(diagnosticsTestFolderPath, "docker-bake.hcl")
	bakeFileURI := uri.URI(fmt.Sprintf("file://%v", bakeFilePath))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			bytes := []byte(tc.content)
			err := os.WriteFile(bakeFilePath, bytes, 0644)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := os.Remove(bakeFilePath)
				require.NoError(t, err)
			})

			collector := &BakeHCLDiagnosticsCollector{docs: manager, scout: scout.NewService()}
			doc := document.NewBakeHCLDocument(bakeFileURI, 1, bytes)
			diagnostics := collector.CollectDiagnostics("docker-language-server", "", doc, "")
			require.Equal(t, tc.diagnostics, diagnostics)
		})
	}
}
