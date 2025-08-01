package hcl

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func CodeLens(ctx context.Context, documentURI string, doc document.BakeHCLDocument) ([]protocol.CodeLens, error) {
	body, ok := doc.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	dp, err := doc.DocumentPath()
	if err != nil {
		return nil, fmt.Errorf("could not parse URI (%v): %w", documentURI, err)
	}

	_, cwd := types.Concatenate(dp.Folder, ".", dp.WSLDollarSignHost)
	result := []protocol.CodeLens{}
	for _, block := range body.Blocks {
		if len(block.Labels) > 0 {
			switch block.Type {
			case "group":
				rng := protocol.Range{
					Start: protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
					End:   protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
				}
				result = append(result, createCodeLens("Build", cwd, "build", block.Labels[0], rng))
				result = append(result, createCodeLens("Check", cwd, "check", block.Labels[0], rng))
				result = append(result, createCodeLens("Print", cwd, "print", block.Labels[0], rng))
			case "target":
				rng := protocol.Range{
					Start: protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
					End:   protocol.Position{Line: uint32(block.Range().Start.Line - 1)},
				}
				result = append(result, createCodeLens("Build", cwd, "build", block.Labels[0], rng))
				result = append(result, createCodeLens("Check", cwd, "check", block.Labels[0], rng))
				result = append(result, createCodeLens("Print", cwd, "print", block.Labels[0], rng))
			}
		}
	}
	return result, nil
}

func createCodeLens(title, cwd, call, target string, rng protocol.Range) protocol.CodeLens {
	return protocol.CodeLens{
		Range: rng,
		Command: &protocol.Command{
			Title:   title,
			Command: types.BakeBuildCommandId,
			Arguments: []any{
				map[string]string{
					"call":   call,
					"target": target,
					"cwd":    cwd,
				},
			},
		},
	}
}
