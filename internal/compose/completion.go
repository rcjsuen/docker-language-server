package compose

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type completionItemText struct {
	label         string
	newText       string
	documentation string
}

type textEditModifier struct {
	isInterested func(attributeName string, path []*ast.MappingValueNode) bool
	modify       func(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit
}

// extendingCurrentFile checks if the extends object's file attribute is
// pointing to the current file.
func extendingCurrentFile(u *url.URL, extendsNode *ast.MappingValueNode) bool {
	if extends, ok := extendsNode.Value.(*ast.MappingNode); ok {
		for _, extendsAttribute := range extends.Values {
			if extendsAttribute.Key.GetToken().Value == "file" {
				path, err := types.AbsolutePath(u, extendsAttribute.Value.GetToken().Value)
				if err != nil || filepath.ToSlash(u.Path) != filepath.ToSlash(path) {
					return false
				}
				break
			}
		}
	}
	return true
}

var buildTargetModifier = textEditModifier{
	isInterested: func(attributeName string, path []*ast.MappingValueNode) bool {
		return attributeName == "target" && len(path) == 3 && path[2].Key.GetToken().Value == "build"
	},
	modify: func(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit {
		if _, ok := path[2].Value.(*ast.NullNode); ok {
			dockerfilePath, err := types.LocalDockerfile(u)
			if err == nil {
				stages := findBuildStages(manager, dockerfilePath, "")
				if len(stages) > 0 {
					edit.NewText = fmt.Sprintf("%v%v", edit.NewText, createChoiceSnippetText(stages))
					return edit
				}
			}
		} else if mappingNode, ok := path[2].Value.(*ast.MappingNode); ok {
			dockerfileAttributePath := "Dockerfile"
			for _, buildAttribute := range mappingNode.Values {
				switch buildAttribute.Key.GetToken().Value {
				case "dockerfile_inline":
					return edit
				case "dockerfile":
					dockerfileAttributePath = buildAttribute.Value.GetToken().Value
				}
			}

			dockerfilePath, err := types.AbsolutePath(u, dockerfileAttributePath)
			if err == nil {
				stages := findBuildStages(manager, dockerfilePath, "")
				if len(stages) > 0 {
					edit.NewText = fmt.Sprintf("%v%v", edit.NewText, createChoiceSnippetText(stages))
					return edit
				}
			}
		}
		return edit
	},
}

var serviceSuggestionModifier = textEditModifier{
	isInterested: func(attributeName string, path []*ast.MappingValueNode) bool {
		return attributeName == "service" && len(path) == 3 && path[0].Key.GetToken().Value == "services" && path[2].Key.GetToken().Value == "extends"
	},
	modify: func(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit {
		if extendingCurrentFile(u, path[2]) {
			services := []completionItemText{}
			for _, service := range findDependencies(file, "services") {
				if service != path[1].Key.GetToken().Value {
					services = append(services, completionItemText{newText: service})
				}
			}
			if len(services) > 0 {
				edit.NewText = fmt.Sprintf("%v%v", edit.NewText, createChoiceSnippetText(services))
			}
		}
		return edit
	},
}

var serviceProviderModifier = textEditModifier{
	isInterested: func(attributeName string, path []*ast.MappingValueNode) bool {
		return attributeName == "provider" && len(path) == 2 && path[0].Key.GetToken().Value == "services"
	},
	modify: func(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit {
		edit.NewText = fmt.Sprintf("provider:\n%vtype: ${1:model}\n%voptions:\n%v  ${2:model}: ${3:ai/example-model}", spacing, spacing, spacing)
		return edit
	},
}

var serviceProviderTypeModifier = textEditModifier{
	isInterested: func(attributeName string, path []*ast.MappingValueNode) bool {
		return attributeName == "type" && len(path) == 3 && path[0].Key.GetToken().Value == "services" && path[2].Key.GetToken().Value == "provider"
	},
	modify: func(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit {
		edit.NewText = "type: ${1:model}"
		return edit
	},
}

var textEditModifiers = []textEditModifier{buildTargetModifier, serviceSuggestionModifier, serviceProviderModifier, serviceProviderTypeModifier}

func prefix(line string, character int) string {
	sb := strings.Builder{}
	for i := range character {
		if unicode.IsSpace(rune(line[i])) {
			sb.Reset()
		} else {
			sb.WriteByte(line[i])
		}
	}
	return sb.String()
}

func array(line string, character int) bool {
	isArray := false
	for i := range character {
		if unicode.IsSpace(rune(line[i])) {
			continue
		} else if line[i] == '-' {
			isArray = true
		} else if isArray && line[i] == ':' {
			return false
		}
	}
	return isArray
}

func createTopLevelItems() []protocol.CompletionItem {
	items := []protocol.CompletionItem{}
	for attributeName, schema := range schemaProperties() {
		item := protocol.CompletionItem{Label: attributeName}
		if schema.Description != "" {
			item.Documentation = schema.Description
		}
		items = append(items, item)
	}
	slices.SortFunc(items, func(a, b protocol.CompletionItem) int {
		return strings.Compare(a.Label, b.Label)
	})
	return items
}

func calculateTopLevelNodeOffset(file *ast.File) int {
	if len(file.Docs) == 1 {
		if m, ok := file.Docs[0].Body.(*ast.MappingNode); ok {
			return m.Values[0].Key.GetToken().Position.Column - 1
		}
	}
	return -1
}

func Completion(ctx context.Context, params *protocol.CompletionParams, manager *document.Manager, doc document.ComposeDocument) (*protocol.CompletionList, error) {
	u, err := url.Parse(params.TextDocument.URI)
	if err != nil {
		return nil, fmt.Errorf("LSP client sent invalid URI: %v", params.TextDocument.URI)
	}

	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}

	lspLine := int(params.Position.Line)
	topLevelNodeOffset := calculateTopLevelNodeOffset(file)
	if topLevelNodeOffset != -1 && params.Position.Character == uint32(topLevelNodeOffset) {
		return &protocol.CompletionList{Items: createTopLevelItems()}, nil
	}

	lines := strings.Split(string(doc.Input()), "\n")
	if len(lines) <= lspLine {
		return nil, nil
	}
	currentLineTrimmed := strings.TrimSpace(lines[lspLine])
	if strings.HasPrefix(currentLineTrimmed, "#") {
		return nil, nil
	}

	whitespaceLine := currentLineTrimmed == ""
	line := int(lspLine) + 1
	character := int(params.Position.Character) + 1
	path := constructCompletionNodePath(file, line)
	if len(path) == 0 {
		return &protocol.CompletionList{Items: createTopLevelItems()}, nil
	} else if len(path) == 1 {
		return nil, nil
	} else if path[1].Key.GetToken().Position.Column >= character {
		return nil, nil
	} else if len(lines[lspLine]) < character-1 {
		return nil, nil
	}

	wordPrefix := prefix(lines[lspLine], character-1)
	path, nodeProps, arrayAttributes := nodeProperties(path, line, character)
	dependencies := dependencyCompletionItems(file, u, path, params, protocol.UInteger(len(wordPrefix)))
	if len(dependencies) > 0 {
		return &protocol.CompletionList{Items: dependencies}, nil
	}
	items, stop := buildTargetCompletionItems(params, manager, path, u, protocol.UInteger(len(wordPrefix)))
	if stop {
		return &protocol.CompletionList{Items: items}, nil
	}

	items = volumeDependencyCompletionItems(file, path, params, protocol.UInteger(len(wordPrefix)), whitespaceLine)
	if len(items) > 0 {
		return &protocol.CompletionList{Items: items}, nil
	}
	items = namedDependencyCompletionItems(file, path, "configs", "configs", params, protocol.UInteger(len(wordPrefix)))
	if len(items) == 0 {
		items = namedDependencyCompletionItems(file, path, "secrets", "secrets", params, protocol.UInteger(len(wordPrefix)))
	}
	isArray := array(lines[lspLine], character-1)
	if isArray != arrayAttributes {
		return nil, nil
	}
	if schema, ok := nodeProps.(*jsonschema.Schema); ok {
		if schema.Enum != nil {
			for _, value := range schema.Enum.Values {
				enumValue := value.(string)
				item := protocol.CompletionItem{
					Label:         enumValue,
					Documentation: schema.Description,
					Detail:        extractDetail(schema),
					TextEdit: protocol.TextEdit{
						NewText: enumValue,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      params.Position.Line,
								Character: params.Position.Character - protocol.UInteger(len(wordPrefix)),
							},
							End: params.Position,
						},
					},
				}
				items = append(items, item)
			}
		}
	} else if properties, ok := nodeProps.(map[string]*jsonschema.Schema); ok {
		sb := strings.Builder{}
		for i := range lines[lspLine] {
			if unicode.IsSpace(rune(lines[lspLine][i])) || lines[lspLine][i] == '-' {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("  ")
		spacing := sb.String()
		for attributeName, schema := range properties {
			item := protocol.CompletionItem{
				Detail: extractDetail(schema),
				Label:  attributeName,
				TextEdit: protocol.TextEdit{
					NewText: insertText(spacing, attributeName, schema),
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character - protocol.UInteger(len(wordPrefix)),
						},
						End: params.Position,
					},
				},
				InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
				InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
			}
			if schema.Description != "" {
				item.Documentation = schema.Description
			} else if schema.Ref != nil && schema.Ref.Description != "" {
				item.Documentation = schema.Ref.Description
			}

			if schema.Enum != nil {
				options := []string{}
				for i := range schema.Enum.Values {
					options = append(options, schema.Enum.Values[i].(string))
				}
				slices.Sort(options)
				sb := strings.Builder{}
				sb.WriteString(attributeName)
				sb.WriteString(": ${1|")
				for i := range options {
					sb.WriteString(options[i])
					if i != len(schema.Enum.Values)-1 {
						sb.WriteString(",")
					}
				}
				sb.WriteString("|}")
				item.TextEdit = protocol.TextEdit{
					NewText: sb.String(),
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      params.Position.Line,
							Character: params.Position.Character - protocol.UInteger(len(wordPrefix)),
						},
						End: params.Position,
					},
				}
			}
			item.TextEdit = modifyTextEdit(file, manager, u, item.TextEdit.(protocol.TextEdit), attributeName, spacing, path)
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, nil
	}
	slices.SortFunc(items, func(a, b protocol.CompletionItem) int {
		return strings.Compare(a.Label, b.Label)
	})
	return &protocol.CompletionList{Items: items}, nil
}

func createChoiceSnippetText(itemTexts []completionItemText) string {
	sb := strings.Builder{}
	sb.WriteString("${1|")
	for i, stage := range itemTexts {
		sb.WriteString(stage.newText)
		if i != len(itemTexts)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString("|}")
	return sb.String()
}

func modifyTextEdit(file *ast.File, manager *document.Manager, u *url.URL, edit protocol.TextEdit, attributeName, spacing string, path []*ast.MappingValueNode) protocol.TextEdit {
	for _, modified := range textEditModifiers {
		if modified.isInterested(attributeName, path) {
			return modified.modify(file, manager, u, edit, attributeName, spacing, path)
		}
	}
	return edit
}

func findDependencies(file *ast.File, dependencyType string) []string {
	services := []string{}
	for _, documentNode := range file.Docs {
		if mappingNode, ok := documentNode.Body.(*ast.MappingNode); ok {
			for _, n := range mappingNode.Values {
				if s, ok := n.Key.(*ast.StringNode); ok {
					if s.Value == dependencyType {
						if mappingNode, ok := n.Value.(*ast.MappingNode); ok {
							for _, service := range mappingNode.Values {
								services = append(services, service.Key.GetToken().Value)
							}
						}
					}
				}
			}
		}
	}
	return services
}

func findBuildStages(manager *document.Manager, dockerfilePath, prefix string) []completionItemText {
	_, nodes := document.OpenDockerfile(context.Background(), manager, dockerfilePath)
	items := []completionItemText{}
	for _, child := range nodes {
		if strings.EqualFold(child.Value, "FROM") {
			if child.Next != nil && child.Next.Next != nil && strings.EqualFold(child.Next.Next.Value, "AS") && child.Next.Next.Next != nil {
				buildStage := child.Next.Next.Next.Value
				if strings.HasPrefix(buildStage, prefix) {
					items = append(items, completionItemText{
						label:         buildStage,
						documentation: child.Next.Value,
						newText:       buildStage,
					})
				}
			}
		}
	}
	return items
}

func buildTargetCompletionItems(params *protocol.CompletionParams, manager *document.Manager, path []*ast.MappingValueNode, u *url.URL, prefixLength protocol.UInteger) ([]protocol.CompletionItem, bool) {
	if len(path) == 4 && path[2].Key.GetToken().Value == "build" && path[3].Key.GetToken().Value == "target" {
		if mappingNode, ok := path[2].Value.(*ast.MappingNode); ok {
			dockerfileAttributePath := "Dockerfile"
			for _, buildAttribute := range mappingNode.Values {
				switch buildAttribute.Key.GetToken().Value {
				case "dockerfile_inline":
					return nil, true
				case "dockerfile":
					dockerfileAttributePath = buildAttribute.Value.GetToken().Value
				}
			}

			dockerfilePath, err := types.AbsolutePath(u, dockerfileAttributePath)
			if err == nil {
				if _, ok := path[3].Value.(*ast.NullNode); ok {
					return createBuildStageItems(params, manager, dockerfilePath, "", prefixLength), true
				} else if prefix, ok := path[3].Value.(*ast.StringNode); ok {
					if int(params.Position.Line) == path[3].Value.GetToken().Position.Line-1 {
						offset := int(params.Position.Character) - path[3].Value.GetToken().Position.Column + 1
						// offset can be greater than the length if there's just empty whitespace after the string value
						if offset <= len(prefix.Value) {
							return createBuildStageItems(params, manager, dockerfilePath, prefix.Value[0:offset], prefixLength), true
						}
					}
				}
			}
		}
	}
	return nil, false
}

func createBuildStageItems(params *protocol.CompletionParams, manager *document.Manager, dockerfilePath, prefix string, prefixLength protocol.UInteger) []protocol.CompletionItem {
	items := []protocol.CompletionItem{}
	for _, itemText := range findBuildStages(manager, dockerfilePath, prefix) {
		items = append(items, protocol.CompletionItem{
			Label:         itemText.label,
			Documentation: itemText.documentation,
			TextEdit: protocol.TextEdit{
				NewText: itemText.newText,
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      params.Position.Line,
						Character: params.Position.Character - prefixLength,
					},
					End: params.Position,
				},
			},
		})
	}
	return items
}

