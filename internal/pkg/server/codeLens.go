package server

import (
	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentCodeLens(ctx *glsp.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	if doc.LanguageIdentifier() == protocol.DockerBakeLanguage {
		return hcl.CodeLens(ctx.Context, string(params.TextDocument.URI), doc.(document.BakeHCLDocument))
	}
	return nil, nil
}
