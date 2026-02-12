#!/usr/bin/osascript -l JavaScript

/**
 * Rich Text Reply Builder
 *
 * Demonstrates that Mail.app's RichText API supports paragraph-level,
 * word-level, and character-level styling (font, size, color).
 *
 * This script creates a new outgoing message with styled content that
 * simulates a rich text reply with a quoted original message.
 *
 * Usage:
 *   ./tmp.js
 *   ./tmp.js "Custom reply text"
 *
 * Findings:
 *   - Mail.make({new: "paragraph", withData, withProperties, at: msg.content}) works
 *   - Paragraph font/size/color can be set after creation
 *   - Character-level and word-level font/size/color work
 *   - attributeRuns cannot be read or created directly
 *   - withProperties {font, size, color} on paragraph creation works
 */

function run(argv) {
  var Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  var replyText =
    argv[0] ||
    "Thanks for the update! I'll review the proposal and get back to you by end of week.";
  var results = [];

  function log(msg) {
    console.log(msg);
    results.push(msg);
  }

  log("=== Rich Text Reply Builder ===");
  log("");

  // ================================================================
  // CONFIG: Styling constants
  // ================================================================
  // Colors are in Apple's 16-bit RGB space (0–65535)
  var COLOR_BLACK = [0, 0, 0];
  var COLOR_GRAY = [32768, 32768, 32768];
  var COLOR_BLUE = [0, 0, 50000];
  var COLOR_DKGRAY = [22000, 22000, 22000];
  var COLOR_PURPLE = [32768, 0, 40000];

  var FONT_BODY = "Helvetica";
  var FONT_BOLD = "Helvetica-Bold";
  var FONT_ITALIC = "Helvetica-Oblique";
  var FONT_MONO = "Menlo-Regular";

  var SIZE_BODY = 13;
  var SIZE_HEADING = 16;
  var SIZE_QUOTE = 12;
  var SIZE_SMALL = 10;

  // Simulated original message
  var originalSender = "Alice Johnson <alice@example.com>";
  var originalDate = "June 15, 2025 at 10:32 AM";
  var originalLines = [
    "Hi there,",
    "",
    "I wanted to follow up on the project proposal we discussed last week.",
    "The key points are:",
    "",
    "1. Timeline has been moved to Q3",
    "2. Budget was approved for the full amount",
    "3. We need to finalize the technical approach by Friday",
    "",
    "Let me know your thoughts.",
    "",
    "Best regards,",
    "Alice",
  ];

  // ================================================================
  // STEP 1: Create the outgoing message
  // ================================================================
  log("STEP 1: Creating outgoing message...");
  var msg;
  try {
    msg = Mail.make({
      new: "outgoingMessage",
      withProperties: {
        subject: "Re: Project Proposal Follow-up",
        visible: true,
      },
    });
    log("  Message created (ID: " + msg.id() + ")");
  } catch (e) {
    log("  FAILED creating message: " + e.toString());
    return results.join("\n");
  }

  // Add a dummy recipient
  try {
    Mail.make({
      new: "toRecipient",
      withProperties: { address: "alice@example.com", name: "Alice Johnson" },
      at: msg.toRecipients,
    });
    log("  Recipient added");
  } catch (e) {
    log("  Recipient FAILED: " + e.toString());
  }
  log("");

  // ================================================================
  // STEP 2: Build the reply body with styling
  // ================================================================
  log("STEP 2: Building styled reply content...");

  // --- Reply text (main body, default styling) ---
  try {
    Mail.make({
      new: "paragraph",
      withData: replyText + "\n",
      withProperties: {
        font: FONT_BODY,
        size: SIZE_BODY,
        color: COLOR_BLACK,
      },
      at: msg.content,
    });
    log("  [para 0] Reply body — OK");
  } catch (e) {
    log("  [para 0] Reply body FAILED: " + e.toString());
  }

  // --- Blank line separator ---
  try {
    Mail.make({
      new: "paragraph",
      withData: "\n",
      at: msg.content,
    });
    log("  [para 1] Blank separator — OK");
  } catch (e) {
    log("  [para 1] Blank separator FAILED: " + e.toString());
  }

  // --- "On ... wrote:" attribution line ---
  try {
    var attribution =
      "On " + originalDate + ", " + originalSender + " wrote:\n";
    Mail.make({
      new: "paragraph",
      withData: attribution,
      withProperties: {
        font: FONT_ITALIC,
        size: SIZE_QUOTE,
        color: COLOR_GRAY,
      },
      at: msg.content,
    });
    log("  [para 2] Attribution — OK");
  } catch (e) {
    log("  [para 2] Attribution FAILED: " + e.toString());
  }

  // --- Blank separator before quote ---
  try {
    Mail.make({
      new: "paragraph",
      withData: "\n",
      at: msg.content,
    });
    log("  [para 3] Blank separator — OK");
  } catch (e) {
    log("  [para 3] Blank separator FAILED: " + e.toString());
  }

  // --- Quoted original message lines ---
  var quoteStartPara = 4; // track which paragraph index the quotes start at
  for (var i = 0; i < originalLines.length; i++) {
    var line = originalLines[i];
    var displayLine = (line.length > 0 ? "> " + line : ">") + "\n";

    try {
      Mail.make({
        new: "paragraph",
        withData: displayLine,
        withProperties: {
          font: FONT_BODY,
          size: SIZE_QUOTE,
          color: COLOR_BLUE,
        },
        at: msg.content,
      });
    } catch (e) {
      log("  Quote line " + i + " FAILED: " + e.toString());
    }
  }
  log(
    "  [para " +
      quoteStartPara +
      "-" +
      (quoteStartPara + originalLines.length - 1) +
      "] Quoted lines — OK",
  );
  log("");

  // ================================================================
  // STEP 3: Apply character-level styling to the ">" prefix
  // ================================================================
  log("STEP 3: Applying character-level styling to quote markers...");
  var styled = 0;

  for (var p = quoteStartPara; p < quoteStartPara + originalLines.length; p++) {
    try {
      var paraText = msg.content.paragraphs[p]();
      // Style the ">" character(s) at the beginning differently
      if (paraText.charAt(0) === ">") {
        // Make the ">" bold and a different shade
        msg.content.paragraphs[p].characters[0].font = FONT_BOLD;
        msg.content.paragraphs[p].characters[0].color = COLOR_PURPLE;
        styled++;
      }
    } catch (e) {
      log("  Char style para[" + p + "] FAILED: " + e.toString());
    }
  }
  log("  Styled " + styled + " quote markers (> chars) bold purple");
  log("");

  // ================================================================
  // STEP 4: Style individual words in the reply body for demo
  // ================================================================
  log("STEP 4: Demonstrating word-level styling in reply body...");

  try {
    var replyWords = msg.content.paragraphs[0].words();
    log("  Reply paragraph has " + replyWords.length + " words");

    // Bold the first 2 words
    var boldCount = Math.min(2, replyWords.length);
    for (var w = 0; w < boldCount; w++) {
      msg.content.paragraphs[0].words[w].font = FONT_BOLD;
    }
    if (boldCount > 0) {
      log("  Bolded first " + boldCount + " words");
    }
  } catch (e) {
    log("  Word styling skipped: " + e.toString());
  }
  log("");

  // ================================================================
  // STEP 5: Verify final content
  // ================================================================
  log("=== Final Verification ===");

  try {
    var finalContent = msg.content();
    var finalParas = msg.content.paragraphs().length;

    log("  Total paragraphs: " + finalParas);
    log("  Total content length: " + finalContent.length + " chars");
    log("");

    // Dump paragraph-level styles
    log("  Paragraph styles:");
    for (var p = 0; p < finalParas; p++) {
      try {
        var pText = msg.content.paragraphs[p]();
        var pFont = msg.content.paragraphs[p].font();
        var pSize = msg.content.paragraphs[p].size();
        var pColor;
        try {
          pColor = msg.content.paragraphs[p].color();
        } catch (ce) {
          pColor = [-1, -1, -1];
        }

        // Truncate text for display
        var displayText = pText.replace(/\n/g, "\\n");
        if (displayText.length > 60) {
          displayText = displayText.substring(0, 57) + "...";
        }

        var colorStr = "?";
        if (pColor && pColor.length >= 3) {
          colorStr =
            Math.round(pColor[0]) +
            "," +
            Math.round(pColor[1]) +
            "," +
            Math.round(pColor[2]);
        }

        log(
          "    [" +
            p +
            "] font=" +
            pFont +
            " size=" +
            pSize +
            " color=[" +
            colorStr +
            "]" +
            " text=" +
            JSON.stringify(displayText),
        );
      } catch (e) {
        log("    [" + p + "] read failed: " + e.toString());
      }
    }
  } catch (e) {
    log("  Final verification FAILED: " + e.toString());
  }

  log("");
  log("=== Done ===");
  log("Compose window left open for visual inspection.");
  log("");
  log("What to check:");
  log("  1. Reply text at top — default size, first words bold");
  log("  2. Attribution line — italic, gray, smaller");
  log("  3. Quoted lines — blue, smaller, '>' in bold purple");
  log("  4. Overall: styled text renders in the compose window");

  // No outer try-catch needed; each step handles its own errors
  // The old "FATAL ERROR" block below is kept as a safety net
  try {
    // intentionally empty — all work is done above
  } catch (e) {
    log("");
    log("FATAL ERROR: " + e.toString());
    log("Stack: " + (e.stack || "N/A"));
  }

  return results.join("\n");
}
