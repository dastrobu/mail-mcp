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

  const logs = [];
  function log(message) {
    logs.push(message);
  }

  // Parse arguments matching the Go caller:
  // argv[0]: accountName
  // argv[1]: mailboxPathStr (JSON array)
  // argv[2]: messageId
  // argv[3]: replyToAll ("true" or "false")
  // Note: content and contentFormat are now handled exclusively by Go via paste.
  const accountName = argv[0] || "";
  const mailboxPathStr = argv[1] || "";
  const messageId = argv[2] ? parseInt(argv[2]) : 0;
  const replyToAll = argv[3] === "true";

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

  if (!messageId || messageId < 1) {
    return JSON.stringify({
      success: false,
      error: "Message ID is required and must be a positive integer",
    });
  }

  try {
    // Find Account
    let targetAccount;
    try {
      targetAccount = Mail.accounts[accountName];
      targetAccount.name(); // Verify existence
    } catch (e) {
      return JSON.stringify({
        success: false,
        error:
          'Account "' +
          accountName +
          '" not found. Please verify the account name is correct.',
      });
    }

    // Navigate to Mailbox
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
          return JSON.stringify({
            success: false,
            error:
              'Mailbox "' +
              part +
              '" not found in "' +
              (i === 0 ? accountName : mailboxPath[i - 1]) +
              '".',
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

    // Find Message (constant-time lookup)
    const matches = targetMailbox.messages.whose({ id: messageId })();

    if (matches.length === 0) {
      return JSON.stringify({
        success: false,
        error:
          "Message with ID " +
          messageId +
          ' not found in mailbox "' +
          mailboxPath.join(" > ") +
          '".',
      });
    }

    const targetMessage = matches[0];

    // Create Reply Window
    // openingWindow is forced to true as Go requires focus for the paste operation
    const replyMessage = targetMessage.reply({
      openingWindow: true,
      replyToAll: replyToAll,
    });

    if (!replyMessage) {
      return JSON.stringify({
        success: false,
        error:
          "Failed to create reply message. The reply() method returned null.",
      });
    }

    // Save the message as a draft
    replyMessage.save();

    // Activate Mail to bring it to front (helps Go code find the window)
    Mail.activate();

    const draftId = replyMessage.id();
    const subject = replyMessage.subject();

    // Return immediate success so Go can take over
    return JSON.stringify({
      success: true,
      data: {
        draft_id: draftId,
        subject: subject,
        message: "Reply window opened",
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to open reply window: " + e.toString(),
      logs: logs.join("\n"),
    });
  }
}
