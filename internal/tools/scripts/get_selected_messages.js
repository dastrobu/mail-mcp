#!/usr/bin/osascript -l JavaScript

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

  const limit = args.limit || 5;
  const startAt = 0;

  if (limit < 1) {
    return JSON.stringify({
      success: false,
      error: "Limit must be at least 1",
    });
  }

  if (limit > 100) {
    return JSON.stringify({
      success: false,
      error: "Limit cannot exceed 100",
    });
  }

  try {
    // Get the selected messages from the frontmost Mail viewer
    const viewers = Mail.messageViewers();

    if (!viewers || viewers.length === 0) {
      return JSON.stringify({
        success: false,
        error: "No Mail viewer windows are open",
      });
    }

    // Get the frontmost viewer
    const viewer = viewers[0];
    const selectedMessages = viewer.selectedMessages();

    if (!selectedMessages || selectedMessages.length === 0) {
      return JSON.stringify({
        success: true,
        data: {
          selectedMessagesCount: 0,
          messages: [],
        },
      });
    }

    const selectedMessagesCount = selectedMessages.length;

    // Check if startAt is beyond available messages
    if (startAt >= selectedMessagesCount) {
      return JSON.stringify({
        success: true,
        data: {
          selectedMessagesCount: selectedMessagesCount,
          messages: [],
        },
      });
    }

    // Collect known account names for mailbox path traversal
    const accountNames = [];
    try {
      const accounts = Mail.accounts();
      for (let j = 0; j < accounts.length; j++) {
        try {
          accountNames.push(accounts[j].name());
        } catch (e) {}
      }
    } catch (e) {}

    // Helper function to build mailbox path
    // Returns array like ["Inbox", "GitHub"] for nested mailboxes
    function getMailboxPath(mailbox, accountName) {
      const path = [];
      let current = mailbox;

      // Walk up the mailbox tree until we reach the account
      while (current) {
        try {
          const name = current.name();

          // Stop if we've reached the account level
          if (
            (accountName && name === accountName) ||
            accountNames.includes(name)
          ) {
            break;
          }

          path.unshift(name);

          // Try to get container (parent mailbox or account)
          try {
            current = current.container();
          } catch (e) {
            // No container, stop
            break;
          }
        } catch (e) {
          break;
        }
      }

      return path;
    }

    // Extract message details (limited by limit parameter, starting at startAt)
    const result = [];
    const endAt = Math.min(startAt + limit, selectedMessagesCount);
    for (let i = startAt; i < endAt; i++) {
      const msg = selectedMessages[i];

      // Get mailbox and account information safely
      let mailboxName = "";
      let accountName = "";
      let mailboxPath = [];

      try {
        const mailbox = msg.mailbox();
        if (mailbox) {
          try {
            mailboxName = mailbox.name();
          } catch (e) {}

          try {
            const acc = mailbox.account();
            if (acc) {
              accountName = acc.name();
            }
          } catch (e) {}

          // If account is still not determined, traverse containers to find if parent is an account
          if (!accountName) {
            let cur = mailbox;
            while (cur) {
              try {
                cur = cur.container();
                if (cur) {
                  const curName = cur.name();
                  if (accountNames.includes(curName)) {
                    accountName = curName;
                    break;
                  }
                }
              } catch (e) {
                break;
              }
            }
          }

          mailboxPath = getMailboxPath(mailbox, accountName);
        }
      } catch (e) {}

      // Dates handling (safely convert to ISO if valid date objects)
      let dateReceivedISO = null;
      try {
        const dr = msg.dateReceived();
        if (dr) {
          dateReceivedISO = dr.toISOString();
        }
      } catch (e) {}

      let dateSentISO = null;
      try {
        const ds = msg.dateSent();
        if (ds) {
          dateSentISO = ds.toISOString();
        }
      } catch (e) {}

      result.push({
        id: msg.id(),
        subject: msg.subject(),
        sender: msg.sender(),
        dateReceived: dateReceivedISO,
        dateSent: dateSentISO,
        readStatus: msg.readStatus(),
        flaggedStatus: msg.flaggedStatus(),
        junkMailStatus: msg.junkMailStatus(),
        mailbox: mailboxName,
        mailboxPath: mailboxPath,
        account: accountName,
      });
    }

    return JSON.stringify({
      success: true,
      data: {
        messages: result,
        count: result.length,
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    let errorCode = "UNKNOWN_ERROR";
    if (
      e.errorNumber === -1743 ||
      e.toString().includes("Automation is not allowed")
    ) {
      errorCode = "MAIL_APP_NO_PERMISSIONS";
    }
    return JSON.stringify({
      success: false,
      error: e.toString(),
      errorCode: errorCode,
    });
  }
}
