package hcl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func CodeLens(ctx context.Context, filename string, doc document.BakeHCLDocument) ([]protocol.CodeLens, error) {
	body, ok := doc.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	url, err := url.Parse(filename)
	if err != nil {
		return nil, fmt.Errorf("could not parse URI (%v): %w", filename, err)
	}

	filename = url.Path
	result := []protocol.CodeLens{}
	for _, block := range body.Blocks {
		if len(block.Labels) > 0 {
			if block.Type == "group" {
				rng := protocol.Range{
					Start: protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
					End:   protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
				}
				result = append(result, createCodeLens("Build", filename, "build", block.Labels[0], rng))
				result = append(result, createCodeLens("Check", filename, "check", block.Labels[0], rng))
				result = append(result, createCodeLens("Print", filename, "print", block.Labels[0], rng))
			} else if block.Type == "target" {
				rng := protocol.Range{
					Start: protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
					End:   protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
				}
				result = append(result, createCodeLens("Build", filename, "build", block.Labels[0], rng))
				result = append(result, createCodeLens("Check", filename, "check", block.Labels[0], rng))
				result = append(result, createCodeLens("Print", filename, "print", block.Labels[0], rng))
			}
		}
	}
	return result, nil
}

func createCodeLens(title, filename, call, target string, rng protocol.Range) protocol.CodeLens {
	return protocol.CodeLens{
		Range: rng,
		Command: &protocol.Command{
			Title:   title,
			Command: types.BakeBuildCommandId,
			Arguments: []any{
				map[string]string{
					"file":   filename,
					"call":   call,
					"target": target,
					"cwd":    filepath.Dir(types.StripLeadingSlash(filename)),
				},
			},
		},
	}
}
