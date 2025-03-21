package hcl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"go.lsp.dev/uri"
)

func LocalDockerfile(u *url.URL) (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Abs(path.Join(u.Path[1:], "../Dockerfile"))
	}
	return filepath.Abs(path.Join(u.Path, "../Dockerfile"))
}

func ParseReferencedDockerfile(documentURI uri.URI, document document.BakeHCLDocument, hclLine, hclCharacter int) (string, error) {
	body, ok := document.File().Body.(*hclsyntax.Body)
	if !ok {
		return "", errors.New("unrecognized body in HCL document")
	}
	url, err := url.Parse(string(documentURI))
	if err != nil {
		return "", fmt.Errorf("LSP client sent invalid URI: %v", string(documentURI))
	}

	input := document.Input()
	for _, b := range body.Blocks {
		if isInsideBodyRangeLines(b.Body, hclLine) {
			attributes := b.Body.Attributes
			dockerfilePath := ""
			if _, ok := attributes["dockerfile-inline"]; ok {
				break
			}

			if attribute, ok := attributes["dockerfile"]; ok {
				dockerfilePath = string(input[attribute.Expr.Range().Start.Byte:attribute.Expr.Range().End.Byte])
				dockerfilePath = strings.TrimPrefix(dockerfilePath, "\"")
				dockerfilePath = strings.TrimSuffix(dockerfilePath, "\"")
				if runtime.GOOS == "windows" {
					dockerfilePath, err = filepath.Abs(path.Join(url.Path[1:], fmt.Sprintf("../%v", dockerfilePath)))
				} else {
					dockerfilePath, err = filepath.Abs(path.Join(url.Path, fmt.Sprintf("../%v", dockerfilePath)))
				}
				if err != nil {
					break
				}
			}

			if dockerfilePath == "" {
				// no dockerfile attribute found, default to ./Dockerfile relative to the HCL file
				dockerfilePath, err = LocalDockerfile(url)
				if err != nil {
					break
				}
			}
			return dockerfilePath, nil
		}
	}
	return "", nil
}

func Definition(ctx context.Context, definitionLinkSupport bool, manager *document.Manager, documentURI uri.URI, document document.BakeHCLDocument, position protocol.Position) (any, error) {
	body, ok := document.File().Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unrecognized body in HCL document")
	}

	dockerfilePath, err := ParseReferencedDockerfile(documentURI, document, int(position.Line+1), int(position.Character+1))
	if err != nil {
		return nil, nil
	}

	for _, b := range body.Blocks {
		if isInsideBodyRangeLines(b.Body, int(position.Line+1)) {
			for _, attribute := range b.Body.Attributes {
				if isInsideRange(attribute.Expr.Range(), position) {
					input := document.Input()
					return ResolveAttributeValue(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, b, attribute), nil
				}
			}
		}
	}

	for _, attribute := range body.Attributes {
		if isInsideRange(attribute.NameRange, position) {
			return []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(attribute.NameRange.Start.Line) - 1,
							Character: uint32(attribute.NameRange.Start.Column) - 1,
						},
						End: protocol.Position{
							Line:      uint32(attribute.NameRange.End.Line) - 1,
							Character: uint32(attribute.NameRange.End.Column) - 1,
						},
					},
					URI: string(documentURI),
				},
			}, nil
		}

		if isInsideRange(attribute.SrcRange, position) {
			return ResolveAttributeValue(ctx, definitionLinkSupport, manager, document.Input(), dockerfilePath, body, documentURI, position, nil, attribute), nil
		}
	}
	return nil, nil
}

func ResolveAttributeValue(ctx context.Context, definitionLinkSupport bool, manager *document.Manager, input []byte, dockerfilePath string, body *hclsyntax.Body, documentURI uri.URI, position protocol.Position, sourceBlock *hclsyntax.Block, attribute *hclsyntax.Attribute) any {
	if tupleConsExpr, ok := attribute.Expr.(*hclsyntax.TupleConsExpr); ok {
		for _, e := range tupleConsExpr.Exprs {
			if isInsideRange(e.Range(), position) {
				if templateExpr, ok := e.(*hclsyntax.TemplateExpr); ok {
					if templateExpr.IsStringLiteral() {
						// look up a target reference if it's inside a
						// target block's inherits attribute, or a
						// group block's targets attribute
						if (sourceBlock.Type == "target" && attribute.Name == "inherits") ||
							(sourceBlock.Type == "group" && attribute.Name == "targets") {
							value, _ := templateExpr.Value(&hcl.EvalContext{})
							target := value.AsString()
							return CalculateBlockLocation(input, body, documentURI, "target", target, false)
						}
					}
				}

				return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attribute.Name, e)
			}
		}
	}

	return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attribute.Name, attribute.Expr)
}

