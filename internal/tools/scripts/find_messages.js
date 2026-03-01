function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

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
  const log = (msg) => logs.push(msg);

  // 3. Argument parsing
  let args;
  try {
    args = JSON.parse(argv[0]);
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to parse input arguments JSON",
    });
  }

  const {
    account: accountName,
    mailboxPath = [],
    limit = 50,
    subject,
    sender,
    readStatus,
    flaggedOnly,
    dateAfter,
    dateBefore,
  } = args;

  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
    });
  }
  if (!Array.isArray(mailboxPath) || mailboxPath.length === 0) {
    return JSON.stringify({ success: false, error: "Mailbox path required" });
  }

  if (limit < 1 || limit > 1000) {
    return JSON.stringify({
      success: false,
      error: "Limit must be between 1 and 1000",
    });
  }

  try {
    const targetAccount = Mail.accounts[accountName];
    try {
      targetAccount.name();
    } catch (e) {
      return JSON.stringify({
        success: false,
        error: `Account "${accountName}" not found.`,
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

    const targetMailbox = findMailboxByPath(targetAccount, mailboxPath);
    if (!targetMailbox) {
      return JSON.stringify({
        success: false,
        error: `Mailbox "${mailboxPath.join(" > ")}" not found in account "${accountName}".`,
      });
    }

    const msgs = targetMailbox.messages;
    const count = msgs.length;
    log(
      `Mailbox contains ${count} messages. Performing bulk property fetch...`,
    );

    // PERFORMANCE OPTIMIZATION:
    // whose({ subject: { _contains: "..." } }) is extremely slow and causes timeouts on large mailboxes.
    // Instead, we fetch property arrays once and filter in JavaScript.
    // This is significantly faster for 10k+ messages as it reduces AppleEvent overhead to a O(1) bulk fetch.

    let filterSubject = subject ? subject.toLowerCase() : null;
    let filterSender = sender ? sender.toLowerCase() : null;
    let filterDateAfter = dateAfter ? new Date(dateAfter) : null;
    let filterDateBefore = dateBefore ? new Date(dateBefore) : null;

    // Fetch only the columns needed for filtering to minimize data transfer
    const subjects = filterSubject ? msgs.subject() : null;
    const senders = filterSender ? msgs.sender() : null;
    const readStatuses =
      readStatus !== undefined && readStatus !== null
        ? msgs.readStatus()
        : null;
    const flaggedStatuses = flaggedOnly ? msgs.flaggedStatus() : null;
    const datesReceived =
      filterDateAfter || filterDateBefore ? msgs.dateReceived() : null;

    const matchingIndices = [];
    for (let i = 0; i < count; i++) {
      if (
        subjects &&
        (!subjects[i] ||
          subjects[i].toLowerCase().indexOf(filterSubject) === -1)
      )
        continue;
      if (
        senders &&
        (!senders[i] || senders[i].toLowerCase().indexOf(filterSender) === -1)
      )
        continue;
      if (readStatuses && readStatuses[i] !== readStatus) continue;
      if (flaggedStatuses && flaggedStatuses[i] !== true) continue;
      if (datesReceived) {
        const d = datesReceived[i];
        if (filterDateAfter && d <= filterDateAfter) continue;
        if (filterDateBefore && d >= filterDateBefore) continue;
      }
      matchingIndices.push(i);
    }

    const totalMatches = matchingIndices.length;
    log(`Found ${totalMatches} matching messages.`);

    const maxProcess = Math.min(totalMatches, limit);
    const resultMessages = [];

    if (maxProcess > 0) {
      const subsetIndices = matchingIndices.slice(0, maxProcess);
      const subsetMsgs = subsetIndices.map((idx) => msgs[idx]);

      for (let i = 0; i < maxProcess; i++) {
        const msg = subsetMsgs[i];

        // 1. Get content safely (often fails on weird/syncing messages)
        let content = "";
        try {
          content = msg.content() || "";
        } catch (e) {
          log("Error reading content for message " + i + ": " + e.toString());
        }

        // 2. Get the rest of the properties
        try {
          // Cache dateSent to avoid double AppleEvents
          const ds = msg.dateSent();

          resultMessages.push({
            id: msg.id(),
            subject: msg.subject(),
            sender: msg.sender(),
            date_received: msg.dateReceived().toISOString(),
            date_sent: ds ? ds.toISOString() : null,
            read_status: msg.readStatus(),
            flagged_status: msg.flaggedStatus(),
            message_size: msg.messageSize(),
            content_preview:
              content.length > 100
                ? content.substring(0, 100) + "..."
                : content,
            content_length: content.length,
            mailbox_path: mailboxPath,
            account: accountName,
          });
        } catch (e) {
          log(
            "Error reading properties for message " + i + ": " + e.toString(),
          );
          // Skip this message and continue
        }
      }
    }

    return JSON.stringify({
      success: true,
      data: {
        messages: resultMessages,
        count: resultMessages.length,
        total_matches: totalMatches,
        limit: limit,
        has_more: totalMatches > limit,
        filters_applied: {
          subject: subject || null,
          sender: sender || null,
          read_status: readStatus !== undefined ? readStatus : null,
          flagged_only: flaggedOnly || false,
          date_after: dateAfter || null,
          date_before: dateBefore || null,
        },
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: e.toString(),
      logs: logs.join("\n"),
    });
  }
}
