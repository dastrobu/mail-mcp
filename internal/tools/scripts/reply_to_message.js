function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  // Parse arguments: accountName, mailboxName, messageId, replyContent, openingWindow, replyToAll
  const accountName = argv[0] || "";
  const mailboxName = argv[1] || "";
  const messageId = argv[2] ? parseInt(argv[2]) : 0;
  const replyContent = argv[3] || "";
  const openingWindow = argv[4] === "true";
  const replyToAll = argv[5] === "true";

  if (!accountName) {
    return JSON.stringify({
      success: false,
      error: "Account name is required",
    });
  }

  if (!mailboxName) {
    return JSON.stringify({
      success: false,
      error: "Mailbox name is required",
    });
  }

  // Prevent replying to drafts - this crashes Mail.app
  if (
    mailboxName.toLowerCase() === "drafts" ||
    mailboxName.toLowerCase() === "entwürfe" ||
    mailboxName.toLowerCase() === "brouillons"
  ) {
    return JSON.stringify({
      success: false,
      error:
        "Cannot reply to draft messages. Drafts are not sent messages and replying to them will crash Mail.app. Use replace_draft to modify drafts instead.",
    });
  }

  if (!messageId || messageId < 1) {
    return JSON.stringify({
      success: false,
      error: "Message ID is required and must be a positive integer",
    });
  }

  if (!replyContent) {
    return JSON.stringify({
      success: false,
      error: "Reply content is required",
    });
  }

  try {
    // Find specific account and mailbox
    let targetAccount = null;
    let targetMailbox = null;
    const accounts = Mail.accounts();

    for (let i = 0; i < accounts.length; i++) {
      const account = accounts[i];
      if (account.name() === accountName) {
        targetAccount = account;
        const mailboxes = account.mailboxes();
        for (let j = 0; j < mailboxes.length; j++) {
          if (mailboxes[j].name() === mailboxName) {
            targetMailbox = mailboxes[j];
            break;
          }
        }
        break;
      }
    }

    if (!targetAccount) {
      return JSON.stringify({
        success: false,
        error:
          'Account "' +
          accountName +
          '" not found. Please verify the account name is correct.',
      });
    }

    if (!targetMailbox) {
      return JSON.stringify({
        success: false,
        error:
          'Mailbox "' +
          mailboxName +
          '" not found in account "' +
          accountName +
          '". Please verify the mailbox name is correct.',
      });
    }

    // Find the message by ID
    // Note: For large mailboxes, this can be slow
    // We limit the search to 1000 messages to prevent hanging
    let targetMessage = null;
    const messages = targetMailbox.messages();
    const maxIterations = Math.min(messages.length, 1000);

    for (let i = 0; i < maxIterations; i++) {
      if (messages[i].id() === messageId) {
        targetMessage = messages[i];
        break;
      }
    }

    // If not found in first 1000 messages, return error
    if (!targetMessage && messages.length > maxIterations) {
      return JSON.stringify({
        success: false,
        error:
          "Message with ID " +
          messageId +
          " not found in the first " +
          maxIterations +
          ' messages of mailbox "' +
          mailboxName +
          '". The message may be in an older part of the mailbox. Please try searching for a more recent message or use a smaller mailbox.',
      });
    }

    if (!targetMessage) {
      return JSON.stringify({
        success: false,
        error:
          "Message with ID " +
          messageId +
          ' not found in mailbox "' +
          mailboxName +
          '". The message may have been deleted or moved.',
      });
    }

    // Use Mail.app's built-in reply method to create the reply.
    // This properly sets up threading, headers (In-Reply-To, References),
    // and recipients.
    //
    // Note on content handling: Mail.app's auto-generated rich text quote
    // lives exclusively in the compose window's HTML/WebView layer and is
    // NOT accessible via the OutgoingMessage.content scripting property
    // (which always returns "" with 0 paragraphs). Any write to the content
    // property destroys the HTML-rendered quote. Therefore, we either need to construct
    // the quoted reply ourselves from the original message's plain text
    // content or simply ignore it.
    const replyMessage = targetMessage.reply({
      openingWindow: openingWindow,
      replyToAll: replyToAll,
    });

    if (!replyMessage) {
      return JSON.stringify({
        success: false,
        error:
          "Failed to create reply message. The reply() method returned null.",
      });
    }

    // Build the reply content without quoted original message.
    const originalContent = targetMessage.content();
    const originalSender = targetMessage.sender();
    const originalDate = targetMessage.dateSent();
    const dateStr = originalDate.toLocaleString();

    // The OutgoingMessage.content property is a RichText object.
    // You cannot assign a plain string directly (fails with
    // "Can't convert types"). Use Mail.make to insert a paragraph
    // into the content object.
    Mail.make({
      new: "paragraph",
      withData: replyContent,
      at: replyMessage.content,
    });

    // The reply is automatically saved as a draft by Mail.app
    // Get the draft message details
    const draftId = replyMessage.id();
    const subject = replyMessage.subject();

    // Get recipient addresses
    const toRecipients = [];
    try {
      const recipients = replyMessage.toRecipients();
      for (let i = 0; i < recipients.length; i++) {
        toRecipients.push(recipients[i].address());
      }
    } catch (e) {
      // If we can't get recipients, continue
    }

    const result = {
      draft_id: draftId,
      subject: subject,
      to_recipients: toRecipients,
      message: "Reply saved to drafts successfully",
    };

    return JSON.stringify({
      success: true,
      data: result,
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to create reply draft: " + e.toString(),
    });
  }
}
