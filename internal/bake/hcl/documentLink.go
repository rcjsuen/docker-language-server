package hcl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func DocumentLink(ctx context.Context, documentURI protocol.URI, document document.BakeHCLDocument) ([]protocol.DocumentLink, error) {
	body, ok := document.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}
	url, err := url.Parse(string(documentURI))
	if err != nil {
		return nil, fmt.Errorf("LSP client sent invalid URI: %v", string(documentURI))
	}

	bytes := document.Input()
	links := []protocol.DocumentLink{}
	for _, b := range body.Blocks {
		attributes := b.Body.Attributes
		for _, v := range attributes {
			if v.Name == "dockerfile" {
				dockerfilePath := string(bytes[v.Expr.Range().Start.Byte:v.Expr.Range().End.Byte])
				if !Quoted(dockerfilePath) {
					continue
				}

				dockerfilePath = strings.TrimPrefix(dockerfilePath, "\"")
				dockerfilePath = strings.TrimSuffix(dockerfilePath, "\"")
				dockerfilePath, err = types.AbsolutePath(url, dockerfilePath)
				if err == nil {
					links = append(links, protocol.DocumentLink{
						Range: protocol.Range{
							Start: protocol.Position{Line: uint32(v.SrcRange.Start.Line) - 1, Character: uint32(v.Expr.Range().Start.Column)},
							End:   protocol.Position{Line: uint32(v.SrcRange.Start.Line) - 1, Character: uint32(v.Expr.Range().End.Column - 2)},
						},
						Target:  types.CreateStringPointer(protocol.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(dockerfilePath), "/")))),
						Tooltip: types.CreateStringPointer(dockerfilePath),
					})
				}
			}
		}
	}
	return links, nil
}
