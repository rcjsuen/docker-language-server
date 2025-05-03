package server

import (
	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/compose"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentDocumentHighlight(ctx *glsp.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if doc.LanguageIdentifier() == protocol.DockerBakeLanguage {
		return hcl.DocumentHighlight(doc.(document.BakeHCLDocument), params.Position)
	} else if doc.LanguageIdentifier() == protocol.DockerComposeLanguage {
		return compose.DocumentHighlight(doc.(document.ComposeDocument), params.Position)
	}
	return nil, nil
}
