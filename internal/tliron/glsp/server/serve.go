package server

import (
	"io"

	"github.com/tliron/commonlog"
)

// See: https://github.com/sourcegraph/go-langserver/blob/master/main.go#L179

func (self *Server) ServeStream(stream io.ReadWriteCloser, log commonlog.Logger) {
	if log == nil {
		log = self.Log
	}
	log.Info("new stream connection")
	<-self.newStreamConnection(stream).DisconnectNotify()
	log.Info("stream connection closed")
}
