package opts

import (
	"fmt"
	"os"

	"github.com/dastrobu/apple-mail-mcp/internal/opts/typed_flags"
	"github.com/dastrobu/apple-mail-mcp/internal/tools"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
)

// Options defines the command-line options for the MCP server
type Options struct {
	Version bool `long:"version" short:"v" description:"Show version information and exit"`

	Run        RunCmd        `command:"run" description:"Run the server"`
	Launchd    LaunchdCmd    `command:"launchd" description:"Manage launchd service"`
	Completion CompletionCmd `command:"completion" description:"Generate completion scripts"`
	Tool       ToolCmd       `command:"tool" description:"Execute a tool directly"`
}

// RunCmd defines the 'run' command
type RunCmd struct {
	Transport      typed_flags.Transport `long:"transport" env:"APPLE_MAIL_MCP_TRANSPORT" description:"Transport type: stdio or http" default:"stdio"`
	Port           int                   `long:"port" env:"APPLE_MAIL_MCP_PORT" description:"HTTP port (only used with --transport=http)" default:"8787"`
	Host           string                `long:"host" env:"APPLE_MAIL_MCP_HOST" description:"HTTP host (only used with --transport=http)" default:"localhost"`
	Debug          bool                  `long:"debug" env:"APPLE_MAIL_MCP_DEBUG" description:"Enable debug logging of tool calls and results to stderr"`
	RichTextStyles string                `long:"rich-text-styles" env:"APPLE_MAIL_MCP_RICH_TEXT_STYLES" description:"Path to custom rich text styles YAML file (uses embedded default if not specified)"`

	Handler func() error
}

// Execute runs the run command
func (c *RunCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
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
	Create  LaunchdCreateCmd  `command:"create" description:"Set up launchd service for automatic startup"`
	Remove  LaunchdRemoveCmd  `command:"remove" description:"Remove launchd service"`
	Restart LaunchdRestartCmd `command:"restart" description:"Restart the launchd service"`
}

// LaunchdCreateCmd represents the 'launchd create' command
type LaunchdCreateCmd struct {
	DisableRunAtLoad bool `long:"disable-run-at-load" description:"Disable automatic startup on login (service must be started manually)"`

	// Configuration for the service
	Port  int    `long:"port" description:"HTTP port for the service" default:"8787"`
	Host  string `long:"host" description:"HTTP host for the service" default:"localhost"`
	Debug bool   `long:"debug" description:"Enable debug logging for the service"`

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

// LaunchdRestartCmd defines the 'restart' subcommand for launchd
type LaunchdRestartCmd struct {
	Handler func() error
}

// Execute runs the launchd remove command
func (c *LaunchdRemoveCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

// Execute runs the launchd restart command
func (c *LaunchdRestartCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

// ToolCmd holds tool subcommands
type ToolCmd struct {
	ListAccounts           ListAccountsCmd           `command:"list_accounts" description:"Lists all configured email accounts"`
	ListMailboxes          ListMailboxesCmd          `command:"list_mailboxes" description:"Lists mailboxes for a specific account"`
	GetMessageContent      GetMessageContentCmd      `command:"get_message_content" description:"Retrieves the full content of a specific message"`
	GetSelectedMessages    GetSelectedMessagesCmd    `command:"get_selected_messages" description:"Gets the currently selected message(s)"`
	CreateReplyDraft       CreateReplyDraftCmd       `command:"create_reply_draft" description:"Creates a reply to a specific message"`
	ReplaceReplyDraft      ReplaceReplyDraftCmd      `command:"replace_reply_draft" description:"Replaces an existing reply draft"`
	ListDrafts             ListDraftsCmd             `command:"list_drafts" description:"Lists draft messages from the Drafts mailbox"`
	CreateOutgoingMessage  CreateOutgoingMessageCmd  `command:"create_outgoing_message" description:"Creates a new outgoing email message"`
	ListOutgoingMessages   ListOutgoingMessagesCmd   `command:"list_outgoing_messages" description:"Lists all OutgoingMessage objects currently in memory"`
	ReplaceOutgoingMessage ReplaceOutgoingMessageCmd `command:"replace_outgoing_message" description:"Replaces an existing outgoing message"`
	FindMessages           FindMessagesCmd           `command:"find_messages" description:"Find messages in a mailbox"`
}

// ListAccountsCmd represents the 'tool list_accounts' command
type ListAccountsCmd struct {
	tools.ListAccountsInput
	Handler func(tools.ListAccountsInput) error
}

// Execute runs the list_accounts tool command
func (c *ListAccountsCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.ListAccountsInput)
	}
	return nil
}

// ListMailboxesCmd represents the 'tool list_mailboxes' command
type ListMailboxesCmd struct {
	tools.ListMailboxesInput
	Handler func(tools.ListMailboxesInput) error
}

// Execute runs the list_mailboxes tool command
func (c *ListMailboxesCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.ListMailboxesInput)
	}
	return nil
}

