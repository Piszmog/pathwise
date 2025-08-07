package server

import (
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/mcp/server/middleware"
	"github.com/Piszmog/pathwise/mcp/tool"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	srv    *server.MCPServer
	logger *slog.Logger
	addr   string
}

type Option func(*Server)

func New(name string, addr string, logger *slog.Logger, database db.Database, option ...Option) *Server {
	authMiddleware := middleware.AuthMiddleware{Logger: logger, Database: database}

	s := &Server{
		srv: server.NewMCPServer(
			name,
			version.Value,
			server.WithToolCapabilities(true),
			server.WithToolHandlerMiddleware(authMiddleware.Handle),
		),
		addr:   addr,
		logger: logger,
	}
	for _, opt := range option {
		opt(s)
	}
	return s
}

func AddTool(t tool.Tool) Option {
	return func(s *Server) {
		s.srv.AddTool(t.Tool, t.HandlerFunc)
	}
}

func (s *Server) Start() error {
	srv := server.NewStreamableHTTPServer(
		s.srv,
		server.WithLogger(&logger{l: s.logger}),
	)
	return srv.Start(s.addr)
}
