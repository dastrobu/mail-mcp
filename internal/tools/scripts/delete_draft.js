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
  function log(message) {
    logs.push(message);
  }

  // 3. Argument Parsing & Validation
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

  const draftId = args.draft_id;

  if (draftId === undefined || draftId === null) {
    return JSON.stringify({
      success: false,
      error: "draft_id is required.",
      errorCode: "MISSING_PARAMETERS",
      logs: logs.join("\n"),
    });
  }

  // 4. Execution wrapped in try/catch
  try {
    let draftFound = null;
    let accountName = "Unknown";

    try {
      const draftsBox = Mail.draftsMailbox();
      const messages = draftsBox.messages.whose({ id: draftId })();

      if (messages.length > 0) {
        draftFound = messages[0];
        log(`Found draft with ID ${draftId} in top-level Drafts mailbox.`);
      }
    } catch (e) {
      log(`Error searching top-level Drafts mailbox: ${e.message}`);
    }

    if (!draftFound) {
      return JSON.stringify({
        success: false,
        error: `Draft with ID ${draftId} not found in the Drafts mailbox.`,
        logs: logs.join("\n"),
      });
    }

    const subject = draftFound.subject();

    try {
      accountName = draftFound.mailbox().account().name();
    } catch (e) {
      log(`Could not get account name for reporting: ${e.message}`);
    }

    // Delete the draft
    Mail.delete(draftFound);
    log(`Deleted draft with ID ${draftId}.`);

    return JSON.stringify({
      success: true,
      data: {
        draft_id: draftId,
        subject: subject,
        account: accountName,
        message: "Draft deleted successfully.",
      },
      logs: logs.join("\n"),
    });
  } catch (e) {
    log(`Error deleting draft: ${e.toString()}`);
    return JSON.stringify({
      success: false,
      error: `Failed to delete draft: ${e.toString()}`,
      logs: logs.join("\n"),
    });
  }
}
