#!/usr/bin/osascript -l JavaScript

/**
 * Lists mailboxes for a specific account with optional sub-mailbox filtering
 *
 * Arguments:
 *   argv[0] - JSON string containing:
 *     - account (required)
 *     - mailboxPath (optional) - Array like [] for top-level or ["Inbox"] for sub-mailboxes
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

  // Parse arguments
  let args;
  try {
    args = JSON.parse(argv[0]);
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to parse input arguments JSON",
    });
  }

  const accountName = args.account || "";
  const mailboxPath = args.mailboxPath || [];

  // Validate account name
  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
    });
  }

  if (!Array.isArray(mailboxPath)) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path must be a JSON array",
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

        // Robust mailbox traversal function
    function findMailboxByPath(account, targetPath) {
        if (!targetPath || targetPath.length === 0) return account;
        
        try {
            let current = account;
            for (let i = 0; i < targetPath.length; i++) {
                const part = targetPath[i];
                let next = null;
                try { next = current.mailboxes.whose({name: part})()[0]; } catch(e){}
                
                if (!next) { try { next = current.mailboxes[part]; next.name(); } catch(e){} }
                if (!next) throw new Error("not found");
                current = next;
            }
            return current;
        } catch(e) {}

        try {
            const allMailboxes = account.mailboxes();
            for (let i = 0; i < allMailboxes.length; i++) {
                const mbx = allMailboxes[i];
                const path = [];
                let current = mbx;
                while (current) {
                    try {
                        const name = current.name();
                        if (name === account.name()) break;
                        path.unshift(name);
                        current = current.container();
                    } catch (e) { break; }
                }
                if (path.length === targetPath.length) {
                    let match = true;
                    for (let j = 0; j < path.length; j++) {
                        if (path[j] !== targetPath[j]) { match = false; break; }
                    }
                    if (match) return mbx;
                }
            }
        } catch(e) {}
        return null;
    }

    let parentMailbox = findMailboxByPath(targetAccount, mailboxPath);
    if (mailboxPath.length > 0 && !parentMailbox) {
        return JSON.stringify({
            success: false,
            error: "Mailbox path '" + mailboxPath.join(" > ") + "' not found in account '" + accountName + "'."
        });
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
