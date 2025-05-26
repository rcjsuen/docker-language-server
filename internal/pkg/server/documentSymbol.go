package server

import (
	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/compose"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentDocumentSymbol(ctx *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	language := doc.LanguageIdentifier()
	if language == protocol.DockerBakeLanguage {
		return hcl.DocumentSymbol(ctx.Context, string(params.TextDocument.URI), doc.(document.BakeHCLDocument))
	} else if language == protocol.DockerComposeLanguage && s.composeSupport {
		return compose.DocumentSymbol(ctx.Context, doc.(document.ComposeDocument))
	}

	return nil, nil
}