func dependencyCompletionItems(file *ast.File, u *url.URL, path []*ast.MappingValueNode, params *protocol.CompletionParams, prefixLength protocol.UInteger) []protocol.CompletionItem {
	dependency := map[string]string{
		"depends_on": "services",
		"networks":   "networks",
	}
	for serviceAttribute, dependencyType := range dependency {
		items := namedDependencyCompletionItems(file, path, serviceAttribute, dependencyType, params, prefixLength)
		if len(items) > 0 {
			return items
		}
	}
	if len(path) >= 3 && path[2].Key.GetToken().Value == "extends" && path[0].Key.GetToken().Value == "services" {
		if (len(path) == 4 && path[3].Key.GetToken().Value == "service") || params.Position.Line == protocol.UInteger(path[2].Key.GetToken().Position.Line)-1 {
			if !extendingCurrentFile(u, path[2]) {
				return nil
			}

			items := []protocol.CompletionItem{}
			for _, service := range findDependencies(file, "services") {
				if service != path[1].Key.GetToken().Value {
					item := protocol.CompletionItem{
						Label: service,
						TextEdit: protocol.TextEdit{
							NewText: service,
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      params.Position.Line,
									Character: params.Position.Character - prefixLength,
								},
								End: params.Position,
							},
						},
					}
					items = append(items, item)
				}
			}
			return items
		}
	}
	return nil
}

