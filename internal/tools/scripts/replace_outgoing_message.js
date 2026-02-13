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
  const outgoingIdStr = argv[0] || "";
  const rawNewSubject = argv[1] || "";
  const newContent = argv[2] || "";
  const contentFormat = argv[3] || "plain";
  const contentJson = argv[4] || "";
  const toRecipientsJson = argv[5] || "";
  const ccRecipientsJson = argv[6] || "";
  const bccRecipientsJson = argv[7] || "";
  const newSender = argv[8] || "";
  const openingWindow = argv[9] === "true";

  const outgoingId = outgoingIdStr ? parseInt(outgoingIdStr) : 0;

  // Validate required arguments
  if (!outgoingId || outgoingId < 1) {
    return JSON.stringify({
      success: false,
      error: "Outgoing message ID is required and must be a positive integer",
    });
  }

  if (!newContent) {
    return JSON.stringify({
      success: false,
      error:
        "Content is required (cannot preserve rich text from existing message)",
    });
  }

  try {
    // Find the OutgoingMessage by ID
    const allOutgoing = Mail.outgoingMessages();
    let foundMessage = null;

    for (let i = 0; i < allOutgoing.length; i++) {
      if (allOutgoing[i].id() === outgoingId) {
        foundMessage = allOutgoing[i];
        break;
      }
    }

    if (!foundMessage) {
      return JSON.stringify({
        success: false,
        error:
          "OutgoingMessage with ID " +
          outgoingId +
          " not found. The message may have been sent, closed, or Mail.app may have been restarted.",
      });
    }

    // Read existing properties from the OutgoingMessage
    const existingSubject = foundMessage.subject();
    const existingSender = foundMessage.sender();
    const existingContent = foundMessage.content();

    // Read existing recipients
    const existingTo = [];
    try {
      const toRecips = foundMessage.toRecipients();
      for (let i = 0; i < toRecips.length; i++) {
        existingTo.push(toRecips[i].address());
      }
    } catch (e) {
      log("Error reading To recipients: " + e.toString());
    }

    const existingCc = [];
    try {
      const ccRecips = foundMessage.ccRecipients();
      for (let i = 0; i < ccRecips.length; i++) {
        existingCc.push(ccRecips[i].address());
      }
    } catch (e) {
      log("Error reading CC recipients: " + e.toString());
    }

    const existingBcc = [];
    try {
      const bccRecips = foundMessage.bccRecipients();
      for (let i = 0; i < bccRecips.length; i++) {
        existingBcc.push(bccRecips[i].address());
      }
    } catch (e) {
      log("Error reading BCC recipients: " + e.toString());
    }

    // Parse recipient arrays from JSON (empty string means keep existing)
    let finalTo = existingTo;
    if (toRecipientsJson) {
      try {
        finalTo = JSON.parse(toRecipientsJson);
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: "Invalid To recipients JSON: " + e.toString(),
        });
      }
    }

    let finalCc = existingCc;
    if (ccRecipientsJson) {
      try {
        finalCc = JSON.parse(ccRecipientsJson);
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: "Invalid CC recipients JSON: " + e.toString(),
        });
      }
    }

    let finalBcc = existingBcc;
    if (bccRecipientsJson) {
      try {
        finalBcc = JSON.parse(bccRecipientsJson);
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: "Invalid BCC recipients JSON: " + e.toString(),
        });
      }
    }

    // Determine final values (use provided values or keep existing)
    const newSubject = rawNewSubject.trim();
    const finalSubject = newSubject || existingSubject;
    const finalSender = newSender || existingSender;

    // Extract email address from sender (ignore display name)
    // Format can be "Name <email@domain.com>" or just "email@domain.com"
    function extractEmail(senderString) {
      const match = senderString.match(/<([^>]+)>/);
      if (match) {
        return match[1].toLowerCase();
      }
      return senderString.toLowerCase();
    }

    const existingSenderEmail = extractEmail(existingSender);

    // Store reference to matching draft for deletion later
    let draftToDelete = null;

    // Try to find the matching draft from Drafts mailbox
    // Match by multiple properties for precision (defer deletion until after new message is created)
    try {
      const draftsMailbox = Mail.draftsMailbox();
      const allDrafts = draftsMailbox.messages();

      log(
        "Searching for matching draft in Drafts mailbox (total drafts: " +
          allDrafts.length +
          ")",
      );
      log(
        "Looking for - Subject: " +
          existingSubject +
          ", Sender: " +
          existingSender +
          ", To: " +
          existingTo.length +
          ", CC: " +
          existingCc.length +
          ", BCC: " +
          existingBcc.length,
      );

      // Find drafts that match the old message properties
      for (let i = 0; i < allDrafts.length; i++) {
        const draft = allDrafts[i];

        try {
          // Check if subject matches
          const draftSubject = draft.subject();
          if (draftSubject !== existingSubject) {
            log(
              "Draft " +
                i +
                ": subject mismatch ('" +
                draftSubject +
                "' != '" +
                existingSubject +
                "')",
            );
            continue;
          }

          // Check if sender email matches (ignore display name)
          const draftSender = draft.sender();
          const draftSenderEmail = extractEmail(draftSender);
          if (draftSenderEmail !== existingSenderEmail) {
            log(
              "Draft " +
                i +
                ": sender email mismatch ('" +
                draftSenderEmail +
                "' != '" +
                existingSenderEmail +
                "')",
            );
            continue;
          }

          // Check if To recipients match
          let toMatches = true;
          const draftToRecips = draft.toRecipients();
          if (draftToRecips.length !== existingTo.length) {
            toMatches = false;
          } else {
            for (let j = 0; j < draftToRecips.length; j++) {
              if (draftToRecips[j].address() !== existingTo[j]) {
                toMatches = false;
                break;
              }
            }
          }

          if (!toMatches) {
            continue;
          }

          // Check if CC recipients match
          let ccMatches = true;
          const draftCcRecips = draft.ccRecipients();
          if (draftCcRecips.length !== existingCc.length) {
            ccMatches = false;
          } else {
            for (let j = 0; j < draftCcRecips.length; j++) {
              if (draftCcRecips[j].address() !== existingCc[j]) {
                ccMatches = false;
                break;
              }
            }
          }

          if (!ccMatches) {
            continue;
          }

          // Check if BCC recipients match
          let bccMatches = true;
          const draftBccRecips = draft.bccRecipients();
          if (draftBccRecips.length !== existingBcc.length) {
            bccMatches = false;
          } else {
            for (let j = 0; j < draftBccRecips.length; j++) {
              if (draftBccRecips[j].address() !== existingBcc[j]) {
                bccMatches = false;
                break;
              }
            }
          }

          if (!bccMatches) {
            continue;
          }

          // All properties match - store for deletion later
          draftToDelete = draft;
          log(
            "Found matching draft in Drafts mailbox at index " +
              i +
              ": " +
              existingSubject +
              " (sender: " +
              existingSender +
              ", " +
              existingTo.length +
              " To, " +
              existingCc.length +
              " CC, " +
              existingBcc.length +
              " BCC)",
          );
          break; // Found the matching draft
        } catch (e) {
          log("Error checking draft " + i + ": " + e.toString());
        }
      }

      if (!draftToDelete) {
        log("No matching draft found in Drafts mailbox");
      }
    } catch (e) {
      log("Error accessing Drafts mailbox: " + e.toString());
    }

    // Create new outgoing message with updated properties
    const msgProps = {
      subject: finalSubject,
      visible: openingWindow,
    };

    if (finalSender) {
      msgProps.sender = finalSender;
    }

    const msg = Mail.make({
      new: "outgoingMessage",
      withProperties: msgProps,
    });

    // Add To recipients
    for (let i = 0; i < finalTo.length; i++) {
      if (finalTo[i]) {
        try {
          const recip = Mail.ToRecipient({ address: finalTo[i] });
          msg.toRecipients.push(recip);
        } catch (e) {
          log("Error adding To recipient: " + e.toString());
        }
      }
    }

    // Add CC recipients
    for (let i = 0; i < finalCc.length; i++) {
      if (finalCc[i]) {
        try {
          const recip = Mail.CcRecipient({ address: finalCc[i] });
          msg.ccRecipients.push(recip);
        } catch (e) {
          log("Error adding CC recipient: " + e.toString());
        }
      }
    }

    // Add BCC recipients
    for (let i = 0; i < finalBcc.length; i++) {
      if (finalBcc[i]) {
        try {
          const recip = Mail.BccRecipient({ address: finalBcc[i] });
          msg.bccRecipients.push(recip);
        } catch (e) {
          log("Error adding BCC recipient: " + e.toString());
        }
      }
    }

    // Set content based on format
    if (contentFormat === "markdown" && contentJson) {
      // Render styled blocks as rich text
      try {
        const styledBlocks = JSON.parse(contentJson);
        renderStyledBlocks(Mail, msg, styledBlocks, log);
      } catch (e) {
        return JSON.stringify({
          success: false,
          error: "Failed to render rich text: " + e.toString(),
        });
      }
    } else {
      // Plain text
      Mail.make({
        new: "paragraph",
        withData: newContent,
        at: msg.content,
      });
    }

    // Save the new message (prevents old message from being saved to Drafts)
    msg.save();

    // Delete the old OutgoingMessage from memory
    try {
      Mail.delete(foundMessage);
      log("Deleted old OutgoingMessage from memory");
    } catch (e) {
      log("Error deleting old OutgoingMessage: " + e.toString());
    }

    // Now that new message is saved, delete the old draft and OutgoingMessage
    if (draftToDelete) {
      try {
        Mail.delete(draftToDelete);
        log(
          "Deleted old draft from Drafts mailbox after new message was saved",
        );
      } catch (e) {
        log("Error deleting old draft: " + e.toString());
      }
    }

    // Get the new OutgoingMessage ID
    const newOutgoingId = msg.id();
    const newSubjectResult = msg.subject();
    const newSenderResult = msg.sender();

    // Read back recipients
    const toAddrs = [];
    try {
      const recipients = msg.toRecipients();
      for (let i = 0; i < recipients.length; i++) {
        toAddrs.push(recipients[i].address());
      }
    } catch (e) {
      log("Error reading To recipients: " + e.toString());
    }

    const ccAddrs = [];
    try {
      const recipients = msg.ccRecipients();
      for (let i = 0; i < recipients.length; i++) {
        ccAddrs.push(recipients[i].address());
      }
    } catch (e) {
      log("Error reading CC recipients: " + e.toString());
    }

    const bccAddrs = [];
    try {
      const recipients = msg.bccRecipients();
      for (let i = 0; i < recipients.length; i++) {
        bccAddrs.push(recipients[i].address());
      }
    } catch (e) {
      log("Error reading BCC recipients: " + e.toString());
    }

    // Check if all recipients were added successfully
    let message =
      "OutgoingMessage replaced successfully (old message deleted, new message created)";
    let warning = null;
    const requestedToCount = finalTo.length;
    const requestedCcCount = finalCc.length;
    const requestedBccCount = finalBcc.length;
    const totalRequested =
      requestedToCount + requestedCcCount + requestedBccCount;
    const totalAdded = toAddrs.length + ccAddrs.length + bccAddrs.length;

    if (totalRequested > 0 && totalAdded < totalRequested) {
      if (totalAdded === 0) {
        warning =
          "No recipients could be added. Please add recipients manually in Mail.app before sending.";
        message =
          "OutgoingMessage replaced successfully, but recipients could not be added";
      } else {
        warning =
          "Some recipients could not be added (" +
          totalAdded +
          " of " +
          totalRequested +
          " added). Please verify recipients in Mail.app.";
        message =
          "OutgoingMessage replaced successfully, but some recipients could not be added";
      }
    }

    const result = {
      outgoing_id: newOutgoingId,
      old_outgoing_id: outgoingId,
      subject: newSubjectResult,
      sender: newSenderResult,
      to_recipients: toAddrs,
      cc_recipients: ccAddrs,
      bcc_recipients: bccAddrs,
      message: message,
    };

    if (warning) {
      result.warning = warning;
    }

    return JSON.stringify({
      success: true,
      data: result,
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to replace outgoing message: " + e.toString(),
    });
  }
}

