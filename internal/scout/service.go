package scout

import (
	"context"
	"strings"

	"github.com/docker/docker-language-server/internal/cache"
	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/pkg/lsp/textdocument"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

type Service interface {
	textdocument.DiagnosticsCollector
	Analyze(image string) ([]Diagnostic, error)
	Hover(ctx context.Context, image string) (*protocol.Hover, error)
}

type ServiceImpl struct {
	manager cache.CacheManager[ImageResponse]
}

func NewService() Service {
	client := NewLanguageGatewayClient()
	return &ServiceImpl{
		manager: cache.NewManager(client),
	}
}

func (s *ServiceImpl) Hover(ctx context.Context, image string) (*protocol.Hover, error) {
	resp, err := s.manager.Get(&ScoutImageKey{Image: image})
	if err == nil {
		hovers := []string{}
		for _, info := range resp.Infos {
			hovers = append(hovers, info.Description.Markdown)
		}

		if len(hovers) > 0 {
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Value: strings.Join(hovers, "\r\n\r\n"),
					Kind:  protocol.MarkupKindMarkdown,
				},
			}, nil
		}
	}
	return nil, err
}

func (s *ServiceImpl) Analyze(image string) ([]Diagnostic, error) {
	resp, err := s.manager.Get(&ScoutImageKey{Image: image})
	if err != nil {
		return nil, err
	}
	return resp.Diagnostics, nil
}

func (s *ServiceImpl) CalculateDiagnostics(ctx context.Context, source string, doc document.Document) ([]protocol.Diagnostic, error) {
	config := configuration.Get(protocol.DocumentUri(doc.URI()))
	if !config.Experimental.VulnerabilityScanning {
		return nil, nil
	}

	lspDiagnostics := []protocol.Diagnostic{}
	lines := strings.Split(string(doc.Input()), "\n")
	for _, child := range doc.(document.DockerfileDocument).Nodes() {
		if strings.EqualFold(child.Value, "FROM") && child.Next != nil {
			resp, err := s.manager.Get(&ScoutImageKey{Image: child.Next.Value})
			if err == nil {
				next := child.Next
				words := []string{child.Value}
				for next != nil {
					words = append(words, next.Value)
					next = next.Next
				}

				for _, diagnostic := range resp.Diagnostics {
					lspDiagnostic := ConvertDiagnostic(diagnostic, words, source, protocol.Range{
						Start: protocol.Position{
							Line:      uint32(child.StartLine - 1),
							Character: 0,
						},
						End: protocol.Position{
							Line:      uint32(child.EndLine - 1),
							Character: uint32(len(lines[child.StartLine-1])),
						},
					}, resp.Edits)
					lspDiagnostics = append(lspDiagnostics, lspDiagnostic)
				}
				continue
			}
		}
	}
	return lspDiagnostics, nil
}

func (s *ServiceImpl) CollectDiagnostics(source, workspaceFolder string, doc document.Document, text string) []protocol.Diagnostic {
	diagnostics, _ := s.CalculateDiagnostics(context.Background(), source, doc)
	return diagnostics
}

func (c *ServiceImpl) SupportsLanguageIdentifier(languageIdentifier protocol.LanguageIdentifier) bool {
	return languageIdentifier == protocol.DockerfileLanguage
}