func volumeDependencyCompletionItems(
	file *ast.File,
	path []*ast.MappingValueNode,
	params *protocol.CompletionParams,
	prefixLength protocol.UInteger,
	whitespaceLine bool,
) []protocol.CompletionItem {
	items := namedDependencyCompletionItems(file, path, "volumes", "volumes", params, prefixLength)
	arrayItemPrefix := ""
	if whitespaceLine {
		arrayItemPrefix = "- "
	}
	for i := range items {
		edit := items[i].TextEdit.(protocol.TextEdit)
		items[i].TextEdit = protocol.TextEdit{
			NewText: fmt.Sprintf("%v%v:${1:/container/path}", arrayItemPrefix, edit.NewText),
			Range:   edit.Range,
		}
		items[i].InsertTextFormat = types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet)
	}
	return items
}

func namedDependencyCompletionItems(file *ast.File, path []*ast.MappingValueNode, serviceAttribute, dependencyType string, params *protocol.CompletionParams, prefixLength protocol.UInteger) []protocol.CompletionItem {
	if len(path) == 3 && path[2].Key.GetToken().Value == serviceAttribute {
		items := []protocol.CompletionItem{}
		for _, service := range findDependencies(file, dependencyType) {
			if service != path[1].Key.GetToken().Value {
				item := protocol.CompletionItem{
					Label: service,
					TextEdit: protocol.TextEdit{
						NewText: service,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      params.Position.Line,
								Character: params.Position.Character - prefixLength,
							},
							End: params.Position,
						},
					},
				}
				items = append(items, item)
			}
		}
		return items
	}
	return nil
}

