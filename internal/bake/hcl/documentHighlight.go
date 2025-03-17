package hcl

import (
	"errors"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func DocumentHighlight(document document.BakeHCLDocument, position protocol.Position) ([]protocol.DocumentHighlight, error) {
	body, ok := document.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	bytes := document.Input()
	target := ""
	for _, block := range body.Blocks {
		if block.Type == "group" {
			if targets, ok := block.Body.Attributes["targets"]; ok {
				if expr, ok := targets.Expr.(*hclsyntax.TupleConsExpr); ok {
					for _, item := range expr.Exprs {
						if template, ok := item.(*hclsyntax.TemplateExpr); ok && len(template.Parts) == 1 && isInsideRange(template.Parts[0].Range(), position) {
							value, _ := template.Parts[0].Value(&hcl.EvalContext{})
							target = value.AsString()
							break
						}
					}
				}
			}
		} else if block.Type == "target" && len(block.LabelRanges) > 0 && isInsideRange(block.LabelRanges[0], position) {
			label := string(bytes[block.LabelRanges[0].Start.Byte:block.LabelRanges[0].End.Byte])
			if Quoted(label) {
				unquotedRange := hcl.Range{
					Start: hcl.Pos{
						Line:   block.LabelRanges[0].Start.Line,
						Column: block.LabelRanges[0].Start.Column + 1,
					},
					End: hcl.Pos{
						Line:   block.LabelRanges[0].End.Line,
						Column: block.LabelRanges[0].End.Column - 1,
					},
				}
				if isInsideRange(unquotedRange, position) {
					target = label[1 : len(label)-1]
				}
			} else {
				target = label
			}
		}
	}

	if target != "" {
		ranges := []protocol.DocumentHighlight{}
		for _, block := range body.Blocks {
			if block.Type == "group" {
				if targets, ok := block.Body.Attributes["targets"]; ok {
					if expr, ok := targets.Expr.(*hclsyntax.TupleConsExpr); ok {
						for _, item := range expr.Exprs {
							if template, ok := item.(*hclsyntax.TemplateExpr); ok && len(template.Parts) == 1 {
								value, _ := template.Parts[0].Value(&hcl.EvalContext{})
								if target == value.AsString() {
									ranges = append(ranges, protocol.DocumentHighlight{
										Kind:  types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindRead),
										Range: createProtocolRange(template.Parts[0].Range(), false),
									})
								}
							}
						}
					}
				}
			} else if block.Type == "target" && len(block.LabelRanges) > 0 {
				label := string(bytes[block.LabelRanges[0].Start.Byte:block.LabelRanges[0].End.Byte])
				quoted := Quoted(label)
				label = strings.TrimPrefix(label, "\"")
				label = strings.TrimSuffix(label, "\"")

				if target == label {
					ranges = append(ranges, protocol.DocumentHighlight{
						Kind:  types.CreateDocumentHighlightKindPointer(protocol.DocumentHighlightKindWrite),
						Range: createProtocolRange(block.LabelRanges[0], quoted),
					})
				}
			}
		}
		return ranges, nil
	}
	return nil, nil
}

func Quoted(s string) bool {
	return s[0] == 34 && s[len(s)-1] == 34
}

func createProtocolRange(hclRange hcl.Range, quoted bool) protocol.Range {
	if quoted {
		return protocol.Range{
			Start: protocol.Position{
				Line:      uint32(hclRange.Start.Line - 1),
				Character: uint32(hclRange.Start.Column),
			},
			End: protocol.Position{
				Line:      uint32(hclRange.End.Line - 1),
				Character: uint32(hclRange.End.Column - 2),
			},
		}
	}

	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(hclRange.Start.Line - 1),
			Character: uint32(hclRange.Start.Column - 1),
		},
		End: protocol.Position{
			Line:      uint32(hclRange.End.Line - 1),
			Character: uint32(hclRange.End.Column - 1),
		},
	}
}