func ResolveExpression(ctx context.Context, definitionLinkSupport bool, manager *document.Manager, input []byte, dockerfilePath string, body *hclsyntax.Body, documentURI uri.URI, position protocol.Position, sourceBlock *hclsyntax.Block, attributeName string, expression hclsyntax.Expression) any {
	if templateExpr, ok := expression.(*hclsyntax.TemplateExpr); ok {
		for _, part := range templateExpr.Parts {
			if isInsideRange(part.Range(), position) {
				return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, part)
			}
		}
	}

	if literalValueExpr, ok := expression.(*hclsyntax.LiteralValueExpr); ok && sourceBlock != nil && sourceBlock.Type == "target" {
		if attributeName == "no-cache-filter" || attributeName == "target" {
			value, _ := literalValueExpr.Value(&hcl.EvalContext{})
			target := value.AsString()

			bytes, nodes := OpenDockerfile(ctx, manager, dockerfilePath)
			lines := strings.Split(string(bytes), "\n")
			for _, child := range nodes {
				if strings.EqualFold(child.Value, "FROM") {
					if child.Next != nil && child.Next.Next != nil && strings.EqualFold(child.Next.Next.Value, "AS") && child.Next.Next.Next != nil && child.Next.Next.Next.Value == target {
						targetRange := protocol.Range{
							Start: protocol.Position{Line: uint32(child.StartLine) - 1, Character: 0},
							End:   protocol.Position{Line: uint32(child.EndLine) - 1, Character: uint32(len(lines[child.EndLine-1]))},
						}

						if runtime.GOOS == "windows" {
							dockerfilePath = "/" + filepath.ToSlash(dockerfilePath)
						}

						if !definitionLinkSupport {
							return []protocol.Location{
								{
									Range: targetRange,
									URI:   protocol.URI(fmt.Sprintf("file://%v", dockerfilePath)),
								},
							}
						}

						return []protocol.LocationLink{
							{
								OriginSelectionRange: &protocol.Range{
									Start: protocol.Position{Line: uint32(literalValueExpr.Range().Start.Line) - 1, Character: uint32(literalValueExpr.Range().Start.Column) - 1},
									End:   protocol.Position{Line: uint32(literalValueExpr.Range().End.Line) - 1, Character: uint32(uint32(literalValueExpr.Range().End.Column) - 1)},
								},
								TargetRange:          targetRange,
								TargetSelectionRange: targetRange,
								TargetURI:            protocol.URI(fmt.Sprintf("file://%v", dockerfilePath)),
							},
						}
					}
				}
			}
		}
	}

	if forExpr, ok := expression.(*hclsyntax.ForExpr); ok {
		if isInsideRange(forExpr.CollExpr.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, forExpr.CollExpr)
		}

		if isInsideRange(forExpr.CondExpr.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, forExpr.CondExpr)
		}
	}

	if objectConsExpression, ok := expression.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range objectConsExpression.Items {
			if isInsideRange(item.KeyExpr.Range(), position) {
				if attributeName == "args" && sourceBlock.Type == "target" {
					start := item.KeyExpr.Range().Start.Byte
					end := item.KeyExpr.Range().End.Byte
					if LiteralValue(item.KeyExpr) {
						start++
						end--
					}
					arg := string(input[start:end])
					bytes, nodes := OpenDockerfile(ctx, manager, dockerfilePath)
					lines := strings.Split(string(bytes), "\n")
					for _, child := range nodes {
						if strings.EqualFold(child.Value, "ARG") {
							node := child
							child = child.Next
							for child != nil {
								value := child.Value
								idx := strings.Index(value, "=")
								if idx != -1 {
									value = value[:idx]
								}

								if value == arg {
									if runtime.GOOS == "windows" {
										dockerfilePath = "/" + filepath.ToSlash(dockerfilePath)
									}
									return []protocol.Location{
										{
											Range: protocol.Range{
												Start: protocol.Position{Line: uint32(node.StartLine) - 1, Character: 0},
												End:   protocol.Position{Line: uint32(node.EndLine) - 1, Character: uint32(len(lines[node.EndLine-1]))},
											},
											URI: protocol.URI(fmt.Sprintf("file://%v", dockerfilePath)),
										},
									}
								}
								child = child.Next
							}
						}
					}
					return nil
				}
			}

			if isInsideRange(item.ValueExpr.Range(), position) {
				return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, item.ValueExpr)
			}
		}
	}

	if binaryExpression, ok := expression.(*hclsyntax.BinaryOpExpr); ok {
		if isInsideRange(binaryExpression.LHS.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, binaryExpression.LHS)
		}

		if isInsideRange(binaryExpression.RHS.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, binaryExpression.RHS)
		}

		return nil
	}

	if conditional, ok := expression.(*hclsyntax.ConditionalExpr); ok {
		if isInsideRange(conditional.Condition.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, conditional.Condition)
		}

		if isInsideRange(conditional.TrueResult.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, conditional.TrueResult)
		}

		if isInsideRange(conditional.FalseResult.Range(), position) {
			return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, conditional.FalseResult)
		}
		return nil
	}

	if _, ok := expression.(*hclsyntax.ScopeTraversalExpr); ok {
		name := string(input[expression.Range().Start.Byte:expression.Range().End.Byte])
		return CalculateBlockLocation(input, body, documentURI, "variable", name, true)
	}

	if templateWrapExpr, ok := expression.(*hclsyntax.TemplateWrapExpr); ok {
		return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, templateWrapExpr.Wrapped)
	}

	if functionCallExpr, ok := expression.(*hclsyntax.FunctionCallExpr); ok {
		if isInsideRange(functionCallExpr.NameRange, position) {
			return CalculateBlockLocation(input, body, documentURI, "function", functionCallExpr.Name, true)
		}

		for _, arg := range functionCallExpr.Args {
			if isInsideRange(arg.Range(), position) {
				return ResolveExpression(ctx, definitionLinkSupport, manager, input, dockerfilePath, body, documentURI, position, sourceBlock, attributeName, arg)
			}
		}
	}
	return nil
}

