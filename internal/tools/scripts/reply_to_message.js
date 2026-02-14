function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  // Check if Mail.app is running
  if (!Mail.running()) {
    return JSON.stringify({
      success: false,
      error: "Mail.app is not running. Please start Mail.app and try again.",
      errorCode: "MAIL_APP_NOT_RUNNING",
    });
  }

  // Collect logs instead of using console.log
  const logs = [];

  // Helper function to log messages
  function log(message) {
    logs.push(message);
  }

  // Parse arguments: accountName, mailboxPathStr (JSON), messageId, replyContent, contentFormat, contentJson, openingWindow, replyToAll
  const accountName = argv[0] || "";
  const mailboxPathStr = argv[1] || "";
  const messageId = argv[2] ? parseInt(argv[2]) : 0;
  const replyContent = argv[3] || "";
  const contentFormat = argv[4] || "plain";
  const contentJson = argv[5] || "";
  const openingWindow = argv[6] === "true";
  const replyToAll = argv[7] === "true";

  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
    });
  }

  if (!mailboxPathStr) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path is required",
    });
  }

  // Parse mailboxPath from JSON
  let mailboxPath;
  try {
    mailboxPath = JSON.parse(mailboxPathStr);
    if (!Array.isArray(mailboxPath)) {
      return JSON.stringify({
        success: false,
        error: "Mailbox path must be a JSON array",
      });
    }
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Invalid mailbox path JSON: " + e.toString(),
    });
  }

  if (mailboxPath.length === 0) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path array cannot be empty",
    });
  }

  // Prevent replying to drafts - this crashes Mail.app
  const firstMailboxName = mailboxPath[0].toLowerCase();
  if (
    firstMailboxName === "drafts" ||
    firstMailboxName === "entwürfe" ||
    firstMailboxName === "brouillons"
  ) {
    return JSON.stringify({
      success: false,
      error:
        "Cannot reply to draft messages. Drafts are not sent messages and replying to them will crash Mail.app. Use replace_outgoing_message to modify drafts instead.",
    });
  }

  if (!messageId || messageId < 1) {
    return JSON.stringify({
      success: false,
      error: "Message ID is required and must be a positive integer",
    });
  }

  if (!replyContent) {
    return JSON.stringify({
      success: false,
      error: "Reply content is required",
    });
  }

  try {
    // Use name lookup syntax to find account directly
    let targetAccount;
    try {
      targetAccount = Mail.accounts[accountName];
    } catch (e) {
      return JSON.stringify({
        success: false,
        error:
          'Account "' + accountName + '" not found. Error: ' + e.toString(),
      });
    }

    // Verify account exists by trying to access a property
    try {
      targetAccount.name();
    } catch (e) {
      return JSON.stringify({
        success: false,
        error:
          'Account "' +
          accountName +
          '" not found. Please verify the account name is correct.',
      });
    }

    // Navigate to the target mailbox using mailboxPath
    let currentContainer = targetAccount;
    let targetMailbox;

    try {
      for (let i = 0; i < mailboxPath.length; i++) {
        const part = mailboxPath[i];
        try {
          const nextMailbox = currentContainer.mailboxes[part];
          nextMailbox.name(); // Verify existence
          currentContainer = nextMailbox;
        } catch (e) {
          // If lookup fails, gather available mailboxes
          let availableNames = [];
          try {
            const available = currentContainer.mailboxes();
            for (let j = 0; j < available.length; j++) {
              availableNames.push(available[j].name());
            }
          } catch (err) {
            availableNames = ["(Error listing mailboxes)"];
          }

          return JSON.stringify({
            success: false,
            error:
              'Mailbox "' +
              part +
              '" not found in "' +
              (i === 0 ? accountName : mailboxPath[i - 1]) +
              '". Available mailboxes: ' +
              availableNames.join(", "),
          });
        }
      }
      targetMailbox = currentContainer;
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: e.message,
      });
    }

    // Use whose() for fast constant-time message lookup
    const matches = targetMailbox.messages.whose({ id: messageId })();

    if (matches.length === 0) {
      return JSON.stringify({
        success: false,
        error:
          "Message with ID " +
          messageId +
          ' not found in mailbox "' +
          mailboxPath.join(" > ") +
          '". The message may have been deleted or moved.',
      });
    }

    const targetMessage = matches[0];

    // Use Mail.app's built-in reply method to create the reply.
    // This properly sets up threading, headers (In-Reply-To, References),
    // and recipients.
    //
    // Note on content handling: Mail.app's auto-generated rich text quote
    // lives exclusively in the compose window's HTML/WebView layer and is
    // NOT accessible via the OutgoingMessage.content scripting property
    // (which always returns "" with 0 paragraphs). Any write to the content
    // property destroys the HTML-rendered quote. Therefore, we either need to construct
    // the quoted reply ourselves from the original message's plain text
    // content or simply ignore it.
    const replyMessage = targetMessage.reply({
      openingWindow: openingWindow,
      replyToAll: replyToAll,
    });

    if (!replyMessage) {
      return JSON.stringify({
        success: false,
        error:
          "Failed to create reply message. The reply() method returned null.",
      });
    }

    // Build the reply content without quoted original message.
    const originalContent = targetMessage.content();
    const originalSender = targetMessage.sender();
    const originalDate = targetMessage.dateSent();
    const dateStr = originalDate.toLocaleString();

    // Set content based on format
    if (contentFormat === "markdown" && contentJson) {
      // Render styled blocks as rich text
      try {
        const styledBlocks = JSON.parse(contentJson);
        renderStyledBlocks(Mail, replyMessage, styledBlocks, log);
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: "Failed to render rich text: " + e.toString(),
        });
      }
    } else {
      // Plain text - The OutgoingMessage.content property is a RichText object.
      // You cannot assign a plain string directly (fails with
      // "Can't convert types"). Use Mail.make to insert a paragraph
      // into the content object.
      Mail.make({
        new: "paragraph",
        withData: replyContent,
        at: replyMessage.content,
      });
    }

    // Save the reply message (required for non-visible messages)
    replyMessage.save();

    // Get the OutgoingMessage ID directly (no Drafts lookup needed)
    const draftId = replyMessage.id();
    const subject = replyMessage.subject();

    // Get recipient addresses
    const toRecipients = [];
    try {
      const recipients = replyMessage.toRecipients();
      for (let i = 0; i < recipients.length; i++) {
        toRecipients.push(recipients[i].address());
      }
    } catch (e) {
      log("Error reading To recipients: " + e.toString());
    }

    const result = {
      draft_id: draftId,
      subject: subject,
      to_recipients: toRecipients,
      message: "Reply saved to drafts successfully",
    };

    return JSON.stringify({
      success: true,
      data: result,
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to create reply draft: " + e.toString(),
    });
  }
}

