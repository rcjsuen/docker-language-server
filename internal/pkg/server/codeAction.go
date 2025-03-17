package server

import (
	"encoding/json"

	"github.com/docker/docker-language-server/internal/telemetry"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
)

func (s *Server) TextDocumentCodeAction(ctx *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	actions := []protocol.CodeAction{}
	for _, diagnostic := range params.Context.Diagnostics {
		bytes, _ := json.Marshal(diagnostic.Data)
		edits := []*types.NamedEdit{}
		_ = json.Unmarshal(bytes, &edits)
		if len(edits) > 0 {
			for _, edit := range edits {
				editRange := protocol.Range{
					Start: protocol.Position{
						Line:      diagnostic.Range.Start.Line,
						Character: diagnostic.Range.Start.Character,
					},
					End: protocol.Position{
						Line:      diagnostic.Range.End.Line,
						Character: diagnostic.Range.End.Character,
					},
				}
				if edit.Range != nil {
					editRange = *edit.Range
				}
				action := protocol.CodeAction{
					Title: edit.Title,
					Edit: &protocol.WorkspaceEdit{
						Changes: map[string][]protocol.TextEdit{
							params.TextDocument.URI: {
								protocol.TextEdit{
									NewText: edit.Edit,
									Range:   editRange,
								},
							},
						},
					},
					Command: &protocol.Command{
						Command: types.TelemetryCallbackCommandId,
						Arguments: []any{
							telemetry.EventServerUserAction,
							map[string]any{
								"action":     telemetry.ServerUserActionTypeCommandExecuted,
								"action_id":  types.CodeActionDiagnosticCommandId,
								"diagnostic": diagnostic.Code,
							},
						},
					},
				}
				actions = append(actions, action)
			}
		}
	}

	return actions, nil
}
