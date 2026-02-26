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

  const accountName = args.account || "";
  const messageId = parseInt(args.message_id, 10) || 0;
  const mailboxPath = args.mailbox_path || [];
  const replyToAll = args.reply_to_all === true;

  log(
    `Received arguments: account='${accountName}', messageId=${messageId}, replyToAll=${replyToAll}, path='${JSON.stringify(mailboxPath)}'`,
  );

  if (!accountName || !messageId) {
    return JSON.stringify({
      success: false,
      error: "Account name and message ID are required.",
      errorCode: "MISSING_PARAMETERS",
      logs: logs.join("\n"),
    });
  }

  if (!Array.isArray(mailboxPath) || mailboxPath.length === 0) {
    return JSON.stringify({
      success: false,
      error: "Mailbox path must be a non-empty array.",
      errorCode: "INVALID_MAILBOX_PATH",
      logs: logs.join("\n"),
    });
  }

  // 4. Execution wrapped in try/catch
  try {
    const accounts = Mail.accounts.whose({ name: accountName })();
    if (accounts.length === 0) {
      return JSON.stringify({
        success: false,
        error: `Account '${accountName}' not found.`,
        errorCode: "ACCOUNT_NOT_FOUND",
        logs: logs.join("\n"),
      });
    }
    log(`Successfully found account '${accountName}'.`);

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

    let targetMailbox = findMailboxByPath(accounts[0], mailboxPath);
    if (!targetMailbox) {
        return JSON.stringify({
            success: false,
            error: "Mailbox path '" + mailboxPath.join(" > ") + "' not found in account '" + accountName + "'."
        });
    }

    // --- End of Traversal Logic ---

    const messages = targetMailbox.messages.whose({ id: messageId })();
    if (messages.length === 0) {
      return JSON.stringify({
        success: false,
        error: `Message with ID ${messageId} not found in mailbox '${mailboxPath.join(" > ")}'.`,
        errorCode: "MESSAGE_NOT_FOUND",
        logs: logs.join("\n"),
      });
    }
    const originalMessage = messages[0];
    log(`Found original message with ID ${messageId}.`);

    const replyMessage = originalMessage.reply({
      openingWindow: true,
      replyToAll: replyToAll,
    });

    // NOTE: We are NOT saving the reply. It exists as an open window (OutgoingMessage).
    log("Reply message window created.");

    Mail.activate();

    const mailProcess = SystemEvents.processes.byName("Mail");
    const pid = mailProcess.unixId();
    log(`Got Mail.app PID: ${pid}.`);

    // 5. CRITICAL: Return 'outgoing_id' for the new message window.
    return JSON.stringify({
      success: true,
      data: {
        outgoing_id: replyMessage.id(), // This is now an OutgoingMessage ID
        subject: replyMessage.subject(),
        pid: pid,
        message: "Reply message created successfully.",
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    let errorCode = "UNKNOWN_ERROR";
    if (e.toString().includes("Automation is not allowed")) {
      errorCode = "MAIL_APP_NO_PERMISSIONS";
    }
    log(`Caught error: ${e.toString()}`);
    return JSON.stringify({
      success: false,
      error: `Failed to create reply: ${e.toString()}`,
      errorCode: errorCode,
      logs: logs.join("\n"),
    });
  }
}
