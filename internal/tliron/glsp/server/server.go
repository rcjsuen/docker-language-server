package server

import (
	"time"

	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/tliron/commonlog"
)

var DefaultTimeout = time.Minute

//
// Server
//

type Server struct {
	Handler     glsp.Handler
	LogBaseName string
	Debug       bool

	Log           commonlog.Logger
	Timeout       time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	StreamTimeout time.Duration
}

func NewServer(handler glsp.Handler, logName string, debug bool) *Server {
	return &Server{
		Handler:       handler,
		LogBaseName:   logName,
		Debug:         debug,
		Log:           commonlog.GetLogger(logName),
		Timeout:       DefaultTimeout,
		ReadTimeout:   DefaultTimeout,
		WriteTimeout:  DefaultTimeout,
		StreamTimeout: DefaultTimeout,
	}
}
