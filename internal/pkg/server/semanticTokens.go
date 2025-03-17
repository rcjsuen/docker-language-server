package server

import (
	"strings"

	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentSemanticTokensFull(ctx *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if strings.HasSuffix(string(params.TextDocument.URI), "hcl") {
		result, err := hcl.SemanticTokensFull(ctx.Context, doc.(document.BakeHCLDocument), string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		return &protocol.SemanticTokens{Data: result.Data}, nil
	}

	return nil, nil
}
