package server

import (
	"context"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

// LanguageClient exposes JSON-RPC notifications and requests that are
// sent from the server to the client. Notifications can be sent freely
// but requests should be done in its own goroutine. Requests must not
// be sent from within an existing JSON-RPC notification or request
// handler. This is because the handling code is single-threaded and is
// expected to return to handle future messages. If a request is sent
// then the handler would never be able to process the response.
type LanguageClient struct {
	call   glsp.CallFunc
	notify glsp.NotifyFunc
}

func (c *LanguageClient) ShowMessage(ctx context.Context, params protocol.ShowMessageParams) {
	c.notify(ctx, protocol.ServerWindowShowMessage, params)
}

func (c *LanguageClient) ShowDocumentRequest(ctx context.Context, params protocol.ShowDocumentParams, result *protocol.ShowDocumentResult) {
	c.call(ctx, protocol.ServerWindowShowDocument, params, result)
}

func (c *LanguageClient) ShowMessageRequest(ctx context.Context, params protocol.ShowMessageRequestParams, result *protocol.MessageActionItem) {
	c.call(ctx, protocol.ServerWindowShowMessageRequest, params, &result)
}

func (c *LanguageClient) WorkspaceCodeLensRefresh(ctx context.Context) {
	c.call(ctx, protocol.ServerWorkspaceCodeLensRefresh, nil, nil)
}

func (c *LanguageClient) WorkspaceSemanticTokensRefresh(ctx context.Context) {
	c.call(ctx, protocol.MethodWorkspaceSemanticTokensRefresh, nil, nil)
}

func (c *LanguageClient) PublishDiagnostics(ctx context.Context, params protocol.PublishDiagnosticsParams) {
	c.notify(ctx, protocol.ServerTextDocumentPublishDiagnostics, params)
}

func (c *LanguageClient) WorkspaceConfiguration(ctx context.Context, params protocol.ConfigurationParams, configurations *[]configuration.Configuration) {
	c.call(ctx, protocol.ServerWorkspaceConfiguration, params, &configurations)
}

// RegisterCapability sends the client/registerCapability request from
// the server to the client to dynamically register capabilities.
func (c *LanguageClient) RegisterCapability(ctx context.Context, params protocol.RegistrationParams) {
	c.call(ctx, protocol.ServerClientRegisterCapability, params, nil)
}
