#!/usr/bin/osascript -l JavaScript

/**
 * Get message content from Mail.app with nested mailbox support
 *
 * Arguments:
 *   argv[0] - accountName (required)
 *   argv[1] - mailboxPath (required) - JSON array like ["Inbox"] or ["Inbox","GitHub"]
 *   argv[2] - messageId (required) - numeric ID
 *
 * Improvements:
 *   - Supports nested mailboxes via mailboxPath array
 *   - Uses chained name lookup for navigation
 *   - Uses whose() for fast message ID filtering (constant time)
 *   - Proper Object Specifier dereferencing with ()
 *   - Better error handling with descriptive messages
 */

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

  // Parse arguments: accountName, mailboxPath (JSON array), messageId
  const accountName = argv[0] || "";
  const mailboxPathStr = argv[1] || "";
  const messageId = argv[2] ? parseInt(argv[2]) : 0;

  // Validate all required arguments explicitly
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

  if (!messageId || messageId < 1) {
    return JSON.stringify({
      success: false,
      error: "Message ID is required and must be a positive integer",
    });
  }

  // Parse mailboxPath from JSON
  let mailboxPath;
  try {
    mailboxPath = JSON.parse(mailboxPathStr);
    if (!Array.isArray(mailboxPath) || mailboxPath.length === 0) {
      return JSON.stringify({
        success: false,
        error: "Mailbox path must be a non-empty JSON array",
      });
    }
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: `Invalid mailbox path JSON: ${e.toString()}`,
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
        error: `Account "${accountName}" not found. Error: ${e.toString()}`,
      });
    }

    // Verify account exists by trying to access a property
    try {
      targetAccount.name();
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: `Account "${accountName}" not found. Please verify the account name is correct.`,
      });
    }

    // Navigate to the target mailbox using chained name lookup
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

    // Use whose() to filter for the specific message ID
    // This is MUCH faster than looping (constant time vs linear time)
    // whose() returns a list of Object Specifiers, so we need to dereference with ()
    const matchingMessages = targetMailbox.messages.whose({
      id: messageId,
    })();

    if (!matchingMessages || matchingMessages.length === 0) {
      return JSON.stringify({
        success: false,
        error: `Message with ID ${messageId} not found in mailbox "${mailboxPath.join(" > ")}". The message may have been deleted or moved.`,
      });
    }

    // Get the first (and should be only) matching message
    const targetMessage = matchingMessages[0];

    // Get message details with error handling for each field
    const result = {};

    try {
      result.id = targetMessage.id();
    } catch (e) {
      result.id = null;
    }

    try {
      result.subject = targetMessage.subject();
    } catch (e) {
      result.subject = "";
    }

    try {
      result.sender = targetMessage.sender();
    } catch (e) {
      result.sender = "";
    }

    try {
      result.replyTo = targetMessage.replyTo();
    } catch (e) {
      result.replyTo = "";
    }

    try {
      result.dateReceived = targetMessage.dateReceived().toISOString();
    } catch (e) {
      result.dateReceived = null;
    }

    try {
      result.dateSent = targetMessage.dateSent().toISOString();
    } catch (e) {
      result.dateSent = null;
    }

    try {
      result.content = targetMessage.content();
    } catch (e) {
      result.content = "";
    }

    try {
      result.readStatus = targetMessage.readStatus();
    } catch (e) {
      result.readStatus = false;
    }

    try {
      result.flaggedStatus = targetMessage.flaggedStatus();
    } catch (e) {
      result.flaggedStatus = false;
    }

    try {
      result.messageSize = targetMessage.messageSize();
    } catch (e) {
      result.messageSize = 0;
    }

    try {
      result.messageId = targetMessage.messageId();
    } catch (e) {
      result.messageId = "";
    }

    try {
      result.allHeaders = targetMessage.allHeaders();
    } catch (e) {
      result.allHeaders = "";
    }

    // Get recipients with error handling
    result.toRecipients = [];
    try {
      const toRecipients = targetMessage.toRecipients();
      for (let i = 0; i < toRecipients.length; i++) {
        try {
          result.toRecipients.push({
            name: toRecipients[i].name(),
            address: toRecipients[i].address(),
          });
        } catch (e) {
          log("Error reading To recipient " + i + ": " + e.toString());
        }
      }
    } catch (e) {
      log("Error getting To recipients list: " + e.toString());
    }

    result.ccRecipients = [];
    try {
      const ccRecipients = targetMessage.ccRecipients();
      for (let i = 0; i < ccRecipients.length; i++) {
        try {
          result.ccRecipients.push({
            name: ccRecipients[i].name(),
            address: ccRecipients[i].address(),
          });
        } catch (e) {
          log("Error reading CC recipient " + i + ": " + e.toString());
        }
      }
    } catch (e) {
      log("Error getting CC recipients list: " + e.toString());
    }

    result.bccRecipients = [];
    try {
      const bccRecipients = targetMessage.bccRecipients();
      for (let i = 0; i < bccRecipients.length; i++) {
        try {
          result.bccRecipients.push({
            name: bccRecipients[i].name(),
            address: bccRecipients[i].address(),
          });
        } catch (e) {
          log("Error reading BCC recipient " + i + ": " + e.toString());
        }
      }
    } catch (e) {
      log("Error getting BCC recipients list: " + e.toString());
    }

    // Get attachments with error handling
    // Note: mimeType() is unreliable in Mail.app and often fails, so we skip it
    result.attachments = [];
    try {
      const attachments = targetMessage.mailAttachments();
      for (let i = 0; i < attachments.length; i++) {
        const att = attachments[i];
        const attInfo = {};

        try {
          attInfo.name = att.name();
        } catch (e) {
          attInfo.name = "unknown";
        }

        try {
          attInfo.fileSize = att.fileSize();
        } catch (e) {
          attInfo.fileSize = 0;
        }

        try {
          attInfo.downloaded = att.downloaded();
        } catch (e) {
          attInfo.downloaded = false;
        }

        result.attachments.push(attInfo);
      }
    } catch (e) {
      log("Error getting attachments list: " + e.toString());
    }

    return JSON.stringify({
      success: true,
      data: {
        message: result,
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: `Failed to retrieve message content: ${e.toString()}`,
    });
  }
}
