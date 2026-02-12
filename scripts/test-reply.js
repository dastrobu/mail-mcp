#!/usr/bin/osascript -l JavaScript

/**
 * Test script for prepending a paragraph to a reply's content in Mail.app
 *
 * This script tests whether we can insert a paragraph at the BEGINNING
 * of an OutgoingMessage's content when paragraphs already exist.
 *
 * Strategy:
 *   1. Create reply with openingWindow: true (gets threading + recipients)
 *   2. Insert several "existing" paragraphs to simulate prior content
 *   3. Try multiple prepend strategies and report which ones work
 *   4. Leave compose window open for visual verification
 *
 * Prepend strategies tested:
 *   A. Mail.make paragraph at msg.content (baseline — known to APPEND)
 *   B. Mail.make paragraph at msg.content.paragraphs.beginning
 *   C. Mail.make paragraph at msg.content.paragraphs[0].before
 *   D. Mail.make paragraph at msg.content.characters.beginning
 *   E. Overwrite paragraphs[0] via .set({to: ...})
 *   F. Styled paragraph at paragraphs.beginning (withProperties)
 *
 * Usage:
 *   1. Select a message in Mail.app
 *   2. Run: ./test-reply.js [reply_content]
 *
 * Example:
 *   ./test-reply.js "Thanks for the update!"
 *   ./test-reply.js  # Uses default reply text
 */