func constructCompletionNodePath(file *ast.File, line int) []*ast.MappingValueNode {
	for i := range len(file.Docs) {
		if i+1 == len(file.Docs) {
			if mappingNode, ok := file.Docs[i].Body.(*ast.MappingNode); ok {
				return NodeStructure(line, mappingNode.Values)
			}
		}

		if m, ok := file.Docs[i].Body.(*ast.MappingNode); ok {
			if n, ok := file.Docs[i+1].Body.(*ast.MappingNode); ok {
				if m.Values[0].Key.GetToken().Position.Line <= line && line <= n.Values[0].Key.GetToken().Position.Line {
					return NodeStructure(line, m.Values)
				}
			}
		}
	}
	return nil
}

func NodeStructure(line int, rootNodes []*ast.MappingValueNode) []*ast.MappingValueNode {
	if len(rootNodes) == 0 {
		return nil
	}

	var candidate *ast.MappingValueNode
	for _, node := range rootNodes {
		if node.GetToken().Position.Line < line {
			candidate = node
		} else if node.GetToken().Position.Line == line {
			return []*ast.MappingValueNode{node}
		} else {
			break
		}
	}
	nodes := []*ast.MappingValueNode{candidate}
	candidates := walkNodes(line, candidate)
	nodes = append(nodes, candidates...)
	return nodes
}

