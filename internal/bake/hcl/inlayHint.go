package hcl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func InlayHint(docs *document.Manager, doc document.BakeHCLDocument, rng protocol.Range) ([]protocol.InlayHint, error) {
	body, ok := doc.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	input := doc.Input()
	hints := []protocol.InlayHint{}
	for _, block := range body.Blocks {
		if block.Type == "target" && len(block.Labels) > 0 {
			if attribute, ok := block.Body.Attributes["args"]; ok {
				if expr, ok := attribute.Expr.(*hclsyntax.ObjectConsExpr); ok && len(expr.Items) > 0 {
					dockerfilePath, err := EvaluateDockerfilePath(block, doc.URI())
					if dockerfilePath != "" && err == nil {
						_, nodes := OpenDockerfile(context.Background(), docs, dockerfilePath)
						args := map[string]string{}
						for _, child := range nodes {
							if strings.EqualFold(child.Value, "ARG") {
								child = child.Next
								for child != nil {
									value := child.Value
									idx := strings.Index(value, "=")
									if idx != -1 {
										defaultValue := value[idx+1:]
										if defaultValue != "" {
											args[value[:idx]] = defaultValue
										}
									}
									child = child.Next
								}
							}
						}

						lines := strings.Split(string(input), "\n")
						for _, item := range expr.Items {
							itemRange := item.KeyExpr.Range()
							if insideProtocol(rng, itemRange.Start) || insideProtocol(rng, itemRange.End) {
								if value, ok := args[string(input[itemRange.Start.Byte:itemRange.End.Byte])]; ok {
									hints = append(hints, protocol.InlayHint{
										Label:       fmt.Sprintf("(default value: %v)", value),
										PaddingLeft: types.CreateBoolPointer(true),
										Position: protocol.Position{
											Line:      uint32(itemRange.Start.Line) - 1,
											Character: uint32(len(lines[itemRange.Start.Line-1])),
										},
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return hints, nil
}

func insideProtocol(rng protocol.Range, position hcl.Pos) bool {
	line := uint32(position.Line - 1)
	character := uint32(position.Column - 1)
	if rng.Start.Line < line {
		if line < rng.End.Line {
			return true
		} else if line == rng.End.Line {
			return character <= rng.End.Character
		}
		return false
	} else if rng.Start.Line == line {
		if line < rng.End.Line {
			return rng.Start.Character <= character
		}
		return rng.Start.Character <= character && character <= rng.End.Character
	}
	return false
}
