package hcl

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker-language-server/internal/bake/hcl/parser"
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Hover(ctx context.Context, params *protocol.HoverParams, document document.BakeHCLDocument) (*protocol.Hover, error) {
	body, ok := document.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	input := document.Input()
	variable := hoveredVariableName(input, body.Blocks, params.Position)
	if variable != "" {
		for _, block := range body.Blocks {
			if block.Type == "variable" && len(block.Labels) > 0 && block.Labels[0] == variable {
				if attribute, ok := block.Body.Attributes["default"]; ok {
					value := string(input[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte])
					return &protocol.Hover{
						Contents: protocol.MarkupContent{
							Kind:  "markdown",
							Value: value,
						},
					}, nil
				}
			}
		}
	}

	filename := string(params.TextDocument.URI)
	hclPos := parser.ConvertToHCLPosition(string(document.Input()), int(params.Position.Line), int(params.Position.Character))
	hover, err := document.Decoder().HoverAtPos(ctx, filename, hclPos)
	if err != nil {
		var positionalError *decoder.PositionalError
		if !errors.As(err, &positionalError) {
			return nil, fmt.Errorf("hover analysis encountered an error: %w", err)
		}
		return nil, nil
	}
	if hover == nil {
		return nil, nil
	}
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: hover.Content.Value,
		},
	}, nil
}

func hoveredVariableName(input []byte, blocks hclsyntax.Blocks, position protocol.Position) string {
	for _, block := range blocks {
		if isInsideBodyRangeLines(block.Body, int(position.Line+1)) {
			if block.Type == "variable" && len(block.LabelRanges) > 0 && isInsideRange(block.LabelRanges[0], position) {
				label := string(input[block.LabelRanges[0].Start.Byte:block.LabelRanges[0].End.Byte])
				if Quoted(label) {
					if block.LabelRanges[0].Start.Column == int(position.Character+1) || block.LabelRanges[0].End.Column == int(position.Character+1) {
						return ""
					}
				}
				return block.Labels[0]
			}

			for _, attribute := range block.Body.Attributes {
				if isInsideRange(attribute.Expr.Range(), position) {
					name := extractVariableName(input, attribute.Expr)
					if name != "" {
						return name
					}
				}
			}
		}
	}
	return ""
}

func extractVariableName(input []byte, expression hclsyntax.Expression) string {
	if tupleCons, ok := expression.(*hclsyntax.TupleConsExpr); ok {
		for _, expr := range tupleCons.Exprs {
			name := extractVariableName(input, expr)
			if name != "" {
				return name
			}
		}
	}

	if scope, ok := expression.(*hclsyntax.ScopeTraversalExpr); ok {
		return string(input[scope.SrcRange.Start.Byte:scope.SrcRange.End.Byte])
	}

	if wrap, ok := expression.(*hclsyntax.TemplateWrapExpr); ok {
		return extractVariableName(input, wrap.Wrapped)
	}

	if template, ok := expression.(*hclsyntax.TemplateExpr); ok {
		for _, part := range template.Parts {
			name := extractVariableName(input, part)
			if name != "" {
				return name
			}
		}
	}

	if objectCons, ok := expression.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range objectCons.Items {
			name := extractVariableName(input, item.ValueExpr)
			if name != "" {
				return name
			}
		}
	}

	return ""
}