func walkNodes(line int, node *ast.MappingValueNode) []*ast.MappingValueNode {
	var candidate ast.Node
	value := node.Value
	if mappingNode, ok := value.(*ast.MappingNode); ok {
		for _, child := range mappingNode.Values {
			if child.GetToken().Position.Line < line {
				candidate = child
			} else if child.GetToken().Position.Line == line {
				candidate = child
				break
			}
		}
	} else if sequenceNode, ok := value.(*ast.SequenceNode); ok {
		for _, child := range sequenceNode.Values {
			if child.GetToken().Position.Line < line {
				if _, ok := child.(*ast.NullNode); ok {
					continue
				}
				candidate = child
			} else if child.GetToken().Position.Line == line {
				if _, ok := child.(*ast.NullNode); ok {
					break
				}
				candidate = child
				break
			}
		}
	}

	if mappingNode, ok := candidate.(*ast.MappingNode); ok {
		for _, child := range mappingNode.Values {
			if child.GetToken().Position.Line < line {
				candidate = child
			} else if child.GetToken().Position.Line == line {
				candidate = child
				break
			}
		}
	}

	if candidate == nil {
		return []*ast.MappingValueNode{}
	}

	if next, ok := candidate.(*ast.MappingValueNode); ok {
		nodes := []*ast.MappingValueNode{next}
		candidates := walkNodes(line, next)
		nodes = append(nodes, candidates...)
		return nodes
	}
	return []*ast.MappingValueNode{}
}

