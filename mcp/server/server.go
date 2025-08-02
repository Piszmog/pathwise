package server

import (
	"log/slog"

	"github.com/Piszmog/pathwise/internal/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	srv    *server.MCPServer
	logger *slog.Logger
	addr   string
}

type Option func(*Server)

func New(name string, addr string, logger *slog.Logger, option ...Option) *Server {
	s := &Server{
		srv: server.NewMCPServer(
			name,
			version.Value,
			server.WithToolCapabilities(true),
		),
		addr:   addr,
		logger: logger,
	}
	for _, opt := range option {
		opt(s)
	}
	return s
}

func AddTool(name string, description string, handler server.ToolHandlerFunc, options ...mcp.ToolOption) Option {
	opts := make([]mcp.ToolOption, 0, len(options)+1)
	opts = append(opts, mcp.WithDescription(description))
	opts = append(opts, options...)
	return func(s *Server) {
		s.srv.AddTool(
			mcp.NewTool(
				name,
				opts...,
			),
			handler,
		)
	}
}

func (s *Server) Start() error {
	srv := server.NewStreamableHTTPServer(
		s.srv,
		server.WithLogger(&logger{l: s.logger}),
	)
	return srv.Start(s.addr)
}