function run(argv) {
  var Mail = Application("Mail");
  Mail.includeStandardAdditions = true;

  var replyContent = argv[0] || "This is my reply text";

  console.log("=== Test Reply Prepend Strategies ===");
  console.log("Reply Content:", replyContent);
  console.log("");

  try {
    // ================================================================
    // Get the selected message
    // ================================================================
    console.log("Looking for selected message...");
    var viewers = Mail.messageViewers();

    if (!viewers || viewers.length === 0) {
      console.log("ERROR: No Mail viewer windows are open");
      console.log("");
      console.log("Please:");
      console.log("  1. Open Mail.app");
      console.log("  2. Open a mailbox window");
      console.log("  3. Select a message");
      console.log("  4. Run this script again");
      return;
    }

    var viewer = viewers[0];
    var selectedMessages = viewer.selectedMessages();

    if (!selectedMessages || selectedMessages.length === 0) {
      console.log("ERROR: No message selected");
      console.log("");
      console.log("Please:");
      console.log("  1. Select a message in Mail.app");
      console.log("  2. Run this script again");
      return;
    }

    var targetMessage = selectedMessages[0];
    console.log("Selected:", targetMessage.subject());
    console.log("  From:", targetMessage.sender());
    console.log("  Date:", targetMessage.dateSent().toLocaleString());
    console.log("");

    // ================================================================
    // STEP 1: Create reply (gets threading headers + recipients)
    // ================================================================
    console.log("STEP 1: Creating reply (openingWindow: true)...");
    var replyMessage = targetMessage.reply({
      openingWindow: true,
      replyToAll: false,
    });

    if (!replyMessage) {
      console.log("ERROR: reply() returned null");
      return;
    }

    console.log("  Subject:", replyMessage.subject());
    console.log("  ID:", replyMessage.id());
    console.log("");

    // Let Mail.app finish rendering
    delay(1);

    // ================================================================
    // STEP 2: Insert "existing" content to simulate a message body
    // The content starts empty after reply(). We add paragraphs so
    // we have something to prepend BEFORE.
    //
    // NOTE: This destroys the auto-generated HTML quote, but we need
    // content to exist so we can test prepend strategies.
    // ================================================================
    console.log("STEP 2: Inserting existing content (simulated quote)...");

    var quoteLines = [
      "On " +
        targetMessage.dateSent().toLocaleString() +
        ", " +
        targetMessage.sender() +
        " wrote:",
      "",
    ];

    // Get a few lines from the original message
    var originalContent = "";
    try {
      originalContent = targetMessage.content();
    } catch (e) {
      originalContent = "(could not read original content)";
    }
    var origLines = originalContent.split("\n");
    for (var i = 0; i < Math.min(origLines.length, 10); i++) {
      quoteLines.push("> " + origLines[i]);
    }
    if (origLines.length > 10) {
      quoteLines.push("> ...(truncated)");
    }

    for (var i = 0; i < quoteLines.length; i++) {
      try {
        Mail.make({
          new: "paragraph",
          withData: quoteLines[i] + "\n",
          at: replyMessage.content,
        });
      } catch (e) {
        console.log("  Failed inserting quote line " + i + ":", e.toString());
      }
    }

    var parasAfterInsert = 0;
    try {
      parasAfterInsert = replyMessage.content.paragraphs().length;
    } catch (e) {}

    console.log("  Inserted", quoteLines.length, "quote lines");
    console.log("  Paragraphs count:", parasAfterInsert);
    console.log("");

    // Show current state
    try {
      var currentContent = replyMessage.content();
      console.log("  Current content (first 300 chars):");
      var preview = currentContent.substring(0, 300);
      console.log("  " + preview.split("\n").join("\n  "));
      if (currentContent.length > 300) {
        console.log("  ...(truncated)");
      }
    } catch (e) {}
    console.log("");

    // ================================================================
    // STEP 3: Try each prepend strategy
    // ================================================================
    console.log(
      "STEP 3: Testing prepend strategies (paragraphs exist: " +
        parasAfterInsert +
        ")...",
    );
    console.log("");

    // Helper: check where marker text appears in content
    function checkPosition(label, marker) {
      try {
        var content = replyMessage.content();
        if (!content) {
          console.log("    " + label + ": content is null/empty");
          return;
        }
        var pos = content.indexOf(marker);
        var atStart = pos === 0;
        console.log(
          "    " +
            label +
            ": position=" +
            pos +
            " atStart=" +
            atStart +
            " contentLen=" +
            content.length,
        );

        // Show first paragraph
        try {
          var firstPara = replyMessage.content.paragraphs[0]();
          console.log(
            "    First paragraph: " +
              JSON.stringify(firstPara.substring(0, 80)),
          );
        } catch (e) {}
      } catch (e) {
        console.log("    " + label + " check error: " + e.toString());
      }
    }

    // ------------------------------------------------------------------
    // Strategy A: Baseline — at msg.content (APPENDS)
    // ------------------------------------------------------------------
    console.log(
      "  --- Strategy A: at replyMessage.content (append baseline) ---",
    );
    try {
      var markerA = "[A-APPEND] " + replyContent;
      Mail.make({
        new: "paragraph",
        withData: markerA + "\n",
        at: replyMessage.content,
      });
      console.log("    OK — Mail.make succeeded");
      checkPosition("Result", markerA);
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ------------------------------------------------------------------
    // Strategy B: at paragraphs.beginning
    // ------------------------------------------------------------------
    console.log(
      "  --- Strategy B: at replyMessage.content.paragraphs.beginning ---",
    );
    try {
      var parasB = 0;
      try {
        parasB = replyMessage.content.paragraphs().length;
      } catch (e) {}
      console.log("    Paragraphs before:", parasB);

      var markerB = "[B-BEGINNING] " + replyContent;
      Mail.make({
        new: "paragraph",
        withData: markerB + "\n",
        at: replyMessage.content.paragraphs.beginning,
      });
      console.log("    OK — Mail.make succeeded");
      checkPosition("Result", markerB);
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ------------------------------------------------------------------
    // Strategy C: at paragraphs[0].before
    // ------------------------------------------------------------------
    console.log(
      "  --- Strategy C: at replyMessage.content.paragraphs[0].before ---",
    );
    try {
      var parasC = 0;
      try {
        parasC = replyMessage.content.paragraphs().length;
      } catch (e) {}
      console.log("    Paragraphs before:", parasC);

      if (parasC > 0) {
        var markerC = "[C-BEFORE] " + replyContent;
        Mail.make({
          new: "paragraph",
          withData: markerC + "\n",
          at: replyMessage.content.paragraphs[0].before,
        });
        console.log("    OK — Mail.make succeeded");
        checkPosition("Result", markerC);
      } else {
        console.log("    SKIPPED — no paragraphs exist");
      }
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ------------------------------------------------------------------
    // Strategy D: at characters.beginning
    // ------------------------------------------------------------------
    console.log(
      "  --- Strategy D: at replyMessage.content.characters.beginning ---",
    );
    try {
      var markerD = "[D-CHARS] " + replyContent;
      Mail.make({
        new: "paragraph",
        withData: markerD + "\n",
        at: replyMessage.content.characters.beginning,
      });
      console.log("    OK — Mail.make succeeded");
      checkPosition("Result", markerD);
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ------------------------------------------------------------------
    // Strategy E: Overwrite paragraphs[0] via .set({to: ...})
    // ------------------------------------------------------------------
    console.log("  --- Strategy E: paragraphs[0].set({to: ...}) ---");
    try {
      var parasE = 0;
      try {
        parasE = replyMessage.content.paragraphs().length;
      } catch (e) {}
      console.log("    Paragraphs before:", parasE);

      if (parasE > 0) {
        var oldFirst = replyMessage.content.paragraphs[0]();
        console.log(
          "    Current paragraphs[0]:",
          JSON.stringify(oldFirst.substring(0, 60)),
        );

        var markerE = "[E-SET] " + replyContent;
        replyMessage.content.paragraphs[0].set({ to: markerE + "\n" });
        console.log("    OK — .set() succeeded");
        checkPosition("Result", markerE);
      } else {
        console.log("    SKIPPED — no paragraphs exist");
      }
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ------------------------------------------------------------------
    // Strategy F: Styled paragraph at paragraphs.beginning
    // ------------------------------------------------------------------
    console.log(
      "  --- Strategy F: styled paragraph at paragraphs.beginning ---",
    );
    try {
      var parasF = 0;
      try {
        parasF = replyMessage.content.paragraphs().length;
      } catch (e) {}
      console.log("    Paragraphs before:", parasF);

      var markerF = "[F-STYLED] " + replyContent;
      Mail.make({
        new: "paragraph",
        withData: markerF + "\n",
        withProperties: {
          font: "Helvetica-Bold",
          size: 14,
          color: [0, 0, 50000],
        },
        at: replyMessage.content.paragraphs.beginning,
      });
      console.log("    OK — Mail.make succeeded");
      checkPosition("Result", markerF);

      // Check styling
      try {
        var f0 = replyMessage.content.paragraphs[0].font();
        var s0 = replyMessage.content.paragraphs[0].size();
        console.log("    First para font:", f0, "size:", s0);
      } catch (e) {
        console.log("    Could not read styling:", e.toString());
      }
    } catch (e) {
      console.log("    FAILED: " + e.toString());
    }
    console.log("");

    // ================================================================
    // STEP 4: Final state summary
    // ================================================================
    console.log("=== FINAL STATE ===");
    console.log("");

    try {
      var finalContent = replyMessage.content();
      var finalParas = 0;
      try {
        finalParas = replyMessage.content.paragraphs().length;
      } catch (e) {}

      console.log(
        "Total content length:",
        finalContent ? finalContent.length : 0,
      );
      console.log("Total paragraphs:", finalParas);
      console.log("");

      // Show first 8 paragraphs
      var showFirst = Math.min(finalParas, 8);
      if (showFirst > 0) {
        console.log("First", showFirst, "paragraphs:");
        for (var p = 0; p < showFirst; p++) {
          try {
            var pText = replyMessage.content.paragraphs[p]();
            var displayText = pText.replace(/\n/g, "\\n");
            if (displayText.length > 100) {
              displayText = displayText.substring(0, 97) + "...";
            }
            console.log("  [" + p + "]", JSON.stringify(displayText));
          } catch (e) {
            console.log("  [" + p + "] read error:", e.toString());
          }
        }
      }

      // Show last 3 if there are more
      if (finalParas > 8) {
        console.log("  ... (" + (finalParas - 11) + " paragraphs omitted) ...");
        var startLast = Math.max(showFirst, finalParas - 3);
        for (var p = startLast; p < finalParas; p++) {
          try {
            var pText = replyMessage.content.paragraphs[p]();
            var displayText = pText.replace(/\n/g, "\\n");
            if (displayText.length > 100) {
              displayText = displayText.substring(0, 97) + "...";
            }
            console.log("  [" + p + "]", JSON.stringify(displayText));
          } catch (e) {
            console.log("  [" + p + "] read error:", e.toString());
          }
        }
      }

      console.log("");
      console.log("Full content (first 800 chars):");
      if (finalContent && finalContent.length > 0) {
        var fullPreview = finalContent.substring(0, 800);
        console.log(fullPreview);
        if (finalContent.length > 800) {
          console.log(
            "...(truncated, total " + finalContent.length + " chars)",
          );
        }
      }
    } catch (e) {
      console.log("Final summary error:", e.toString());
    }

    console.log("");

    // Recipients
    try {
      var recipients = replyMessage.toRecipients();
      if (recipients && recipients.length > 0) {
        console.log("Recipients:");
        for (var r = 0; r < recipients.length; r++) {
          console.log("  -", recipients[r].address());
        }
        console.log("");
      }
    } catch (e) {}

    console.log("=== Test Complete ===");
    console.log("Subject:", replyMessage.subject());
    console.log("");
    console.log("Compose window left OPEN for visual verification.");
    console.log(
      "Check which [STRATEGY-X] markers appear at the TOP of the message.",
    );
    console.log(
      "Successful prepend = marker text appears BEFORE the quote lines.",
    );
  } catch (e) {
    console.log("");
    console.log("ERROR:", e.toString());
    console.log("Stack:", e.stack || "N/A");
  }
}
