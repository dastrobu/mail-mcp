function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;
  const SystemEvents = Application("System Events");

  // 1. CRITICAL: Check if running FIRST
  if (!Mail.running()) {
    return JSON.stringify({
      success: false,
      error: "Mail.app is not running. Please start Mail.app and try again.",
      errorCode: "MAIL_APP_NOT_RUNNING",
    });
  }

  // 2. Logging setup
  const logs = [];
  function log(message) {
    logs.push(message);
  }

  // 3. Argument parsing & validation
  let args;
  try {
    args = JSON.parse(argv[0]);
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to parse input arguments JSON",
      logs: logs.join("\n"),
    });
  }

  const outgoingIdToReplace = parseInt(args.outgoing_id, 10) || 0;
  const messageId = parseInt(args.message_id, 10) || 0;
  const accountName = args.account || "";
  const mailboxPath = args.mailbox_path || [];
  const replyToAll = args.reply_to_all === true;

  log(
    `Replacing reply. Old outgoing_id: ${outgoingIdToReplace}, Original message_id: ${messageId}`,
  );

  if (
    !outgoingIdToReplace ||
    !messageId ||
    !accountName ||
    mailboxPath.length === 0
  ) {
    return JSON.stringify({
      success: false,
      error: "outgoing_id, message_id, account, and mailbox_path are required.",
      errorCode: "MISSING_PARAMETERS",
      logs: logs.join("\n"),
    });
  }

  // 4. Execution wrapped in try/catch
  try {
    // --- Step 1: Find and delete the old reply message window ---
    const oldReplies = Mail.outgoingMessages.whose({
      id: outgoingIdToReplace,
    })();
    if (oldReplies.length > 0) {
      const oldReply = oldReplies[0];
      log(
        `Found old reply window to replace (Subject: "${oldReply.subject()}"). Deleting it.`,
      );
      Mail.delete(oldReply);
    } else {
      log(
        `Warning: Outgoing message with ID ${outgoingIdToReplace} not found. It might have been closed or sent. Proceeding to create a new reply.`,
      );
    }

    // --- Step 2: Find the original message to reply to ---
    const account = Mail.accounts[accountName];
    try {
      account.name();
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: `Account '${accountName}' not found.`,
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
          try {
            next = current.mailboxes.whose({ name: part })()[0];
          } catch (e) {}

          if (!next) {
            try {
              next = current.mailboxes[part];
              next.name();
            } catch (e) {}
          }
          if (!next) throw new Error("not found");
          current = next;
        }
        return current;
      } catch (e) {}

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
            } catch (e) {
              break;
            }
          }
          if (path.length === targetPath.length) {
            let match = true;
            for (let j = 0; j < path.length; j++) {
              if (path[j] !== targetPath[j]) {
                match = false;
                break;
              }
            }
            if (match) return mbx;
          }
        }
      } catch (e) {}
      return null;
    }

    let targetMailbox = findMailboxByPath(account, mailboxPath);
    if (!targetMailbox) {
      return JSON.stringify({
        success: false,
        error:
          "Mailbox path '" +
          mailboxPath.join(" > ") +
          "' not found in account '" +
          accountName +
          "'.",
      });
    }

    const messages = targetMailbox.messages.whose({ id: messageId })();
    if (messages.length === 0) {
      return JSON.stringify({
        success: false,
        error: `Original message with ID ${messageId} not found in mailbox '${mailboxPath.join(" > ")}'.`,
        errorCode: "MESSAGE_NOT_FOUND",
        logs: logs.join("\n"),
      });
    }
    const originalMessage = messages[0];
    log(`Found original message with ID ${messageId}.`);

    // --- Step 3: Create a new reply from the original message ---
    const newReplyMessage = originalMessage.reply({
      openingWindow: true,
      replyToAll: replyToAll,
    });
    log("New reply message window created.");

    // --- Step 4 (Optional): Apply overrides to the new reply ---
    if (args.subject !== undefined) {
      newReplyMessage.subject = args.subject;
      log(`Set new subject: "${args.subject}"`);
    }

    const updateRecipients = (collection, newRecipients) => {
      if (newRecipients && Array.isArray(newRecipients)) {
        // Clear existing recipients
        const existing = collection();
        for (let i = existing.length - 1; i >= 0; i--) {
          existing[i].delete();
        }
        // Add new ones
        newRecipients.forEach((addr) =>
          collection.push(Mail.Recipient({ address: addr })),
        );
        return true;
      }
      return false;
    };

    if (updateRecipients(newReplyMessage.toRecipients, args.to_recipients))
      log("Replaced To: recipients.");
    if (updateRecipients(newReplyMessage.ccRecipients, args.cc_recipients))
      log("Replaced Cc: recipients.");
    if (updateRecipients(newReplyMessage.bccRecipients, args.bcc_recipients))
      log("Replaced Bcc: recipients.");

    // NOTE: We do NOT save the reply. It remains an open OutgoingMessage.
    Mail.activate();

    const mailProcess = SystemEvents.processes.byName("Mail");
    const pid = mailProcess.unixId();

    // 5. CRITICAL: Return 'outgoing_id' for the new message.
    return JSON.stringify({
      success: true,
      data: {
        outgoing_id: newReplyMessage.id(),
        subject: newReplyMessage.subject(),
        pid: pid,
        message: "Reply was successfully replaced.",
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    log(`Error during reply replacement: ${e.toString()}`);
    return JSON.stringify({
      success: false,
      error: `Failed to replace reply: ${e.toString()}`,
      errorCode: "UNKNOWN_ERROR",
      logs: logs.join("\n"),
    });
  }
}
