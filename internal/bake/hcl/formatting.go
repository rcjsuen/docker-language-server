package hcl

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func indentation(options protocol.FormattingOptions) string {
	if insertSpaces, ok := options[protocol.FormattingOptionInsertSpaces].(bool); ok && insertSpaces {
		sb := strings.Builder{}
		tabSize := int(options[protocol.FormattingOptionTabSize].(float64))
		for range tabSize {
			sb.WriteString(" ")
		}
		return sb.String()
	}
	return "\t"
}

func Formatting(doc document.BakeHCLDocument, options protocol.FormattingOptions) ([]protocol.TextEdit, error) {
	body, ok := doc.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	input := doc.Input()
	_, diagnostics := hclsyntax.ParseConfig(input, "", hcl.InitialPos)
	if diagnostics.HasErrors() {
		return nil, nil
	}

	lines := strings.Split(string(input), "\n")
	edits := []protocol.TextEdit{}
	indentation := indentation(options)
	for _, block := range body.Blocks {
		edit := indentLine(lines, block.TypeRange.Start.Line-1, "")
		if edit != nil {
			edits = append(edits, *edit)
		}

		if len(block.LabelRanges) > 0 {
			edit := spaceBetween(input, block.TypeRange.End, block.LabelRanges[0].Start)
			if edit != nil {
				edits = append(edits, *edit)
			}

			for i := range len(block.LabelRanges) - 1 {
				edit := spaceBetween(input, block.LabelRanges[i].End, block.LabelRanges[i+1].Start)
				if edit != nil {
					edits = append(edits, *edit)
				}
			}

			edit = spaceBetween(input, block.LabelRanges[len(block.LabelRanges)-1].End, block.OpenBraceRange.Start)
			if edit != nil {
				edits = append(edits, *edit)
			}
		} else {
			edit := spaceBetween(input, block.TypeRange.End, block.OpenBraceRange.Start)
			if edit != nil {
				edits = append(edits, *edit)
			}
		}

		if block.Body.SrcRange.Start.Line == block.Body.SrcRange.End.Line {
			if len(block.Body.Attributes) == 1 {
				for _, attribute := range block.Body.Attributes {
					edit := spaceBetween(input, block.OpenBraceRange.End, attribute.NameRange.Start)
					if edit != nil {
						edits = append(edits, *edit)
					}

					edit = spaceBetween(input, attribute.NameRange.End, attribute.EqualsRange.Start)
					if edit != nil {
						edits = append(edits, *edit)
					}

					edit = spaceBetween(input, attribute.EqualsRange.End, attribute.Expr.Range().Start)
					if edit != nil {
						edits = append(edits, *edit)
					}

					edit = spaceBetween(input, attribute.Expr.Range().End, block.CloseBraceRange.Start)
					if edit != nil {
						edits = append(edits, *edit)
					}
					break
				}
			}
			continue
		}

		for _, attribute := range block.Body.Attributes {
			edit := indentLine(lines, attribute.SrcRange.Start.Line-1, indentation)
			if edit != nil {
				edits = append(edits, *edit)
			}

			edit = spaceBetween(input, attribute.EqualsRange.End, attribute.Expr.Range().Start)
			if edit != nil {
				edits = append(edits, *edit)
			}

			if expr, ok := attribute.Expr.(*hclsyntax.ObjectConsExpr); ok && attribute.SrcRange.Start.Line != attribute.SrcRange.End.Line {
				if len(expr.Items) > 0 {
					if attribute.SrcRange.Start.Line == expr.Items[0].KeyExpr.Range().Start.Line {
						objectEdits := formatAttributeObjectValues(input, expr.Items[1:], lines, indentation+indentation)
						edits = append(edits, objectEdits...)
					} else {
						objectEdits := formatAttributeObjectValues(input, expr.Items, lines, indentation+indentation)
						edits = append(edits, objectEdits...)
					}
				}

				edit := indentLine(lines, expr.SrcRange.End.Line-1, indentation)
				if edit != nil {
					edits = append(edits, *edit)
				}
			}
		}

		edit = indentLine(lines, block.CloseBraceRange.End.Line-1, "")
		if edit != nil {
			edits = append(edits, *edit)
		}
	}

	return edits, nil
}

func indentLine(lines []string, line int, indentation string) *protocol.TextEdit {
	lineContent := lines[line]
	leftTrimmed := strings.TrimLeftFunc(lineContent, unicode.IsSpace)
	excess := len(lineContent) - len(leftTrimmed)
	if lineContent[0:excess] == indentation {
		return nil
	}

	return &protocol.TextEdit{
		NewText: indentation,
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(line), Character: 0},
			End:   protocol.Position{Line: uint32(line), Character: uint32(excess)},
		},
	}
}

func formatAttributeObjectValues(input []byte, items []hclsyntax.ObjectConsItem, lines []string, indentation string) []protocol.TextEdit {
	edits := []protocol.TextEdit{}
	for _, item := range items {
		edit := indentLine(lines, item.KeyExpr.Range().Start.Line-1, indentation)
		if edit != nil {
			edits = append(edits, *edit)
		}

		edit = formatRight(input, item.KeyExpr.Range().End, item.ValueExpr.Range().Start)
		if edit != nil {
			edits = append(edits, *edit)
		}
	}
	return edits
}

func spaceBetween(input []byte, start hcl.Pos, end hcl.Pos) *protocol.TextEdit {
	after := string(input[start.Byte:end.Byte])
	if len(after) != 1 {
		trimmed := strings.TrimSpace(after)
		if trimmed == "" {
			return &protocol.TextEdit{
				NewText: " ",
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(start.Line - 1), Character: uint32(start.Column - 1)},
					End:   protocol.Position{Line: uint32(start.Line - 1), Character: uint32(end.Column - 1)},
				},
			}
		} else if after != fmt.Sprintf(" %v ", trimmed) {
			return &protocol.TextEdit{
				NewText: fmt.Sprintf(" %v ", trimmed),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(start.Line - 1), Character: uint32(start.Column - 1)},
					End:   protocol.Position{Line: uint32(start.Line - 1), Character: uint32(end.Column - 1)},
				},
			}
		}
	}

	return nil
}

func formatRight(input []byte, start hcl.Pos, end hcl.Pos) *protocol.TextEdit {
	after := string(input[start.Byte:end.Byte])
	if len(after) != 1 {
		trimmed := strings.TrimRightFunc(after, unicode.IsSpace)
		if trimmed == "" {
			return &protocol.TextEdit{
				NewText: " ",
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(start.Line - 1), Character: uint32(start.Column - 1)},
					End:   protocol.Position{Line: uint32(start.Line - 1), Character: uint32(end.Column - 1)},
				},
			}
		} else if after != fmt.Sprintf("%v ", trimmed) {
			return &protocol.TextEdit{
				NewText: fmt.Sprintf("%v ", trimmed),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(start.Line - 1), Character: uint32(start.Column - 1)},
					End:   protocol.Position{Line: uint32(start.Line - 1), Character: uint32(end.Column - 1)},
				},
			}
		}
	}

	return nil
}
