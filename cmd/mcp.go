package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/mcpserver"
	"github.com/spf13/cobra"
)

var (
	mcpWorkspaceRoot   string
	mcpStandardsBinary string
	mcpListenAddress   string
)

var mcpCmd = &cobra.Command{
	Use:          "mcp",
	Short:        "Run the bounded local ProfitCtl MCP server",
	Hidden:       true,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE:         runMCP,
}

var mcpServeCmd = &cobra.Command{
	Use:          "serve",
	Short:        "Serve the bounded local ProfitCtl MCP endpoint over Streamable HTTP",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE:         runMCPHTTP,
}

func init() {
	mcpCmd.PersistentFlags().StringVar(&mcpWorkspaceRoot, "workspace-root", "", "Fixed local workspace root")
	mcpCmd.PersistentFlags().StringVar(&mcpStandardsBinary, "standards-binary", "", "Path to local profitctl-standards executable")
	mcpServeCmd.Flags().StringVar(&mcpListenAddress, "listen", "127.0.0.1:8173", "Loopback address for the local Streamable HTTP endpoint")
	mcpCmd.AddCommand(mcpServeCmd)
}

func runMCP(_ *cobra.Command, _ []string) error {
	server, err := newMCPServer()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return server.Run(ctx)
}

func runMCPHTTP(_ *cobra.Command, _ []string) error {
	if err := validateLoopbackAddress(mcpListenAddress); err != nil {
		return err
	}
	server, err := newMCPServer()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", mcpListenAddress)
	if err != nil {
		return fmt.Errorf("listen for local ProfitCtl MCP: %w", err)
	}
	defer listener.Close()

	httpServer := &http.Server{
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	mux := httpServer.Handler.(*http.ServeMux)
	mux.Handle("/mcp", server.HTTPHandler())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	fmt.Fprintf(os.Stderr, "ProfitCtl MCP listening at http://%s/mcp\n", listener.Addr())
	err = httpServer.Serve(listener)
	if err == nil || errors.Is(err, http.ErrServerClosed) || ctx.Err() != nil {
		return nil
	}
	return fmt.Errorf("serve local ProfitCtl MCP: %w", err)
}

func newMCPServer() (*mcpserver.Server, error) {
	workspaceRoot := mcpWorkspaceRoot
	if workspaceRoot == "" {
		workspaceRoot = os.Getenv("PROFITCTL_WORKSPACE_ROOT")
	}
	if workspaceRoot == "" {
		return nil, fmt.Errorf("profitctl mcp requires --workspace-root or PROFITCTL_WORKSPACE_ROOT")
	}

	return mcpserver.New(mcpserver.Config{
		WorkspaceRoot:   workspaceRoot,
		StandardsBinary: mcpStandardsBinary,
	})
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("local ProfitCtl MCP listen address must include host and port: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("local ProfitCtl MCP listen address must use a loopback IP")
	}
	return nil
}
