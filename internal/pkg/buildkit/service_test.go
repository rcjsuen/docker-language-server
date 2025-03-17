package buildkit

import (
	"os"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		diagnostics []protocol.Diagnostic
	}{
		{
			name:    "empty file",
			content: "",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
					Message:  "the Dockerfile cannot be empty",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
				},
			},
		},
		{
			name:    "unrecognized instruction",
			content: "FROM scratch\nUNKNOWN abc",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 11},
					},
					Message:  "dockerfile parse error on line 2: unknown instruction: UNKNOWN",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
				},
			},
		},
		{
			name:    "unrecognized flag, suggestion provided",
			content: "FROM --platform2=linux/amd64 scratch",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 36},
					},
					Message:  "dockerfile parse error on line 1: unknown flag: --platform2 (did you mean platform?)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Data: []types.NamedEdit{
						{
							Title: "Change flag name to platform",
							Edit:  "FROM --platform=linux/amd64 scratch",
						},
					},
				},
			},
		},
		{
			name:    "unrecognized flag, no suggestion provided",
			content: "FROM --abc=linux/amd64 scratch",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 30},
					},
					Message:  "dockerfile parse error on line 1: unknown flag: --abc",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Data: []types.NamedEdit{
						{
							Title: "Remove unrecognized flag",
							Edit:  "FROM scratch",
						},
					},
				},
			},
		},
		{
			name:    "deprecated MAINTAINER is a warning",
			content: "FROM scratch\nMAINTAINER test123@docker.com",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 29},
					},
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []types.NamedEdit{
						{
							Title: "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
							Edit:  "LABEL org.opencontainers.image.authors=\"test123@docker.com\"",
						},
					},
				},
			},
		},
		{
			name:    "MAINTAINER with multiple words",
			content: "FROM scratch\nMAINTAINER hello world",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 22},
					},
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []types.NamedEdit{
						{
							Title: "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
							Edit:  "LABEL org.opencontainers.image.authors=\"hello world\"",
						},
					},
				},
			},
		},
		{
			name:    "MAINTAINER does not add additional quotes if already enclosed",
			content: "FROM scratch\nMAINTAINER \"hello world\"",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 24},
					},
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []types.NamedEdit{
						{
							Title: "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
							Edit:  "LABEL org.opencontainers.image.authors=\"hello world\"",
						},
					},
				},
			},
		},
		{
			name:    "MAINTAINER code action only adds trailing quote",
			content: "FROM scratch\nMAINTAINER \"hello world",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 23},
					},
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []types.NamedEdit{
						{
							Title: "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
							Edit:  "LABEL org.opencontainers.image.authors=\"hello world\"",
						},
					},
				},
			},
		},
		{
			name:    "MAINTAINER code action only adds leading quote",
			content: "FROM scratch\nMAINTAINER hello world\"",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 23},
					},
					Message:  "The MAINTAINER instruction is deprecated, use a label instead to define an image author (Maintainer instruction is deprecated in favor of using label)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MaintainerDeprecated"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/maintainer-deprecated/",
					},
					Tags: []protocol.DiagnosticTag{protocol.DiagnosticTagDeprecated},
					Data: []types.NamedEdit{
						{
							Title: "Convert MAINTAINER to a org.opencontainers.image.authors LABEL",
							Edit:  "LABEL org.opencontainers.image.authors=\"hello world\"",
						},
					},
				},
			},
		},
		{
			name:    "stage name all uppercase",
			content: "FROM scratch AS TEST",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 20},
					},
					Message:  "Stage names should be lowercase (Stage name 'TEST' should be lowercase)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "StageNameCasing"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/stage-name-casing/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Convert stage name (TEST) to lowercase (test)",
							Edit:  "FROM scratch AS test",
						},
					},
				},
			},
		},
		{
			name:    "stage name mixed case",
			content: "FROM scratch AS MixeD",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					Message:  "Stage names should be lowercase (Stage name 'MixeD' should be lowercase)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "StageNameCasing"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/stage-name-casing/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Convert stage name (MixeD) to lowercase (mixed)",
							Edit:  "FROM scratch AS mixed",
						},
					},
				},
			},
		},
		{
			name:    "redundant $TARGETPLATFORM suggests code action to remove the flag",
			content: "FROM --platform=$TARGETPLATFORM alpine AS builder",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 49},
					},
					Message:  "Setting platform to predefined $TARGETPLATFORM in FROM is redundant as this is the default behavior (Setting platform to predefined $TARGETPLATFORM in FROM is redundant as this is the default behavior)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "RedundantTargetPlatform"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/redundant-target-platform/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Remove unnecessary --platform flag",
							Edit:  "FROM alpine AS builder",
						},
					},
				},
			},
		},
		{
			name:    "inconsistent casing suggests uppercase code action",
			content: "FROM scratch\ncopy  --link . .",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 16},
					},
					Message:  "All commands within the Dockerfile should use the same casing (either upper or lower) (Command 'copy' should match the case of the command majority (uppercase))",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "ConsistentInstructionCasing"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/consistent-instruction-casing/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Convert to uppercase",
							Edit:  "COPY --link . .",
						},
					},
				},
			},
		},
		{
			name:    "inconsistent casing suggests lowercase code action",
			content: "from scratch\nfrom scratch\nCOPY  --link . .",
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 16},
					},
					Message:  "All commands within the Dockerfile should use the same casing (either upper or lower) (Command 'COPY' should match the case of the command majority (lowercase))",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "ConsistentInstructionCasing"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/consistent-instruction-casing/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Convert to lowercase",
							Edit:  "copy --link . .",
						},
					},
				},
			},
		},
	}

	contextPath := os.TempDir()
	collector := NewBuildKitDiagnosticsCollector()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewDocument(uri.URI("uri:///Dockerfile"), protocol.DockerfileLanguage, 1, []byte(tc.content))
			diagnostics := collector.CollectDiagnostics("buildkit-testing-source", contextPath, doc, tc.content)
			require.Equal(t, tc.diagnostics, diagnostics)
		})
	}
}
