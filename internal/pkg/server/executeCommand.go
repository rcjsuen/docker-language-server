package server

import (
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
)

func (s *Server) WorkspaceExecuteCommand(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
	if params.Command == types.TelemetryCallbackCommandId && len(params.Arguments) == 2 {
		if event, ok := params.Arguments[0].(string); ok {
			if properties, ok := params.Arguments[1].(map[string]any); ok {
				s.Enqueue(event, properties)
			}
		}
	}
	return nil, nil
}