// CalculateBlockLocation finds a block with the specified name and
// returns it. If variable is true then it will also look at the
// top-level attributes of the HCL file and resolve to those if the
// names match.
func CalculateBlockLocation(input []byte, body *hclsyntax.Body, documentURI uri.URI, blockName, name string, variable bool) any {
	for _, b := range body.Blocks {
		if b.Type == blockName && b.Labels[0] == name {
			startCharacter := uint32(b.LabelRanges[0].Start.Column)
			endCharacter := uint32(b.LabelRanges[0].End.Column)
			variableNameDeclaration := string(input[b.LabelRanges[0].Start.Byte:b.LabelRanges[0].End.Byte])
			if Quoted(variableNameDeclaration) {
				endCharacter -= 2
			} else {
				startCharacter--
				endCharacter--
			}
			return []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(b.LabelRanges[0].Start.Line) - 1,
							Character: startCharacter,
						},
						End: protocol.Position{
							Line:      uint32(b.LabelRanges[0].End.Line) - 1,
							Character: endCharacter,
						},
					},
					URI: string(documentURI),
				},
			}
		}
	}

	if attribute, ok := body.Attributes[name]; ok && variable {
		return []protocol.Location{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      uint32(attribute.NameRange.Start.Line) - 1,
						Character: uint32(attribute.NameRange.Start.Column) - 1,
					},
					End: protocol.Position{
						Line:      uint32(attribute.NameRange.End.Line) - 1,
						Character: uint32(attribute.NameRange.End.Column) - 1,
					},
				},
				URI: string(documentURI),
			},
		}
	}
	return nil
}

func ParseDockerfile(dockerfilePath string) ([]byte, *parser.Result, error) {
	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return nil, nil, err
	}
	result, err := parser.Parse(bytes.NewReader(dockerfileBytes))
	return dockerfileBytes, result, err
}

func OpenDockerfile(ctx context.Context, manager *document.Manager, path string) ([]byte, []*parser.Node) {
	doc := manager.Get(ctx, uri.URI(fmt.Sprintf("file://%v", path)))
	if doc != nil {
		if dockerfile, ok := doc.(document.DockerfileDocument); ok {
			return dockerfile.Input(), dockerfile.Nodes()
		}
	}
	dockerfileBytes, result, err := ParseDockerfile(path)
	if err != nil {
		return nil, nil
	}
	return dockerfileBytes, result.AST.Children
}