/**
 * Renders styled blocks as rich text in the message content
 * @param {Application} Mail - Mail application object
 * @param {Object} msg - Message object
 * @param {Array} styledBlocks - Array of styled block objects
 * @param {Function} log - Logging function
 */
function renderStyledBlocks(Mail, msg, styledBlocks, log) {
  for (let i = 0; i < styledBlocks.length; i++) {
    const block = styledBlocks[i];

    // Create paragraph with styling (all properties are optional)
    // Go code adds newlines to block.text, so no need to append "\n" here
    const props = {};
    if (block.font) {
      props.font = block.font;
    }
    if (block.size) {
      props.size = block.size;
    }
    if (block.color) {
      props.color = block.color;
    }

    Mail.make({
      new: "paragraph",
      withData: block.text,
      withProperties: props,
      at: msg.content,
    });

    // Apply inline styles if present
    if (block.inline_styles && block.inline_styles.length > 0) {
      const paraIndex = msg.content.paragraphs.length - 1;

      for (let j = 0; j < block.inline_styles.length; j++) {
        const style = block.inline_styles[j];

        try {
          // Apply character-level styling
          for (let charIdx = style.start; charIdx < style.end; charIdx++) {
            const char = msg.content.paragraphs[paraIndex].characters[charIdx];
            if (style.font) {
              char.font = style.font;
            }
            if (style.size) {
              char.size = style.size;
            }
            if (style.color) {
              char.color = style.color;
            }
          }
        } catch (e) {
          log("Error applying inline style: " + e.toString());
        }
      }
    }
  }
}
