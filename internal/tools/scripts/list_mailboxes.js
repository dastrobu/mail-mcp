#!/usr/bin/osascript -l JavaScript

/**
 * Lists mailboxes for a specific account with optional sub-mailbox filtering
 *
 * Arguments:
 *   argv[0] - accountName (required)
 *   argv[1] - mailboxPath (optional) - JSON array like [] for top-level or ["Inbox"] for sub-mailboxes
 *
 * Features:
 *   - Lists top-level mailboxes by default
 *   - Lists sub-mailboxes when mailboxPath is provided
 *   - Returns mailboxPath for each mailbox (supports nested navigation)
 *   - Returns unreadCount for each mailbox
 *   - Detects and reports which mailboxes have sub-mailboxes
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

  // Parse arguments: accountName, mailboxPath (JSON array, optional)
  const accountName = argv[0] || "";
  // Handle empty string or missing parameter - treat as empty array
  const mailboxPathStr = (argv[1] && argv[1].trim()) || "[]";

  // Validate account name
  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
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

    // Navigate to the parent mailbox if mailboxPath is provided
    let parentMailbox = null;
    if (mailboxPath.length > 0) {
      let currentContainer = targetAccount;
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
        parentMailbox = currentContainer;
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: e.message,
        });
      }
    }

    // Get mailboxes to list (either from account or parent mailbox)
    const mailboxes = [];
    let sourceMailboxes;

    if (parentMailbox) {
      // List sub-mailboxes of the specified parent
      sourceMailboxes = parentMailbox.mailboxes();
    } else {
      // List top-level mailboxes of the account
      sourceMailboxes = targetAccount.mailboxes();
    }

    // Process each mailbox
    for (let i = 0; i < sourceMailboxes.length; i++) {
      const mailbox = sourceMailboxes[i];

      // Build the full mailbox path for this mailbox
      const currentMailboxPath = [...mailboxPath, mailbox.name()];

      // Check if this mailbox has sub-mailboxes
      let hasSubMailboxes = false;
      let subMailboxCount = 0;
      try {
        const subMailboxes = mailbox.mailboxes();
        subMailboxCount = subMailboxes.length;
        hasSubMailboxes = subMailboxCount > 0;
      } catch (e) {
        log("Error reading sub-mailboxes: " + e.toString());
      }

      // Get unread count
      let unreadCount = 0;
      try {
        unreadCount = mailbox.unreadCount();
      } catch (e) {
        log("Error reading unread count: " + e.toString());
      }

      // Get total message count
      let messageCount = 0;
      try {
        messageCount = mailbox.messages.length;
      } catch (e) {
        log("Error reading message count: " + e.toString());
      }

      mailboxes.push({
        name: mailbox.name(),
        mailboxPath: currentMailboxPath,
        account: accountName,
        unreadCount: unreadCount,
        messageCount: messageCount,
        hasSubMailboxes: hasSubMailboxes,
        subMailboxCount: subMailboxCount,
      });
    }

    // Return success with mailbox data
    return JSON.stringify({
      success: true,
      data: {
        mailboxes: mailboxes,
        count: mailboxes.length,
        parentMailboxPath: mailboxPath.length > 0 ? mailboxPath : null,
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to list mailboxes: " + e.toString(),
    });
  }
}