/**
 * Renders styled blocks as rich text in the message content
 * @param {Application} Mail - Mail application object
 * @param {Object} msg - Message object
 * @param {Array} styledBlocks - Array of styled block objects
 * @param {Function} log - Logging function
 */
function renderStyledBlocks(Mail, msg, styledBlocks, log) {
  for (let i = 0; i < styledBlocks.length; i++) {
    const block = styledBlocks[i];

    // Create paragraph with styling (all properties are optional)
    // Go code adds newlines to block.text, so no need to append "\n" here
    const props = {};
    if (block.font) {
      props.font = block.font;
    }
    if (block.size) {
      props.size = block.size;
    }
    if (block.color) {
      props.color = block.color;
    }

    Mail.make({
      new: "paragraph",
      withData: block.text,
      withProperties: props,
      at: msg.content,
    });

    // Apply inline styles if present
    if (block.inline_styles && block.inline_styles.length > 0) {
      const paraIndex = msg.content.paragraphs.length - 1;

      for (let j = 0; j < block.inline_styles.length; j++) {
        const style = block.inline_styles[j];

        try {
          // Apply character-level styling
          for (let charIdx = style.start; charIdx < style.end; charIdx++) {
            const char = msg.content.paragraphs[paraIndex].characters[charIdx];
            if (style.font) {
              char.font = style.font;
            }
            if (style.size) {
              char.size = style.size;
            }
            if (style.color) {
              char.color = style.color;
            }
          }
        } catch (e) {
          log("Error applying inline style: " + e.toString());
        }
      }
    }
  }
}
