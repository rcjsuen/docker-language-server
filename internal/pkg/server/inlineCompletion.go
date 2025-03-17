package server

import (
	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentInlineCompletion(ctx *glsp.Context, params *protocol.InlineCompletionParams) (any, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if doc.LanguageIdentifier() == protocol.DockerBakeLanguage {
		return hcl.InlineCompletion(ctx.Context, params, s.docs, doc.(document.BakeHCLDocument))
	}
	return nil, nil
}
