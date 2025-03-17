package server

import (
	"errors"
	"strings"

	"github.com/docker/docker-language-server/internal/bake/hcl"
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
	language := doc.LanguageIdentifier()
	if language == protocol.DockerBakeLanguage {
		return hcl.Hover(ctx.Context, params, doc.(document.BakeHCLDocument))
	} else if language == protocol.DockerComposeLanguage {
		return nil, nil
	}

	dockerfileDocument, ok := doc.(document.DockerfileDocument)
	if ok {
		instruction := dockerfileDocument.Instruction(params.Position)
		if instruction != nil && strings.EqualFold(instruction.Value, "FROM") && instruction.Next != nil {
			return s.scoutService.Hover(ctx.Context, instruction.Next.Value)
		}
		return nil, nil
	}
	return nil, errors.New("URI did not map to a recognized document")
}
