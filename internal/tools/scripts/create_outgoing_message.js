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
  const subject = argv[0] || "";
  const toRecipientsJson = argv[1] || "";
  const ccRecipientsJson = argv[2] || "";
  const bccRecipientsJson = argv[3] || "";
  const fromAccount = argv[4] || ""; // Account to send from (optional)
  const senderOverride = argv[5] || ""; // Specific sender email (optional)

  try {
    // Validate subject (optional but recommended)
    if (!subject) {
      log("Warning: No subject provided.");
    }

    // Prepare recipients
    let toList = [];
    try {
      if (toRecipientsJson) toList = JSON.parse(toRecipientsJson) || [];
    } catch (e) {
      log("Error parsing To recipients: " + e.toString());
    }

    let ccList = [];
    try {
      if (ccRecipientsJson) ccList = JSON.parse(ccRecipientsJson) || [];
    } catch (e) {
      log("Error parsing CC recipients: " + e.toString());
    }

    let bccList = [];
    try {
      if (bccRecipientsJson) bccList = JSON.parse(bccRecipientsJson) || [];
    } catch (e) {
      log("Error parsing BCC recipients: " + e.toString());
    }

    // Determine the sender
    let senderProperty = {};
    if (senderOverride) {
      senderProperty = { sender: senderOverride };
    } else if (fromAccount) {
      try {
        const accounts = Mail.accounts.whose({ name: fromAccount })();
        if (accounts.length > 0) {
          senderProperty = { sender: accounts[0].emailAddresses()[0] };
        } else {
          log(
            `Warning: Account '${fromAccount}' not found. Using default account.`,
          );
        }
      } catch (e) {
        log(
          `Warning: Failed to find account '${fromAccount}': ${e.toString()}`,
        );
      }
    }

    // Create the message
    const msg = Mail.OutgoingMessage({
      subject: subject,
      visible: true,
      ...senderProperty,
    });

    Mail.outgoingMessages.push(msg);

    // Add recipients
    if (toList.length > 0) {
      toList.forEach((addr) => {
        msg.toRecipients.push(Mail.Recipient({ address: addr }));
      });
    }

    if (ccList.length > 0) {
      ccList.forEach((addr) => {
        msg.ccRecipients.push(Mail.Recipient({ address: addr }));
      });
    }

    if (bccList.length > 0) {
      bccList.forEach((addr) => {
        msg.bccRecipients.push(Mail.Recipient({ address: addr }));
      });
    }

    // Save the message as a draft
    msg.save();

    // Activate Mail.app to bring the window to front
    Mail.activate();

    // Return the draft ID so Go can track it
    return JSON.stringify({
      success: true,
      data: {
        draft_id: msg.id(),
        subject: msg.subject(),
        message:
          "Draft created successfully. Window opened for content pasting.",
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    return JSON.stringify({
      success: false,
      error: "Failed to create draft: " + e.toString(),
      logs: logs.join("\n"),
    });
  }
}
