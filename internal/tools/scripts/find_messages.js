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

  // Parse arguments
  const accountName = argv[0] || "";
  const mailboxPathStr = argv[1] || "";
  const filterOptionsStr = argv[2] || "";
  const limitStr = argv[3] || "";

  // Validate account name
  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
    });
  }

  // Validate mailbox path
  if (!mailboxPathStr) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path is required",
    });
  }

  // Parse mailbox path
  let mailboxPath;
  try {
    mailboxPath = JSON.parse(mailboxPathStr);
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Invalid mailbox path JSON: " + e.toString(),
    });
  }

  if (!Array.isArray(mailboxPath) || mailboxPath.length === 0) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path must be a non-empty array",
    });
  }

  // Parse filter options
  let filterOptions = {};
  if (filterOptionsStr) {
    try {
      filterOptions = JSON.parse(filterOptionsStr);
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: "Invalid filter options JSON: " + e.toString(),
      });
    }
  }

  // Parse and validate limit
  const limit = limitStr ? parseInt(limitStr) : 50;
  if (limit < 1 || limit > 1000) {
    return JSON.stringify({
      success: false,
      error: "Limit must be between 1 and 1000",
    });
  }

  try {
    // Find account using name lookup
    let targetAccount;
    try {
      targetAccount = Mail.accounts[accountName];
      targetAccount.name(); // Verify account exists
    } catch (e) {
      return JSON.stringify({
        success: false,
        error:
          'Account "' +
          accountName +
          '" not found. Please verify the account name is correct.',
      });
    }

    // Navigate to target mailbox using path
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

          throw new Error(
            'Mailbox "' +
              part +
              '" not found in "' +
              (i === 0 ? accountName : mailboxPath[i - 1]) +
              '". Available mailboxes: ' +
              availableNames.join(", "),
          );
        }
      }
      targetMailbox = currentContainer;
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: e.message,
      });
    }

    // Build whose() filter conditions
    const conditions = [];

    // Filter by subject (substring match using _contains)
    if (filterOptions.subject) {
      conditions.push({ subject: { _contains: filterOptions.subject } });
    }

    // Filter by sender (substring match using _contains)
    if (filterOptions.sender) {
      conditions.push({ sender: { _contains: filterOptions.sender } });
    }

    // Filter by read status
    if (filterOptions.readStatus !== undefined) {
      conditions.push({ readStatus: filterOptions.readStatus });
    }

    // Filter by flagged status
    if (filterOptions.flaggedOnly) {
      conditions.push({ flaggedStatus: true });
    }

    // Filter by date after (greater than)
    if (filterOptions.dateAfter) {
      try {
        const dateAfter = new Date(filterOptions.dateAfter);
        conditions.push({ dateReceived: { ">": dateAfter } });
      } catch (e) {
        log("Invalid dateAfter format: " + e.toString());
      }
    }

    // Filter by date before (less than)
    if (filterOptions.dateBefore) {
      try {
        const dateBefore = new Date(filterOptions.dateBefore);
        conditions.push({ dateReceived: { "<": dateBefore } });
      } catch (e) {
        log("Invalid dateBefore format: " + e.toString());
      }
    }

    // Get filtered messages using whose()
    // Note: conditions.length should always be > 0 due to validation in Go
    let filteredMessages;
    if (conditions.length === 1) {
      // Single condition
      filteredMessages = targetMailbox.messages.whose(conditions[0])();
    } else {
      // Multiple conditions - use _and
      filteredMessages = targetMailbox.messages.whose({ _and: conditions })();
    }

    const totalMatches = filteredMessages.length;
    log("Found " + totalMatches + " matching messages");

    // Limit the number of messages to process
    const maxProcess = Math.min(totalMatches, limit);
    const messages = [];

    for (let i = 0; i < maxProcess; i++) {
      const msg = filteredMessages[i];

      try {
        // Get basic properties
        const id = msg.id();
        const subject = msg.subject();
        const sender = msg.sender();
        const dateReceived = msg.dateReceived();
        const dateSent = msg.dateSent();
        const readStatus = msg.readStatus();
        const flaggedStatus = msg.flaggedStatus();
        const messageSize = msg.messageSize();

        // Get content preview
        let content = "";
        try {
          content = msg.content();
        } catch (e) {
          log("Error reading content for message " + id + ": " + e.toString());
          content = "";
        }

        const contentPreview =
          content.length > 100 ? content.substring(0, 100) + "..." : content;

        // Get recipient counts
        let toCount = 0;
        let ccCount = 0;

        try {
          toCount = msg.toRecipients.length;
        } catch (e) {
          log("Error reading To recipients count: " + e.toString());
          toCount = 0;
        }

        try {
          ccCount = msg.ccRecipients.length;
        } catch (e) {
          log("Error reading CC recipients count: " + e.toString());
          ccCount = 0;
        }

        // Get mailbox path for this message
        function getMailboxPath(mailbox, accountName) {
          const path = [];
          let current = mailbox;

          while (current) {
            try {
              const name = current.name();
              if (name === accountName) break;
              path.unshift(name);
              current = current.container();
            } catch (e) {
              break;
            }
          }
          return path;
        }

        let messageMboxPath = mailboxPath;
        try {
          const msgMailbox = msg.mailbox();
          messageMboxPath = getMailboxPath(msgMailbox, accountName);
        } catch (e) {
          log(
            "Error reading mailbox path for message " +
              id +
              ": " +
              e.toString(),
          );
        }

        messages.push({
          id: id,
          subject: subject,
          sender: sender,
          date_received: dateReceived.toISOString(),
          date_sent: dateSent ? dateSent.toISOString() : null,
          read_status: readStatus,
          flagged_status: flaggedStatus,
          message_size: messageSize,
          content_preview: contentPreview,
          content_length: content.length,
          to_count: toCount,
          cc_count: ccCount,
          total_recipients: toCount + ccCount,
          mailbox_path: messageMboxPath,
          account: accountName,
        });
      } catch (e) {
        log("Error reading message " + i + ": " + e.toString());
        // Skip this message and continue
      }
    }

    return JSON.stringify({
      success: true,
      data: {
        messages: messages,
        count: messages.length,
        total_matches: totalMatches,
        limit: limit,
        has_more: totalMatches > limit,
        filters_applied: {
          subject: filterOptions.subject || null,
          sender: filterOptions.sender || null,
          read_status:
            filterOptions.readStatus !== undefined
              ? filterOptions.readStatus
              : null,
          flagged_only: filterOptions.flaggedOnly || false,
          date_after: filterOptions.dateAfter || null,
          date_before: filterOptions.dateBefore || null,
        },
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to find messages: " + e.toString(),
    });
  }
}
