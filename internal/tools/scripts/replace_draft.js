function run(argv) {
  const Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  // Parse arguments
  const draftIdStr = argv[0] || "";
  const rawNewSubject = argv[1] || "";
  const newContent = argv[2] || "";
  const toRecipientsJson = argv[3] || "";
  const ccRecipientsJson = argv[4] || "";
  const bccRecipientsJson = argv[5] || "";
  const newSender = argv[6] || "";
  const openingWindow = argv[7] === "true";

  const draftId = draftIdStr ? parseInt(draftIdStr) : 0;

  // Validate required arguments
  if (!draftId || draftId < 1) {
    return JSON.stringify({
      success: false,
      error: "Draft ID is required and must be a positive integer",
    });
  }

  try {
    // Find the draft in the Drafts mailbox
    const draftsMailbox = Mail.draftsMailbox();
    const messages = draftsMailbox.messages();

    let draftMessage = null;
    for (let i = 0; i < messages.length; i++) {
      if (messages[i].id() === draftId) {
        draftMessage = messages[i];
        break;
      }
    }

    if (!draftMessage) {
      return JSON.stringify({
        success: false,
        error: "Draft with ID " + draftId + " not found in Drafts mailbox",
      });
    }

    // Read existing properties from the draft
    const existingSubject = draftMessage.subject();
    const existingSender = draftMessage.sender();
    const existingContent = draftMessage.content();

    // Read existing recipients
    const existingTo = [];
    try {
      const toRecips = draftMessage.toRecipients();
      for (let i = 0; i < toRecips.length; i++) {
        existingTo.push(toRecips[i].address());
      }
    } catch (e) {
      console.log("Error reading To recipients: " + e.toString());
    }

    const existingCc = [];
    try {
      const ccRecips = draftMessage.ccRecipients();
      for (let i = 0; i < ccRecips.length; i++) {
        existingCc.push(ccRecips[i].address());
      }
    } catch (e) {
      console.log("Error reading CC recipients: " + e.toString());
    }

    const existingBcc = [];
    try {
      const bccRecips = draftMessage.bccRecipients();
      for (let i = 0; i < bccRecips.length; i++) {
        existingBcc.push(bccRecips[i].address());
      }
    } catch (e) {
      console.log("Error reading BCC recipients: " + e.toString());
    }

    // Trim and determine final values (new values override existing)
    const newSubject = rawNewSubject.trim();
    const finalSubject = newSubject || existingSubject;
    const finalContent = newContent || existingContent;
    const finalSender = newSender || existingSender;

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

    // Delete the old draft
    Mail.delete(draftMessage);

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
    // Use Mail.ToRecipient() constructor and push() - Mail.make() doesn't work
    for (let i = 0; i < finalTo.length; i++) {
      if (finalTo[i]) {
        try {
          const recip = Mail.ToRecipient({ address: finalTo[i] });
          msg.toRecipients.push(recip);
        } catch (e) {
          console.log("Error adding To recipient: " + e.toString());
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
          console.log("Error adding CC recipient: " + e.toString());
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
          console.log("Error adding BCC recipient: " + e.toString());
        }
      }
    }

    // Set content
    Mail.make({
      new: "paragraph",
      withData: finalContent,
      at: msg.content,
    });

    // Save the draft (required for visible: false messages)
    msg.save();

    // Wait for draft to be saved to Drafts mailbox
    delay(2);

    // Get the OutgoingMessage details
    const newDraftSubject = msg.subject();
    const newDraftSender = msg.sender();

    // Find the actual draft in Drafts mailbox by subject
    // Note: OutgoingMessage.id() is different from the Message.id() in Drafts
    // Reuse draftsMailbox from earlier in the function
    const draftsMessages = draftsMailbox.messages();
    let newDraftId = null;

    // Search for our draft by subject (most recent match)
    for (let i = draftsMessages.length - 1; i >= 0; i--) {
      if (draftsMessages[i].subject() === newDraftSubject) {
        newDraftId = draftsMessages[i].id();
        break;
      }
    }

    if (!newDraftId) {
      return JSON.stringify({
        success: false,
        error:
          "Draft was created but could not be found in Drafts mailbox. Please check Mail.app manually.",
      });
    }

    // Read back recipients
    const toAddrs = [];
    try {
      const recipients = msg.toRecipients();
      for (let i = 0; i < recipients.length; i++) {
        toAddrs.push(recipients[i].address());
      }
    } catch (e) {
      console.log("Error reading To recipients: " + e.toString());
    }

    const ccAddrs = [];
    try {
      const recipients = msg.ccRecipients();
      for (let i = 0; i < recipients.length; i++) {
        ccAddrs.push(recipients[i].address());
      }
    } catch (e) {
      console.log("Error reading CC recipients: " + e.toString());
    }

    const bccAddrs = [];
    try {
      const recipients = msg.bccRecipients();
      for (let i = 0; i < recipients.length; i++) {
        bccAddrs.push(recipients[i].address());
      }
    } catch (e) {
      console.log("Error reading BCC recipients: " + e.toString());
    }

    // Check if all recipients were added successfully
    let message =
      "Draft replaced successfully (old draft deleted, new draft created with updated properties)";
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
          "Draft replaced successfully, but recipients could not be added";
      } else {
        warning =
          "Some recipients could not be added (" +
          totalAdded +
          " of " +
          totalRequested +
          " added). Please verify recipients in Mail.app.";
        message =
          "Draft replaced successfully, but some recipients could not be added";
      }
    }

    const result = {
      draft_id: newDraftId,
      old_draft_id: draftId,
      subject: newDraftSubject,
      sender: newDraftSender,
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
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to replace draft: " + e.toString(),
    });
  }
}
