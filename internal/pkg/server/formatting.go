package server

import (
	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/compose"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentFormatting(ctx *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if doc.LanguageIdentifier() == protocol.DockerComposeLanguage {
		return compose.Formatting(doc.(document.ComposeDocument), params.Options)
	} else if doc.LanguageIdentifier() == protocol.DockerBakeLanguage {
		return hcl.Formatting(doc.(document.BakeHCLDocument), params.Options)
	}
	return nil, nil
}
