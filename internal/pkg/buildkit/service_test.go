package buildkit

import (
	"os"
	"runtime"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestParse(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("DOCKER_LANGUAGE_SERVER_WINDOWS_CI") == "true" {
		// not easy to setup Docker and Buildx in Windows CI, skipping for now
		t.Skip("skipping test on Windows CI")
		return
	}

	testCases := []struct {
		name        string
		content     string
		overlaps    bool
		diagnostics []protocol.Diagnostic
	}{
		{
			name:     "empty file",
			content:  "",
			overlaps: true,
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
			name:     "unrecognized instruction",
			content:  "FROM scratch\nUNKNOWN abc",
			overlaps: true,
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
			name:     "unrecognized flag, suggestion provided",
			content:  "FROM --platform2=linux/amd64 scratch",
			overlaps: true,
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
			name:     "unrecognized flag, no suggestion provided",
			content:  "FROM --abc=linux/amd64 scratch",
			overlaps: true,
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
			name:     "deprecated MAINTAINER is a warning",
			content:  "FROM scratch\nMAINTAINER test123@docker.com",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=MaintainerDeprecated",
							Edit:  "# check=skip=MaintainerDeprecated\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "MAINTAINER with multiple words",
			content:  "FROM scratch\nMAINTAINER hello world",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=MaintainerDeprecated",
							Edit:  "# check=skip=MaintainerDeprecated\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "MAINTAINER does not add additional quotes if already enclosed",
			content:  "FROM scratch\nMAINTAINER \"hello world\"",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=MaintainerDeprecated",
							Edit:  "# check=skip=MaintainerDeprecated\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "MAINTAINER code action only adds trailing quote",
			content:  "FROM scratch\nMAINTAINER \"hello world",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=MaintainerDeprecated",
							Edit:  "# check=skip=MaintainerDeprecated\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "MAINTAINER code action only adds leading quote",
			content:  "FROM scratch\nMAINTAINER hello world\"",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=MaintainerDeprecated",
							Edit:  "# check=skip=MaintainerDeprecated\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "stage name all uppercase",
			content:  "FROM scratch AS TEST",
			overlaps: false,
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
						{
							Title: "Ignore this type of error with check=skip=StageNameCasing",
							Edit:  "# check=skip=StageNameCasing\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "stage name mixed case",
			content:  "FROM scratch AS MixeD",
			overlaps: false,
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
						{
							Title: "Ignore this type of error with check=skip=StageNameCasing",
							Edit:  "# check=skip=StageNameCasing\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "redundant $TARGETPLATFORM suggests code action to remove the flag",
			content:  "FROM --platform=$TARGETPLATFORM alpine AS builder",
			overlaps: false,
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
						{
							Title: "Ignore this type of error with check=skip=RedundantTargetPlatform",
							Edit:  "# check=skip=RedundantTargetPlatform\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "inconsistent casing suggests uppercase code action",
			content:  "FROM scratch\ncopy  --link . .",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=ConsistentInstructionCasing",
							Edit:  "# check=skip=ConsistentInstructionCasing\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "inconsistent casing suggests lowercase code action",
			content:  "from scratch\nfrom scratch\nCOPY  --link . .",
			overlaps: true,
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
						{
							Title: "Ignore this type of error with check=skip=ConsistentInstructionCasing",
							Edit:  "# check=skip=ConsistentInstructionCasing\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:        "ignore failed to resolve source metadata errors from Docker Hub (https://hub.docker.com/_/docker123)",
			content:     "FROM docker123",
			overlaps:    true,
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:        "ignore failed to resolve source metadata errors from Amazon ECR",
			content:     "FROM aws_account_id.dkr.ecr.region.amazonaws.com/docker123:testtag",
			overlaps:    true,
			diagnostics: []protocol.Diagnostic{},
		},
		{
			name:     "DuplicateStageName",
			content:  "FROM scratch AS base\nFROM scratch AS base",
			overlaps: true,
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 20},
					},
					Message:  "Stage names should be unique (Duplicate stage name \"base\", stage names should be unique)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "DuplicateStageName"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/duplicate-stage-name/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Ignore this type of error with check=skip=DuplicateStageName",
							Edit:  "# check=skip=DuplicateStageName\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "MultipleInstructionsDisallowed",
			content:  "FROM scratch\nCMD [ \"ls\" ]\nCMD [ \"ls\" ]",
			overlaps: true,
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 12},
					},
					Message:  "Multiple instructions of the same type should not be used in the same stage (Multiple CMD instructions should not be used in the same stage because only the last one will be used)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "MultipleInstructionsDisallowed"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/multiple-instructions-disallowed/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Ignore this type of error with check=skip=MultipleInstructionsDisallowed",
							Edit:  "# check=skip=MultipleInstructionsDisallowed\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "NoEmptyContinuation",
			content:  "FROM scratch\nRUN ls \\\n\nls",
			overlaps: true,
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 2},
					},
					Message:  "Empty continuation lines will become errors in a future release (Empty continuation line)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "NoEmptyContinuation"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/no-empty-continuation/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Ignore this type of error with check=skip=NoEmptyContinuation",
							Edit:  "# check=skip=NoEmptyContinuation\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
						},
					},
				},
			},
		},
		{
			name:     "WorkdirRelativePath",
			content:  "FROM scratch\nWORKDIR dir",
			overlaps: true,
			diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 11},
					},
					Message:  "Relative workdir without an absolute workdir declared within the build can have unexpected results if the base image changes (Relative workdir \"dir\" can have unexpected results if the base image changes)",
					Source:   types.CreateStringPointer("buildkit-testing-source"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning),
					Code:     &protocol.IntegerOrString{Value: "WorkdirRelativePath"},
					CodeDescription: &protocol.CodeDescription{
						HRef: "https://docs.docker.com/go/dockerfile/rule/workdir-relative-path/",
					},
					Data: []types.NamedEdit{
						{
							Title: "Ignore this type of error with check=skip=WorkdirRelativePath",
							Edit:  "# check=skip=WorkdirRelativePath\n",
							Range: &protocol.Range{
								Start: protocol.Position{Line: 0, Character: 0},
								End:   protocol.Position{Line: 0, Character: 0},
							},
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
			RemoveOverlappingIssues = false
			doc := document.NewDocument(document.NewDocumentManager(), uri.URI("uri:///Dockerfile"), protocol.DockerfileLanguage, 1, []byte(tc.content))
			diagnostics := collector.CollectDiagnostics("buildkit-testing-source", contextPath, doc, tc.content)
			require.Equal(t, tc.diagnostics, diagnostics)

			RemoveOverlappingIssues = true
			if tc.overlaps {
				diagnostics := collector.CollectDiagnostics("buildkit-testing-source", contextPath, doc, tc.content)
				require.Len(t, diagnostics, 0)
			} else {
				diagnostics := collector.CollectDiagnostics("buildkit-testing-source", contextPath, doc, tc.content)
				require.Equal(t, tc.diagnostics, diagnostics)
			}
		})
	}
}
