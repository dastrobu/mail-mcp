package opts

import (
	"fmt"
	"os"

	"github.com/dastrobu/apple-mail-mcp/internal/launchd"
	"github.com/dastrobu/apple-mail-mcp/internal/opts/typed_flags"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
)

// Options defines the command-line options for the MCP server
type Options struct {
	Transport      typed_flags.Transport `long:"transport" env:"APPLE_MAIL_MCP_TRANSPORT" description:"Transport type: stdio or http" default:"stdio"`
	Port           int                   `long:"port" env:"APPLE_MAIL_MCP_PORT" description:"HTTP port (only used with --transport=http)" default:"8787"`
	Host           string                `long:"host" env:"APPLE_MAIL_MCP_HOST" description:"HTTP host (only used with --transport=http)" default:"localhost"`
	Debug          bool                  `long:"debug" env:"APPLE_MAIL_MCP_DEBUG" description:"Enable debug logging of tool calls and results to stderr"`
	RichTextStyles string                `long:"rich-text-styles" env:"APPLE_MAIL_MCP_RICH_TEXT_STYLES" description:"Path to custom rich text styles YAML file (uses embedded default if not specified)"`

	Launchd    LaunchdCmd    `command:"launchd" description:"Manage launchd service"`
	Completion CompletionCmd `command:"completion" description:"Generate completion scripts"`
}

// CompletionCmd holds completion subcommands
type CompletionCmd struct {
	Bash CompletionBashCmd `command:"bash" description:"Generate bash completion script"`
}

// CompletionBashCmd represents the 'completion bash' command
type CompletionBashCmd struct {
	Handler func() error
}

// Execute runs the completion bash command
func (c *CompletionBashCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

// LaunchdCmd holds launchd subcommands
type LaunchdCmd struct {
	Create LaunchdCreateCmd `command:"create" description:"Set up launchd service for automatic startup"`
	Remove LaunchdRemoveCmd `command:"remove" description:"Remove launchd service"`
}

// LaunchdCreateCmd represents the 'launchd create' command
type LaunchdCreateCmd struct {
	Handler func() error
}

// Execute runs the launchd create command
func (c *LaunchdCreateCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

// LaunchdRemoveCmd represents the 'launchd remove' command
type LaunchdRemoveCmd struct {
	Handler func() error
}

// Execute runs the launchd remove command
func (c *LaunchdRemoveCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

var GlobalOpts = Options{}

// Parse parses command-line arguments and environment variables
// It also loads .env file if present (but doesn't fail if missing)
func Parse() (*flags.Parser, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	// This allows local development with .env files while working in production with env vars
	_ = godotenv.Load()

	// Set defaults from launchd constants
	if GlobalOpts.Port == 0 {
		GlobalOpts.Port = launchd.DefaultPort
	}
	if GlobalOpts.Host == "" {
		GlobalOpts.Host = launchd.DefaultHost
	}

	parser := flags.NewParser(&GlobalOpts, flags.HelpFlag)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			switch flagsErr.Type {
			case flags.ErrHelp:
				// Help was displayed, exit cleanly
				os.Exit(0)
			case flags.ErrCommandRequired:
				// No command specified - that's OK, we'll run the server
				return parser, nil
			default:
				return nil, fmt.Errorf("failed to parse options: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to parse options: %w", err)
	}

	return parser, nil
}
