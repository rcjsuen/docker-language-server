package server

import (
	contextpkg "context"
	"errors"
	"fmt"

	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/sourcegraph/jsonrpc2"
)

// See: https://github.com/sourcegraph/go-langserver/blob/master/langserver/handler.go#L206

func (self *Server) newHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(self.handle)
}

func (self *Server) handle(context contextpkg.Context, connection *jsonrpc2.Conn, request *jsonrpc2.Request) (any, error) {
	glspContext := glsp.Context{
		Method: request.Method,
		Notify: func(ctx contextpkg.Context, method string, params any) {
			if err := connection.Notify(ctx, method, params); err != nil {
				self.Log.Error(err.Error())
			}
		},
		Call: func(ctx contextpkg.Context, method string, params any, result any) {
			if err := connection.Call(ctx, method, params, result); err != nil {
				self.Log.Error(err.Error())
			}
		},
		Context: context,
	}

	if request.Params != nil {
		glspContext.Params = *request.Params
	}

	switch request.Method {
	case "exit":
		// We're giving the attached handler a chance to handle it first, but we'll ignore any result
		self.Handler.Handle(&glspContext)
		err := connection.Close()
		return nil, err

	default:
		// Note: jsonrpc2 will not even call this function if reqest.Params is invalid JSON,
		// so we don't need to handle jsonrpc2.CodeParseError here
		result, validMethod, validParams, err := self.Handler.Handle(&glspContext)
		if !validMethod {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: fmt.Sprintf("method not supported: %s", request.Method),
			}
		} else if !validParams {
			if err == nil {
				return nil, &jsonrpc2.Error{
					Code: jsonrpc2.CodeInvalidParams,
				}
			} else {
				return nil, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: err.Error(),
				}
			}
		} else if err != nil {
			var jsonsrpcErr *jsonrpc2.Error
			if errors.As(err, &jsonsrpcErr) {
				return nil, jsonsrpcErr
			}

			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: err.Error(),
			}
		} else {
			return result, nil
		}
	}
}
