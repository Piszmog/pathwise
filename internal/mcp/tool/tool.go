package tool

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Tool struct {
	mcp.Tool

	HandlerFunc server.ToolHandlerFunc
}
