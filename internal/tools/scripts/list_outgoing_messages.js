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

  try {
    // Get all OutgoingMessage objects
    const allOutgoing = Mail.outgoingMessages();

    // Build array of message info
    const messages = [];

    for (let i = 0; i < allOutgoing.length; i++) {
      const msg = allOutgoing[i];

      try {
        // Get basic properties
        const id = msg.id();
        const subject = msg.subject();
        const sender = msg.sender();

        // Get content (may be empty)
        let content = "";
        try {
          content = msg.content();
        } catch (e) {
          log("Error reading content: " + e.toString());
          content = "";
        }

        // Get content preview (first 100 chars)
        const contentPreview =
          content.length > 100 ? content.substring(0, 100) + "..." : content;

        // Get recipient counts
        let toCount = 0;
        let ccCount = 0;
        let bccCount = 0;

        try {
          toCount = msg.toRecipients.length;
        } catch (e) {
          log("Error reading To recipients count: " + e.toString());
          toCount = 0;
        }

        try {
          ccCount = msg.ccRecipients.length;
        } catch (e) {
          log("Error reading CC recipients count: " + e.toString());
          ccCount = 0;
        }

        try {
          bccCount = msg.bccRecipients.length;
        } catch (e) {
          log("Error reading BCC recipients count: " + e.toString());
          bccCount = 0;
        }

        // Get recipient addresses
        const toRecipients = [];
        try {
          const toRecips = msg.toRecipients();
          for (let j = 0; j < toRecips.length; j++) {
            toRecipients.push(toRecips[j].address());
          }
        } catch (e) {
          log("Error reading To recipients: " + e.toString());
        }

        const ccRecipients = [];
        try {
          const ccRecips = msg.ccRecipients();
          for (let j = 0; j < ccRecips.length; j++) {
            ccRecipients.push(ccRecips[j].address());
          }
        } catch (e) {
          log("Error reading CC recipients: " + e.toString());
        }

        const bccRecipients = [];
        try {
          const bccRecips = msg.bccRecipients();
          for (let j = 0; j < bccRecips.length; j++) {
            bccRecipients.push(bccRecips[j].address());
          }
        } catch (e) {
          log("Error reading BCC recipients: " + e.toString());
        }

        messages.push({
          outgoing_id: id,
          subject: subject,
          sender: sender,
          content_preview: contentPreview,
          content_length: content.length,
          to_recipients: toRecipients,
          cc_recipients: ccRecipients,
          bcc_recipients: bccRecipients,
          to_count: toCount,
          cc_count: ccCount,
          bcc_count: bccCount,
          total_recipients: toCount + ccCount + bccCount,
        });
      } catch (e) {
        log("Error reading OutgoingMessage " + i + ": " + e.toString());
        // Skip this message and continue
      }
    }

    return JSON.stringify({
      success: true,
      data: {
        messages: messages,
        count: messages.length,
        total_outgoing: allOutgoing.length,
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to list outgoing messages: " + e.toString(),
    });
  }
}
