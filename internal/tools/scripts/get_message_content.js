#!/usr/bin/osascript -l JavaScript

function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  // Parse arguments: accountName, mailboxName, messageId
  const accountName = argv[0] || "";
  const mailboxName = argv[1] || "";
  const messageId = argv[2] ? parseInt(argv[2]) : 0;

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

  if (!messageId || messageId < 1) {
    return JSON.stringify({
      success: false,
      error: "Message ID is required and must be a positive integer",
    });
  }

  try {
    // Find specific account and mailbox
    let targetMailbox = null;
    const accounts = Mail.accounts();

    for (let i = 0; i < accounts.length; i++) {
      const account = accounts[i];
      if (account.name() === accountName) {
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

    if (!targetMailbox) {
      return JSON.stringify({
        success: false,
        error: `Mailbox "${mailboxName}" not found in account "${accountName}". Please verify the account and mailbox names are correct.`,
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
        error: `Message with ID ${messageId} not found in the first ${maxIterations} messages of mailbox "${mailboxName}". The message may be in an older part of the mailbox. Please try searching for a more recent message or use a smaller mailbox.`,
      });
    }

    if (!targetMessage) {
      return JSON.stringify({
        success: false,
        error: `Message with ID ${messageId} not found in mailbox "${mailboxName}". The message may have been deleted or moved.`,
      });
    }

    // Get message details with error handling for each field
    const result = {};

    try {
      result.id = targetMessage.id();
    } catch (e) {
      result.id = null;
    }

    try {
      result.subject = targetMessage.subject();
    } catch (e) {
      result.subject = "";
    }

    try {
      result.sender = targetMessage.sender();
    } catch (e) {
      result.sender = "";
    }

    try {
      result.replyTo = targetMessage.replyTo();
    } catch (e) {
      result.replyTo = "";
    }

    try {
      result.dateReceived = targetMessage.dateReceived().toISOString();
    } catch (e) {
      result.dateReceived = null;
    }

    try {
      result.dateSent = targetMessage.dateSent().toISOString();
    } catch (e) {
      result.dateSent = null;
    }

    try {
      result.content = targetMessage.content();
    } catch (e) {
      result.content = "";
    }

    try {
      result.readStatus = targetMessage.readStatus();
    } catch (e) {
      result.readStatus = false;
    }

    try {
      result.flaggedStatus = targetMessage.flaggedStatus();
    } catch (e) {
      result.flaggedStatus = false;
    }

    try {
      result.messageSize = targetMessage.messageSize();
    } catch (e) {
      result.messageSize = 0;
    }

    try {
      result.messageId = targetMessage.messageId();
    } catch (e) {
      result.messageId = "";
    }

    try {
      result.allHeaders = targetMessage.allHeaders();
    } catch (e) {
      result.allHeaders = "";
    }

    // Get recipients with error handling
    result.toRecipients = [];
    try {
      const toRecipients = targetMessage.toRecipients();
      for (let i = 0; i < toRecipients.length; i++) {
        try {
          result.toRecipients.push({
            name: toRecipients[i].name(),
            address: toRecipients[i].address(),
          });
        } catch (e) {
          // Skip recipient if error accessing properties
        }
      }
    } catch (e) {
      // Leave as empty array if error getting recipients list
    }

    result.ccRecipients = [];
    try {
      const ccRecipients = targetMessage.ccRecipients();
      for (let i = 0; i < ccRecipients.length; i++) {
        try {
          result.ccRecipients.push({
            name: ccRecipients[i].name(),
            address: ccRecipients[i].address(),
          });
        } catch (e) {
          // Skip recipient if error accessing properties
        }
      }
    } catch (e) {
      // Leave as empty array if error getting recipients list
    }

    result.bccRecipients = [];
    try {
      const bccRecipients = targetMessage.bccRecipients();
      for (let i = 0; i < bccRecipients.length; i++) {
        try {
          result.bccRecipients.push({
            name: bccRecipients[i].name(),
            address: bccRecipients[i].address(),
          });
        } catch (e) {
          // Skip recipient if error accessing properties
        }
      }
    } catch (e) {
      // Leave as empty array if error getting recipients list
    }

    // Get attachments with error handling
    // Note: mimeType() is unreliable in Mail.app and often fails, so we skip it
    result.attachments = [];
    try {
      const attachments = targetMessage.mailAttachments();
      for (let i = 0; i < attachments.length; i++) {
        const att = attachments[i];
        const attInfo = {};

        try {
          attInfo.name = att.name();
        } catch (e) {
          attInfo.name = "unknown";
        }

        try {
          attInfo.fileSize = att.fileSize();
        } catch (e) {
          attInfo.fileSize = 0;
        }

        try {
          attInfo.downloaded = att.downloaded();
        } catch (e) {
          attInfo.downloaded = false;
        }

        result.attachments.push(attInfo);
      }
    } catch (e) {
      // Leave as empty array if error getting attachments list
    }

    return JSON.stringify({
      success: true,
      data: {
        message: result,
      },
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: `Failed to retrieve message content: ${e.toString()}`,
    });
  }
}
