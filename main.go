package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dastrobu/apple-mail-mcp/internal/completion"
	"github.com/dastrobu/apple-mail-mcp/internal/launchd"
	applog "github.com/dastrobu/apple-mail-mcp/internal/log"
	"github.com/dastrobu/apple-mail-mcp/internal/opts"
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/dastrobu/apple-mail-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// Version information (set by GoReleaser via ldflags)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	serverName = "apple-mail"
)

func main() {
	// Set up command handlers before parsing
	opts.GlobalOpts.Completion.Bash.Handler = func() error {
		completion.GenerateBash()
		return nil
	}
	opts.GlobalOpts.Launchd.Create.Handler = func() error {
		return createLaunchd(&opts.GlobalOpts)
	}
	opts.GlobalOpts.Launchd.Remove.Handler = func() error {
		return removeLaunchd()
	}

	// Parse command-line options
	parser, err := opts.Parse()
	if err != nil {
		log.Fatalf("Failed to parse options: %v", err)
	}

	// Handle version flag
	if opts.GlobalOpts.Version {
		fmt.Printf("apple-mail-mcp version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		return
	}

	// Check if a command was executed
	if parser.Command.Active != nil {
		// Command was executed via Execute() method
		return
	}

	// No command specified - run the server
	if err := run(&opts.GlobalOpts); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// setupLogger creates and adds the appropriate logger to the context
func setupLogger(ctx context.Context, debug bool) context.Context {
	if debug {
		return applog.WithLogger(ctx, log.Default())
	}
	return applog.WithLogger(ctx, log.New(io.Discard, "", 0))
}

// debugMiddleware logs all MCP requests and responses when debug is enabled
func debugMiddleware(debug bool) func(mcp.MethodHandler) mcp.MethodHandler {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Add logger to context for this request
			ctx = setupLogger(ctx, debug)

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
}

// createServer creates and configures a new MCP server instance
func createServer(options *opts.Options, richtextConfig *richtext.PreparedConfig) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: version,
	}, nil)

	// Add debug middleware if debug mode is enabled
	if options.Debug {
		srv.AddReceivingMiddleware(debugMiddleware(options.Debug))
	}

	// Register all tools
	tools.RegisterAll(srv, richtextConfig)

	return srv
}

func run(options *opts.Options) error {
	// Convert Transport to string for comparison
	transport := string(options.Transport)

	ctx := context.Background()

	// Always add a logger to context (real logger if debug, no-op otherwise)
	ctx = setupLogger(ctx, options.Debug)

	// Load rich text rendering configuration
	richtextConfig, err := richtext.LoadConfig(options.RichTextStyles)
	if err != nil {
		return fmt.Errorf("failed to load rich text styles configuration: %w", err)
	}
	if options.Debug {
		log.Printf("[DEBUG] Rich text styles loaded from: %s\n",
			func() string {
				if options.RichTextStyles == "" {
					return "embedded default"
				}
				return options.RichTextStyles
			}())
	}

	// Note: We don't check Mail.app connectivity at startup because:
	// 1. Mail.app may not be running yet (e.g., launchd starts before user opens Mail)
	// 2. Each tool call will detect and report Mail.app availability gracefully
	// 3. This allows the server to start without requiring Mail.app to be running

	// Log to stderr (stdout is used for MCP communication in stdio mode)
	log.Printf("Apple Mail MCP Server v%s (commit: %s, built: %s) initialized\n", version, commit, date)

	srv := createServer(options, richtextConfig)

	// Run the server with the selected transport
	switch transport {
	case "stdio":
		log.Println("Using STDIO transport")
		log.Println("⚠️  WARNING: STDIO transport requires high permissions and grants automation access to the parent process (Terminal, Claude Desktop, etc.)")
		log.Println("⚠️  It is strongly recommended to use launchd instead: apple-mail-mcp launchd create")
		log.Printf("⚠️  If STDIO is required for testing, consider running 'tccutil reset AppleEvents %s' afterwards\n", os.Args[0])
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
		return fmt.Errorf("unsupported transport: %s", transport)
	}

	return nil
}

// createLaunchd creates the launchd service
func createLaunchd(options *opts.Options) error {
	cfg, err := launchd.DefaultConfig()
	if err != nil {
		return err
	}

	// Override defaults with command-line options if provided
	if options.Host != launchd.DefaultHost {
		cfg.Host = options.Host
	}
	if options.Port != launchd.DefaultPort {
		cfg.Port = options.Port
	}
	if options.Debug {
		cfg.Debug = options.Debug
	}

	return launchd.Create(cfg)
}

// removeLaunchd removes the launchd service
func removeLaunchd() error {
	return launchd.Remove()
}