func referencedTypes(schema *jsonschema.Schema) []string {
	if schema.Types != nil {
		return schema.Types.ToStrings()
	} else if schema.Ref != nil {
		if schema.Ref.Types != nil {
			return schema.Ref.Types.ToStrings()
		}
		schema = schema.Ref
	}
	schemaTypes := []string{}
	for _, referenced := range schema.OneOf {
		if referenced.Types != nil {
			schemaTypes = append(schemaTypes, referenced.Types.ToStrings()[0])
		} else if referenced.Ref != nil {
			schemaTypes = append(schemaTypes, referenced.Ref.Types.ToStrings()[0])
		}
	}
	return schemaTypes
}

func extractDetail(schema *jsonschema.Schema) *string {
	schemaTypes := referencedTypes(schema)
	slices.Sort(schemaTypes)
	return types.CreateStringPointer(strings.Join(schemaTypes, " or "))
}

func requiredFieldsText(spacing string, schema *jsonschema.Schema, schemaTypes []string) []string {
	if len(schemaTypes) == 1 {
		if slices.Contains(schemaTypes, "array") {
			if schema.Ref != nil {
				schema = schema.Ref
			}
			if itemSchema, ok := schema.Items.(*jsonschema.Schema); ok {
				if itemSchema.Ref != nil {
					itemSchema = itemSchema.Ref
				}
				if itemSchema.Types != nil {
					if slices.Contains(itemSchema.Types.ToStrings(), "object") {
						requiredTexts := []string{}
						for _, r := range itemSchema.Required {
							requiredTexts = append(requiredTexts, insertText(fmt.Sprintf("%v  ", spacing), r, itemSchema.Properties[r]))
						}
						return requiredTexts
					}
				}
			}
		}
	}
	return nil
}

func insertText(spacing, attributeName string, schema *jsonschema.Schema) string {
	schemaTypes := referencedTypes(schema)
	if slices.Contains(schemaTypes, "array") {
		if len(schemaTypes) == 1 {
			required := requiredFieldsText(spacing, schema, schemaTypes)
			if len(required) > 0 {
				slices.Sort(required)
				sb := strings.Builder{}
				sb.WriteString(attributeName)
				sb.WriteString(":")
				for i, requiredAttribute := range required {
					sb.WriteString("\n")
					sb.WriteString(spacing)
					if i == 0 {
						sb.WriteString("- ")
					} else {
						sb.WriteString("  ")
					}
					sb.WriteString(requiredAttribute)
					if len(required) != 1 {
						sb.WriteString(fmt.Sprintf("${%v}", i+1))
					}
				}
				return sb.String()
			}
			return fmt.Sprintf("%v:\n%v- ", attributeName, spacing)
		} else if len(schemaTypes) == 2 && slices.Contains(schemaTypes, "object") {
			return fmt.Sprintf("%v:\n%v", attributeName, spacing)
		}
		return fmt.Sprintf("%v:", attributeName)
	}
	if slices.Contains(schemaTypes, "object") {
		if len(schemaTypes) == 1 {
			return fmt.Sprintf("%v:\n%v", attributeName, spacing)
		}
		return fmt.Sprintf("%v:", attributeName)
	}
	if slices.Contains(schemaTypes, "boolean") {
		return fmt.Sprintf("%v: ${1|true,false|}", attributeName)
	}
	return fmt.Sprintf("%v: ", attributeName)
}