// GetMessageContentCmd represents the 'tool get_message_content' command
type GetMessageContentCmd struct {
	tools.GetMessageContentInput
	Handler func(tools.GetMessageContentInput) error
}

// Execute runs the get_message_content tool command
func (c *GetMessageContentCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.GetMessageContentInput)
	}
	return nil
}

// GetSelectedMessagesCmd represents the 'tool get_selected_messages' command
type GetSelectedMessagesCmd struct {
	tools.GetSelectedMessagesInput
	Handler func(tools.GetSelectedMessagesInput) error
}

// Execute runs the get_selected_messages tool command
func (c *GetSelectedMessagesCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.GetSelectedMessagesInput)
	}
	return nil
}

// CreateReplyDraftCmd represents the 'tool create_reply_draft' command
type CreateReplyDraftCmd struct {
	tools.CreateReplyDraftInput
	Handler func(tools.CreateReplyDraftInput) error
}

// Execute runs the create_reply_draft tool command
func (c *CreateReplyDraftCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.CreateReplyDraftInput)
	}
	return nil
}

// ReplaceReplyDraftCmd represents the 'tool replace_reply_draft' command
type ReplaceReplyDraftCmd struct {
	tools.ReplaceReplyDraftInput
	Handler func(tools.ReplaceReplyDraftInput) error
}

// Execute runs the replace_reply_draft tool command
func (c *ReplaceReplyDraftCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.ReplaceReplyDraftInput)
	}
	return nil
}

// ListDraftsCmd represents the 'tool list_drafts' command
type ListDraftsCmd struct {
	tools.ListDraftsInput
	Handler func(tools.ListDraftsInput) error
}

// Execute runs the list_drafts tool command
func (c *ListDraftsCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.ListDraftsInput)
	}
	return nil
}

// CreateOutgoingMessageCmd represents the 'tool create_outgoing_message' command
type CreateOutgoingMessageCmd struct {
	tools.CreateOutgoingMessageInput
	Handler func(tools.CreateOutgoingMessageInput) error
}

// Execute runs the create_outgoing_message tool command
func (c *CreateOutgoingMessageCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.CreateOutgoingMessageInput)
	}
	return nil
}

// ListOutgoingMessagesCmd represents the 'tool list_outgoing_messages' command
type ListOutgoingMessagesCmd struct {
	Handler func() error
}

// Execute runs the list_outgoing_messages tool command
func (c *ListOutgoingMessagesCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler()
	}
	return nil
}

// ReplaceOutgoingMessageCmd represents the 'tool replace_outgoing_message' command
type ReplaceOutgoingMessageCmd struct {
	tools.ReplaceOutgoingMessageInput
	Handler func(tools.ReplaceOutgoingMessageInput) error
}

// Execute runs the replace_outgoing_message tool command
func (c *ReplaceOutgoingMessageCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.ReplaceOutgoingMessageInput)
	}
	return nil
}

// FindMessagesCmd represents the 'tool find_messages' command
type FindMessagesCmd struct {
	tools.FindMessagesInput
	Handler func(tools.FindMessagesInput) error
}

// Execute runs the find_messages tool command
func (c *FindMessagesCmd) Execute(args []string) error {
	if c.Handler != nil {
		return c.Handler(c.FindMessagesInput)
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

	// Defaults are now handled by struct tags in subcommands or local to handlers.

	parser := flags.NewParser(&GlobalOpts, flags.HelpFlag|flags.PassDoubleDash)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			switch flagsErr.Type {
			case flags.ErrHelp:
				// Print help message
				parser.WriteHelp(os.Stdout)
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
