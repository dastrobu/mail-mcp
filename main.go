package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/dastrobu/apple-mail-mcp/internal/opts"
	"github.com/dastrobu/apple-mail-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "apple-mail"
	serverVersion = "0.1.0"
)

func main() {
	// Parse command-line options
	options, err := opts.Parse()
	if err != nil {
		log.Fatalf("Failed to parse options: %v", err)
	}

	if err := run(options); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// debugMiddleware logs all MCP requests and responses when debug is enabled
func debugMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Log the request
		if req != nil {
			p := req.GetParams()
			j, _ := json.MarshalIndent(p, "", "  ")
			log.Printf("[DEBUG] MCP Request: %s\nParams: %s\n", method, string(j))
		} else {
			log.Printf("[DEBUG] MCP Request: %s\n", method)
		}

		// Call the next handler
		result, err := next(ctx, method, req)

		// Log the response
		if err != nil {
			log.Printf("[DEBUG] MCP Response: %s\nError: %v\n", method, err)
		} else if result != nil {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			log.Printf("[DEBUG] MCP Response: %s\nResult: %s\n", method, string(resultJSON))
		} else {
			log.Printf("[DEBUG] MCP Response: %s\n", method)
		}

		return result, err
	}
}

// createServer creates and configures a new MCP server instance
func createServer(options *opts.Options) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	// Add debug middleware if debug mode is enabled
	if options.Debug {
		srv.AddReceivingMiddleware(debugMiddleware)
	}

	// Register all tools
	tools.RegisterAll(srv)

	return srv
}

func run(options *opts.Options) error {
	ctx := context.Background()

	// Run startup check to verify Mail.app is accessible
	log.Println("Running Mail.app connectivity check...")
	startupData, err := jxa.StartupCheck(ctx)
	if err != nil {
		return fmt.Errorf(`Mail.app connectivity check failed: %w

This usually means either:
1. Mail.app is not running - Please start Mail.app
2. Missing automation permissions - Grant permission in System Settings > Privacy & Security > Automation

For detailed troubleshooting, see: https://github.com/dastrobu/apple-mail-mcp#troubleshooting`, err)
	}
	log.Println("Mail.app is accessible and ready")

	// Print Mail.app properties as JSON when debugging is enabled
	if options.Debug {
		if properties, ok := startupData["properties"].(map[string]any); ok {
			propertiesJSON, err := json.MarshalIndent(properties, "", "  ")
			if err != nil {
				log.Printf("[DEBUG] Failed to marshal properties: %v\n", err)
			} else {
				log.Printf("[DEBUG] Mail.app Properties:\n%s\n", string(propertiesJSON))
			}
		}
	}

	// Log to stderr (stdout is used for MCP communication in stdio mode)
	log.Printf("Apple Mail MCP Server v%s initialized\n", serverVersion)

	srv := createServer(options)

	// Run the server with the selected transport
	switch options.Transport {
	case "stdio":
		log.Println("Using STDIO transport")
		if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
			return err
		}
	case "http":
		addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
		log.Printf("Starting HTTP server on http://%s\n", addr)

		handler := mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server {
				// since we are stateless, we can return the same server instance
				return srv
			},
			&mcp.StreamableHTTPOptions{
				Stateless: true,
			},
		)

		// Create HTTP server
		httpServer := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		// Run the HTTP server
		log.Printf("HTTP server listening on http://%s\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
	default:
		return fmt.Errorf("unsupported transport: %s", options.Transport)
	}

	return nil
}
