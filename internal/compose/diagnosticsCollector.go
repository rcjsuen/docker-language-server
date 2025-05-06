package compose

import (
	"errors"
	"math"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/pkg/lsp/textdocument"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml"
)

type ComposeDiagnosticsCollector struct {
}

func NewComposeDiagnosticsCollector() textdocument.DiagnosticsCollector {
	return &ComposeDiagnosticsCollector{}
}

func (c *ComposeDiagnosticsCollector) SupportsLanguageIdentifier(languageIdentifier protocol.LanguageIdentifier) bool {
	return languageIdentifier == protocol.DockerComposeLanguage
}

func (c *ComposeDiagnosticsCollector) CollectDiagnostics(source, workspaceFolder string, doc document.Document, text string) []protocol.Diagnostic {
	err := doc.(document.ComposeDocument).ParsingError()
	if err != nil {
		var syntaxError *yaml.SyntaxError
		if errors.As(err, &syntaxError) {
			return []protocol.Diagnostic{
				{
					Message:  syntaxError.Message,
					Source:   types.CreateStringPointer(source),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      protocol.UInteger(syntaxError.Token.Position.Line) - 1,
							Character: protocol.UInteger(syntaxError.Token.Position.Column) - 1,
						},
						End: protocol.Position{
							Line:      protocol.UInteger(syntaxError.Token.Position.Line) - 1,
							Character: math.MaxUint32,
						},
					},
				},
			}
		}
	}
	return nil
}
