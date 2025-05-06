package compose

import (
	"strings"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/sourcegraph/jsonrpc2"
)

type indentation struct {
	original int
	desired  int
}

type comment struct {
	line       int
	whitespace int
}

func formattingOptionTabSize(options protocol.FormattingOptions) (int, error) {
	if tabSize, ok := options[protocol.FormattingOptionTabSize].(float64); ok && tabSize > 0 {
		return int(tabSize), nil
	}
	return -1, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "tabSize is not a positive integer"}
}

func indent(indentation int) string {
	sb := strings.Builder{}
	for range indentation {
		sb.WriteString(" ")
	}
	return sb.String()
}

func Formatting(doc document.ComposeDocument, options protocol.FormattingOptions) ([]protocol.TextEdit, error) {
	file := doc.File()
	if file == nil || len(file.Docs) == 0 {
		return nil, nil
	}
	tabSize, err := formattingOptionTabSize(options)
	if err != nil {
		return nil, err
	}

	edits := []protocol.TextEdit{}
	indentations := []indentation{}
	comments := []comment{}
	topLevelNodeDetected := false
	lines := strings.Split(string(doc.Input()), "\n")
lineCheck:
	for lineNumber, line := range lines {
		lineIndentation := 0
		stop := 0
		isComment := false
		empty := true
		for i := range len(line) {
			if line[i] == 32 {
				lineIndentation++
			} else if line[i] == '#' {
				empty = false
				isComment = true
				comments = append(comments, comment{line: lineNumber, whitespace: i})
				break
			} else {
				empty = false
				if strings.HasPrefix(lines[lineNumber], "---") {
					edits = append(edits, formatComments(comments, 0)...)
					comments = nil
					indentations = nil
					topLevelNodeDetected = false
					continue lineCheck
				}

				if !topLevelNodeDetected {
					topLevelNodeDetected = true
					if lineIndentation > 0 {
						newIndentation, _ := updateIndentation(indentations, lineIndentation, 0)
						indentations = append(indentations, newIndentation)
					}
				}
				break
			}
			stop++
		}

		if isComment {
			continue
		}

		if lineIndentation != 0 {
			newIndentation, resetIndex := updateIndentation(indentations, lineIndentation, tabSize)
			if resetIndex == -1 {
				indentations = append(indentations, newIndentation)
			} else {
				indentations = indentations[:resetIndex+1]
			}
			edits = append(edits, formatComments(comments, newIndentation.desired)...)
			comments = nil
			if lineIndentation != newIndentation.desired {
				edits = append(edits, protocol.TextEdit{
					NewText: indent(newIndentation.desired),
					Range: protocol.Range{
						Start: protocol.Position{Line: protocol.UInteger(lineNumber), Character: 0},
						End:   protocol.Position{Line: protocol.UInteger(lineNumber), Character: protocol.UInteger(stop)},
					},
				})
			}
		} else if !empty {
			edits = append(edits, formatComments(comments, 0)...)
			comments = nil
			indentations = nil
		}
	}
	return edits, nil
}

// formatComments goes over the list of comments and corrects its
// indentation to the desired indentation only if it differs. Any
// comment that needs to have its indentation changed will have a
// TextEdit created for it and included in the returned result.
func formatComments(comments []comment, desired int) []protocol.TextEdit {
	edits := []protocol.TextEdit{}
	for _, c := range comments {
		if desired != c.whitespace {
			edits = append(edits, protocol.TextEdit{
				NewText: indent(desired),
				Range: protocol.Range{
					Start: protocol.Position{Line: protocol.UInteger(c.line), Character: 0},
					End:   protocol.Position{Line: protocol.UInteger(c.line), Character: protocol.UInteger(c.whitespace)},
				},
			})
		}
	}
	return edits
}

func updateIndentation(indentations []indentation, original, tabSpacing int) (indentation, int) {
	last := tabSpacing
	for i := range indentations {
		if indentations[i].original == original {
			return indentations[i], i
		}
		last = indentations[i].desired + tabSpacing
	}
	return indentation{
		original: original,
		desired:  last,
	}, -1
}
