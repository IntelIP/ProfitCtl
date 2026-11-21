package mcpserver

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPHandler exposes the read-only scenario operations over the stateless
// Streamable HTTP transport. Callers must bind it only to a trusted local
// listener or add their own authentication and authorization boundary.
func (s *Server) HTTPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return s.mcpServer(false)
		},
		&mcp.StreamableHTTPOptions{
			Stateless:                    true,
			JSONResponse:                 true,
			MaxRequestBodyBytes:          64 << 10,
			PropagateRequestCancellation: true,
		},
	)
}
