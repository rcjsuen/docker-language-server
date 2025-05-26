package server

import (
	"errors"
	"strings"

	"github.com/docker/docker-language-server/internal/bake/hcl"
	"github.com/docker/docker-language-server/internal/compose"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	doc, err := s.docs.Read(ctx.Context, uri.URI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	switch doc.LanguageIdentifier() {
	case protocol.DockerBakeLanguage:
		return hcl.Hover(ctx.Context, params, doc.(document.BakeHCLDocument))
	case protocol.DockerComposeLanguage:
		if s.composeSupport {
			return compose.Hover(ctx.Context, params, doc.(document.ComposeDocument))
		}
		return nil, nil
	case protocol.DockerfileLanguage:
		instruction := doc.(document.DockerfileDocument).Instruction(params.Position)
		if instruction != nil && strings.EqualFold(instruction.Value, "FROM") && instruction.Next != nil {
			return s.scoutService.Hover(ctx.Context, params.TextDocument.URI, instruction.Next.Value)
		}
		return nil, nil
	}
	return nil, errors.New("URI did not map to a recognized document")
}
