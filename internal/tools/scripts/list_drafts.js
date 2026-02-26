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

  const targetAccountName = args.account || "";
  const limit = args.limit || 50;

  // Validate limit
  if (limit < 1 || limit > 1000) {
    return JSON.stringify({
      success: false,
      error: "Limit must be between 1 and 1000",
    });
  }

  try {
    if (targetAccountName) {
      // Verify account exists by trying to access a property
      try {
        const targetAccount = Mail.accounts[targetAccountName];
        targetAccount.name();
      } catch (e) {
        return JSON.stringify({
          success: false,
          error:
            'Account "' +
            targetAccountName +
            '" not found. Please verify the account name is correct.',
        });
      }
    }

    // Get Drafts mailbox
    // Use Mail.draftsMailbox() which is locale-independent and top-level
    const draftsMailbox = Mail.draftsMailbox();

    // Get all draft messages
    const allDrafts = draftsMailbox.messages();
    const totalDrafts = allDrafts.length;

    const drafts = [];
    let hasMore = false;

    for (let i = 0; i < totalDrafts; i++) {
      if (drafts.length >= limit) {
        hasMore = true;
        break;
      }

      const msg = allDrafts[i];
      let msgAccountName = "";

      try {
        msgAccountName = msg.mailbox().account().name();
      } catch (e) {
        // Local drafts or other edge cases might not have an account name
      }

      // Filter by account if specified
      if (targetAccountName && msgAccountName !== targetAccountName) {
        continue;
      }

      try {
        // Get basic properties
        const id = msg.id();
        const subject = msg.subject();
        const sender = msg.sender();
        const dateReceived = msg.dateReceived();
        const dateSent = msg.dateSent();

        // Get content preview
        let content = "";
        try {
          content = msg.content();
        } catch (e) {
          content = "";
        }

        const contentPreview =
          content.length > 100 ? content.substring(0, 100) + "..." : content;

        // Get recipient counts
        let toCount = 0;
        let ccCount = 0;
        let bccCount = 0;

        try {
          toCount = msg.toRecipients.length;
        } catch (e) {
          toCount = 0;
        }

        try {
          ccCount = msg.ccRecipients.length;
        } catch (e) {
          ccCount = 0;
        }

        try {
          bccCount = msg.bccRecipients.length;
        } catch (e) {
          bccCount = 0;
        }

        // Get recipient addresses
        const toRecipients = [];
        try {
          const toRecips = msg.toRecipients();
          for (let j = 0; j < toRecips.length; j++) {
            toRecipients.push(toRecips[j].address());
          }
        } catch (e) {}

        const ccRecipients = [];
        try {
          const ccRecips = msg.ccRecipients();
          for (let j = 0; j < ccRecips.length; j++) {
            ccRecipients.push(ccRecips[j].address());
          }
        } catch (e) {}

        const bccRecipients = [];
        try {
          const bccRecips = msg.bccRecipients();
          for (let j = 0; j < bccRecips.length; j++) {
            bccRecipients.push(bccRecips[j].address());
          }
        } catch (e) {}

        // Get mailbox name
        let mailboxName = "Drafts";
        try {
          mailboxName = msg.mailbox().name();
        } catch (e) {}

        drafts.push({
          draft_id: id,
          subject: subject,
          sender: sender,
          date_received: dateReceived.toISOString(),
          date_sent: dateSent ? dateSent.toISOString() : null,
          content_preview: contentPreview,
          content_length: content.length,
          to_recipients: toRecipients,
          cc_recipients: ccRecipients,
          bcc_recipients: bccRecipients,
          to_count: toCount,
          cc_count: ccCount,
          bcc_count: bccCount,
          total_recipients: toCount + ccCount + bccCount,
          mailbox: mailboxName,
          account: msgAccountName,
        });
      } catch (e) {
        log("Error reading draft " + i + ": " + e.toString());
        // Skip this draft and continue
      }
    }

    return JSON.stringify({
      success: true,
      data: {
        drafts: drafts,
        count: drafts.length,
        total_drafts: totalDrafts,
        limit: limit,
        has_more: hasMore,
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to list drafts: " + e.toString(),
    });
  }
}
