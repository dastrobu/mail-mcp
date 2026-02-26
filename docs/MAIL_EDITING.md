# Mail.app Content Editing via JXA — Analysis & Findings

This document summarises the investigation into programmatically editing email
content in Apple Mail via JXA (JavaScript for Automation). It covers what works,
what does not, and why — so future contributors do not have to repeat the same
experiments.

---

## Table of Contents

- [Architecture: Two Content Layers](#architecture-two-content-layers)
- [Key Scripting Objects](#key-scripting-objects)
- [Setting Content on an OutgoingMessage](#setting-content-on-an-outgoingmessage)
- [RichText Styling Capabilities](#richtext-styling-capabilities)
- [The Auto-Generated Quote Problem](#the-auto-generated-quote-problem)
- [Approaches Tested](#approaches-tested)
- [The Working Approach](#the-working-approach)
- [Modifying Existing Drafts](#modifying-existing-drafts)
- [Nested Mailbox Support](#nested-mailbox-support)
- [JXA Best Practices](#jxa-best-practices)
- [Alternatives to JXA](#alternatives-to-jxa)
- [ObjC Bridge from JXA](#objc-bridge-from-jxa)
- [Useful Properties & Helpers](#useful-properties--helpers)
- [Implications for New Message Composition](#implications-for-new-message-composition)
- [Recommendations](#recommendations)

---

## Architecture: Two Content Layers

Mail.app's compose window internally maintains **two independent layers** for
message content:

| Layer | Accessible via scripting? | Description |
|---|---|---|
| **HTML / WebView** | ❌ No | The rich text rendering shown in the compose window. This is where Mail.app places the auto-generated quoted original message (with styled `<blockquote>`, coloured quote bar, formatted headers). It is part of the MIME HTML body. |
| **`content` (RichText)** | ✅ Yes | A parallel `RichText` scripting object exposed on `OutgoingMessage`. It starts empty (`""`, 0 paragraphs) even when the HTML layer already contains content. |

These two layers are **not synchronised**:

- Reading `content()` does **not** return the HTML-rendered text.
- Writing to `content` does **not** merge with the HTML — it **replaces** it
  entirely.

This disconnect is the root cause of every content-editing difficulty described
below.

## Key Scripting Objects

### `OutgoingMessage` (Mail Suite)

Returned by `reply()`, `forward()`, or `Mail.make({new: "outgoingMessage"})`.

| Property | Type | Access | Notes |
|---|---|---|---|
| `content` | RichText | read-write | Writable only via `Mail.make` (see below). Direct string assignment fails. |
| `subject` | text | read-write | |
| `sender` | text | read-write | |
| `visible` | boolean | read-write | Controls compose window visibility. |
| `id` | integer | read-only | Changes when the draft is saved/reopened. |

### `Message` (Mail Framework)

A received or saved message in a mailbox (including Drafts).

| Property | Type | Access | Notes |
|---|---|---|---|
| `content` | RichText | **read-only** | `Mail.make` calls appear to succeed but have no effect. |
| `source` | text | read-only | Full MIME source including HTML. |
| `id` | integer | read-only | Stable within a mailbox. Different from `OutgoingMessage.id`. |

### `RichText` (Text Suite)

The type of the `content` property. Contains `paragraphs`, `words`,
`characters`, and `attributeRuns`.

**Critical constraint:** You cannot assign a plain JavaScript string to a
RichText property. Doing so produces:

```
Error: Can't convert types.
```

## Setting Content on an OutgoingMessage

The **only working method** for setting content on an `OutgoingMessage` is:

```javascript
Mail.make({
  new: "paragraph",
  withData: "Your text here\n",
  at: replyMessage.content,
});
```

### What does NOT work

| Attempt | Error |
|---|---|
| `replyMessage.content = "text"` | `Can't convert types.` |
| `replyMessage.content = Mail.make({new: "richText", withData: "text"})` | `Can't make or move that element into that container.` |
| `Mail.make({new: "richText", at: msg})` | `Can't make or move that element into that container.` |
| `Mail.make({new: "attributeRun", …, at: msg.content})` | `Can't make or move that element into that container.` |
| `Mail.make({new: "paragraph", …, at: msg.content.paragraphs.end})` | `Invalid key form.` (when paragraphs array is empty) |
| `Mail.make({new: "paragraph", …, at: msg.content.paragraphs.beginning})` | `Invalid key form.` (fails even when paragraphs exist — see below) |
| `Mail.make({new: "paragraph", …, at: msg.content.paragraphs[0].before})` | `Invalid index.` (fails even when paragraphs exist — see below) |
| `Mail.make({new: "paragraph", …, at: msg.content.paragraphs[0]})` | Works only when paragraphs already exist; replaces rather than prepends. |
| `Mail.make({new: "character", …, at: msg.content.characters.beginning})` | `Invalid key form.` (fails even when characters exist — see below) |
| `msg.content.paragraphs[0].set({to: "text"})` | `Can't convert types.` |

### Reading content

Always use the method-call form with parentheses:

```javascript
// ✅ Correct — returns a string
const text = message.content();

// ❌ Wrong — returns the RichText specifier object, which cannot be
//    logged or converted (throws "Can't convert types")
console.log(message.content);
```

### Prepending: no working insertion point

**There is no way to insert a paragraph at the beginning of existing content.**

All JXA insertion-point specifiers for prepending were tested with existing
paragraphs present (confirmed via `scripts/test-reply.js`). Every strategy
fails:

| Strategy | Specifier | Error |
|---|---|---|
| `paragraphs.beginning` | `at: msg.content.paragraphs.beginning` | `Invalid key form.` |
| `paragraphs[0].before` | `at: msg.content.paragraphs[0].before` | `Invalid index.` |
| `characters.beginning` | `at: msg.content.characters.beginning` | `Invalid key form.` |
| `paragraphs[0].set()` | `msg.content.paragraphs[0].set({to: "text"})` | `Can't convert types.` |
| styled `paragraphs.beginning` | `at: msg.content.paragraphs.beginning` (with `withProperties`) | `Invalid key form.` |

The **only working insertion point** is `at: msg.content`, which always
**appends** to the end.

**Implication:** Reply content must be inserted **before** the quoted text.
Build the message in the correct order from the start — reply body first,
attribution line second, quoted lines last. This is the approach used in the
production `reply_to_message.js` script.

## RichText Styling Capabilities

Despite the limitations with the HTML layer, the RichText scripting API
supports **paragraph-level, word-level, and character-level styling** through
the `font`, `size`, and `color` properties. This was confirmed experimentally
using `scripts/tmp.js`.

### Styled paragraph creation

Paragraphs can be created with inline styling using `withProperties`:

```javascript
Mail.make({
  new: "paragraph",
  withData: "Bold heading text\n",
  withProperties: {
    font: "Helvetica-Bold",
    size: 16,
    color: [0, 0, 0],
  },
  at: msg.content,
});
```

### Post-creation styling

Styles can be applied to existing paragraphs, words, and characters after
creation:

```javascript
// Paragraph-level
msg.content.paragraphs[0].font = "Helvetica-Bold";
msg.content.paragraphs[0].size = 18;
msg.content.paragraphs[0].color = [65535, 0, 0]; // red

// Word-level
msg.content.paragraphs[0].words[0].font = "Helvetica-Bold";
msg.content.paragraphs[0].words[0].color = [0, 0, 65535]; // blue

// Character-level
msg.content.paragraphs[0].characters[0].font = "Courier-Bold";
msg.content.paragraphs[0].characters[0].color = [32768, 0, 40000]; // purple
```

### Style properties

| Property | Type | Set format | Read format | Notes |
|---|---|---|---|---|
| `font` | text | Font PostScript name | Font PostScript name | e.g. `"Helvetica"`, `"Helvetica-Bold"`, `"Helvetica-Oblique"`, `"Menlo-Regular"`, `"Courier"`, `"Georgia-Bold"` |
| `size` | number | Points (integer) | Points (integer) | e.g. `13`, `16`, `20` |
| `color` | RGBColor | 16-bit array `[R, G, B]` | Normalised 0–1 array | Set with values 0–65535; reads back as 0.0–1.0 floats |

**Color value inconsistency:** Colors are set using Apple's 16-bit RGB space
(0–65535) but read back as normalised floats (0.0–1.0):

```javascript
msg.content.paragraphs[0].color = [65535, 0, 0];  // set red
msg.content.paragraphs[0].color();                 // → [1, 0.149..., 0]
```

### Common font names

| Style | PostScript Name |
|---|---|
| Helvetica regular | `Helvetica` |
| Helvetica bold | `Helvetica-Bold` |
| Helvetica italic | `Helvetica-Oblique` |
| Helvetica bold italic | `Helvetica-BoldOblique` |
| Courier regular | `Courier` |
| Menlo regular | `Menlo-Regular` |
| Georgia bold | `Georgia-Bold` |
| System font | `-apple-system` (may not work in all contexts) |

### What does NOT work for styling

| Attempt | Outcome |
|---|---|
| `Mail.make({new: "richText", …})` | `Can't make or move that element into that container.` |
| `Mail.make({new: "attributeRun", …, at: msg.content})` | `Can't make or move that element into that container.` |
| `msg.content.attributeRuns()` (after mixed styles) | `Can't get object.` — reading attribute runs fails once multiple styles exist |
| `msg.content.attributeRuns[0].font()` | Same error — individual runs unreadable |

**Summary:** You can create `paragraph` elements and then style them (or their
words/characters) individually. You cannot create `richText` or `attributeRun`
elements directly, and you cannot reliably read `attributeRuns` after applying
mixed styles.

### Practical application: styled quote simulation

Using paragraph and character styling, a visually distinct quoted reply can be
constructed without relying on the inaccessible HTML layer:

```javascript
// Reply body — normal styling
Mail.make({
  new: "paragraph",
  withData: "Thanks for the update!\n",
  withProperties: { font: "Helvetica", size: 13, color: [0, 0, 0] },
  at: msg.content,
});

// Attribution line — italic, gray, smaller
Mail.make({
  new: "paragraph",
  withData: "On June 15, 2025, alice@example.com wrote:\n",
  withProperties: {
    font: "Helvetica-Oblique",
    size: 12,
    color: [32768, 32768, 32768],
  },
  at: msg.content,
});

// Quoted lines — blue, smaller
Mail.make({
  new: "paragraph",
  withData: "> Original message text here\n",
  withProperties: { font: "Helvetica", size: 12, color: [0, 0, 50000] },
  at: msg.content,
});

// Character-level: style the ">" marker differently
msg.content.paragraphs[2].characters[0].font = "Helvetica-Bold";
msg.content.paragraphs[2].characters[0].color = [32768, 0, 40000];
```

This produces a reply where:
- The reply text is in normal black body font
- The attribution ("On … wrote:") is gray and italic
- The quoted text is blue with a bold purple `>` marker

While not as polished as Mail.app's native HTML blockquote (which includes a
coloured vertical bar and indentation), this styled approach is significantly
better than plain `> ` text quoting and is fully controllable via the
scripting API.

## The Auto-Generated Quote Problem

When `reply()` is called (regardless of `openingWindow`), Mail.app:

1. Creates an `OutgoingMessage` with proper threading headers (`In-Reply-To`,
   `References`) and recipients.
2. If `openingWindow: true`, renders the quoted original message in the compose
   window's **HTML layer** (styled `<blockquote>` with
   `AppleOriginalContents` marker).

However, the **scripting `content` property remains empty**:

```javascript
const reply = targetMessage.reply({ openingWindow: true, replyToAll: false });
delay(3); // even after waiting

reply.content();                    // → "" (empty string)
reply.content.paragraphs().length;  // → 0
reply.content.characters().length;  // → 0
```

The quote is present in the HTML (verifiable via saving as draft and reading
`draftMessage.source()`), but it is invisible to the scripting API.

**Any write to `content` destroys the HTML quote.** This was confirmed
experimentally: after `Mail.make({new: "paragraph", …, at: reply.content})`,
the compose window shows only the inserted text; the quoted original is gone.

## Approaches Tested

The following table summarises every approach that was tested during the
investigation, along with the outcome:

| # | Approach | Outcome |
|---|---|---|
| 1 | `reply(openingWindow: false)` + `Mail.make paragraph at content` | ✅ Text is inserted. ❌ No quote (content starts empty, no HTML layer). |
| 2 | `reply(openingWindow: true)` + `Mail.make paragraph at content` | ✅ Text is inserted. ❌ HTML quote is destroyed by the write. |
| 3 | `reply(openingWindow: true)` + read `content()` + prepend + write back | ❌ `content()` returns `""` — nothing to prepend to. Write destroys quote. |
| 4 | `reply()` + `Mail.make({new: "richText", withData: text})` + assign | ❌ `Can't make or move that element into that container.` |
| 5 | Direct string assignment `reply.content = "text"` | ❌ `Can't convert types.` |
| 6 | System Events keystrokes (`SE.keystroke(…)`) | ❌ `osascript is not allowed to send keystrokes.` (without Accessibility permissions) |
| 7 | Clipboard paste via `SE.keystroke("v", {using: "command down"})` | ❌ Same — requires Accessibility permissions for the host process. |
| 8 | Close draft → find in `Mail.draftsMailbox()` → modify `Message.content` | ❌ `Message.content` is read-only. `Mail.make` silently does nothing. |
| 9 | Close draft → find in Drafts → `Mail.open()` → modify `OutgoingMessage.content` | ✅ API reports success. ❌ Change is not reflected in the compose window. |
| 10 | Varying delays (0–3 s) before writing to content | ❌ No effect — `content()` remains empty regardless of delay. |
| 11 | `reply(openingWindow: false)` + manual quote construction from `targetMessage.content()` | ✅ **Works.** Full control over content. Threading/headers preserved by `reply()`. |
| 12 | `Mail.make({new: "paragraph", withProperties: {font, size, color}, …})` | ✅ **Works.** Styled paragraphs render correctly in the compose window. |
| 13 | Post-creation paragraph/word/character `.font`/`.size`/`.color` assignment | ✅ **Works.** Granular styling confirmed at all three levels. |
| 14 | `Mail.make({new: "attributeRun", …})` | ❌ `Can't make or move that element into that container.` |
| 15 | `msg.content.attributeRuns()` after mixed styles | ❌ `Can't get object.` — attribute runs become unreadable. |
| 16 | `Mail.make paragraph at paragraphs.beginning` (paragraphs exist) | ❌ `Invalid key form.` — fails even when paragraphs are present. |
| 17 | `Mail.make paragraph at paragraphs[0].before` (paragraphs exist) | ❌ `Invalid index.` — JXA `.before` specifier not supported. |
| 18 | `Mail.make paragraph at characters.beginning` (characters exist) | ❌ `Invalid key form.` — no character-level insertion point works. |
| 19 | `paragraphs[0].set({to: "text"})` | ❌ `Can't convert types.` — cannot overwrite paragraph text. |

## The Working Approach

Since the auto-generated rich text quote cannot be preserved, the implemented
solution is:

1. **Call `reply()`** to create the `OutgoingMessage` — this sets up threading
   headers (`In-Reply-To`, `References`), recipient addresses, and the `Re: `
   subject prefix.
2. **Read the original message's plain text** via `targetMessage.content()`.
3. **Construct the quoted reply** ourselves (plain text or styled — see below).
4. **Insert into the OutgoingMessage** using `Mail.make`.

### Plain text approach (current implementation)

```javascript
const quotedReply =
  replyContent + "\n\n" +
  "On " + dateStr + ", " + originalSender + " wrote:\n\n" +
  "> " + originalContent.split("\n").join("\n> ");

Mail.make({
  new: "paragraph",
  withData: quotedReply,
  at: replyMessage.content,
});
```

### Styled approach (using RichText properties)

For a more visually polished result, paragraphs can be inserted with styling:

```javascript
// Reply body
Mail.make({
  new: "paragraph",
  withData: replyContent + "\n\n",
  withProperties: { font: "Helvetica", size: 13, color: [0, 0, 0] },
  at: replyMessage.content,
});

// Attribution
Mail.make({
  new: "paragraph",
  withData: "On " + dateStr + ", " + sender + " wrote:\n\n",
  withProperties: {
    font: "Helvetica-Oblique",
    size: 12,
    color: [32768, 32768, 32768],
  },
  at: replyMessage.content,
});

// Quoted lines (one paragraph per line for individual styling)
var lines = originalContent.split("\n");
for (var i = 0; i < lines.length; i++) {
  Mail.make({
    new: "paragraph",
    withData: "> " + lines[i] + "\n",
    withProperties: { font: "Helvetica", size: 12, color: [0, 0, 50000] },
    at: replyMessage.content,
  });
}
```

### Trade-offs

| Aspect | Auto-generated quote | Plain text `>` quote | Styled RichText quote |
|---|---|---|---|
| Formatting | Rich HTML with styled blockquote, coloured bar | Plain text with `> ` prefix | Coloured text, different fonts/sizes |
| Accessibility via API | ❌ Not accessible | ✅ Fully controlled | ✅ Fully controlled |
| Preservable on edit | ❌ No | ✅ Yes | ✅ Yes |
| Threading / headers | ✅ Set by `reply()` | ✅ Set by `reply()` | ✅ Set by `reply()` |
| Recipient handling | ✅ Automatic | ✅ Automatic | ✅ Automatic |
| Visual quality | ✅ Best | ⚠️ Basic | ✅ Good |
| Complexity | N/A (unusable) | Low | Moderate |

## Modifying Existing Drafts

When a draft is saved in the Drafts mailbox, it becomes a `Message` object.
Key findings:

- **`Message.content` is read-only** (marked `r/o` in the sdef). Writes via
  `Mail.make` appear to succeed at the API level but have no visible or
  persistent effect.
- **`Message.source` is read-only.** The full MIME source (including HTML) can
  be read but not written.
- **Re-opening a draft** via `Mail.open(draftMessage)` creates a new
  `OutgoingMessage` (findable via `Mail.outgoingMessages()`). Its `content`
  property is again empty, and writing to it does not affect the HTML layer
  (same disconnect as described above).
- **`OutgoingMessage.id` ≠ `Message.id`**: The integer IDs are in different
  namespaces. To correlate the two, match by `subject` or use timing (most
  recent draft).

### Finding the Drafts Mailbox

The `draftsMailbox()` method is a **top-level property** of the `Application` object (`Mail.draftsMailbox()`), not a property of individual `Account` objects. Attempting to call `account.draftsMailbox()` will throw an error in JXA.

Use the global `Application` property instead of searching by name:

```javascript
// ✅ Works regardless of locale (Drafts, Entwürfe, Brouillons, etc.)
// And represents the unified, top-level Drafts mailbox
const draftsMailbox = Mail.draftsMailbox();

// ❌ Throws an error: accounts do not have a draftsMailbox property
// const draftsBox = account.draftsMailbox();
```

To find drafts for a specific account, you must iterate through the global drafts and filter by the message's account name:

```javascript
const allDrafts = Mail.draftsMailbox().messages();
for (let i = 0; i < allDrafts.length; i++) {
  try {
    if (allDrafts[i].mailbox().account().name() === "Target Account") {
      // Found a draft for the target account
    }
  } catch (e) {
    // Some local drafts might not have an account name
  }
}
```

## Nested Mailbox Support

Mail.app supports hierarchical mailboxes (e.g., `Inbox > GitHub` or `Archive > 2024 > Q1`). Properly accessing these requires understanding JXA's mailbox navigation.

### The Problem

Nested mailboxes cannot be accessed with simple name lookups when using flat mailbox names. For example:

```javascript
// ❌ This fails for nested mailboxes
const mailbox = account.mailboxes["GitHub"];
```

This only finds top-level mailboxes named "GitHub", not `Inbox > GitHub`.

### The Solution: Mailbox Paths

Represent mailboxes as **JSON arrays** containing the full path:

```javascript
// Top-level mailbox
["Inbox"]

// Nested mailbox  
["Inbox", "GitHub"]

// Deeply nested
["Archive", "2024", "Q1"]
```

### Building Mailbox Paths

Walk up the mailbox hierarchy until reaching the account:

```javascript
function getMailboxPath(mailbox, accountName) {
  const path = [];
  let current = mailbox;

  while (current) {
    const name = current.name();
    
    // Stop at account level
    if (name === accountName) {
      break;
    }
    
    path.unshift(name);
    
    try {
      current = current.container();
    } catch (e) {
      break;
    }
  }
  
  return path;
}
```

**Key Points:**
- Stops when `mailbox.name() === accountName`
- Builds path from leaf to root, then reverses with `unshift()`
- Returns `["Inbox", "GitHub"]` not including account name

### Navigating to Nested Mailboxes

Simply chaining name lookups (e.g. `account.mailboxes["Inbox"]`) can fail due to localization, IMAP container folders (like `[Google Mail]`), and special characters.

Instead, use a robust hierarchical lookup with a fallback search. If direct name lookups or `whose()` filters fail at any step in the path, fallback to iterating over all mailboxes in the account, reconstructing their paths, and comparing them.

```javascript
function findMailboxByPath(account, targetPath) {
    if (!targetPath || targetPath.length === 0) return account;
    
    // 1. Try hierarchical traversal
    try {
        let current = account;
        for (let i = 0; i < targetPath.length; i++) {
            const part = targetPath[i];
            let next = null;
            // Attempt multiple lookup strategies
            try { next = current.mailboxes.whose({name: part})()[0]; } catch(e){}
            if (!next) { try { next = current.mailboxes.byName(part); next.name(); } catch(e){} }
            if (!next) { try { next = current.mailboxes[part]; next.name(); } catch(e){} }
            if (!next) throw new Error("not found");
            current = next;
        }
        return current;
    } catch(e) {}

    // 2. Fallback: iterate all mailboxes and match paths
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
                } catch (e) { break; }
            }
            if (path.length === targetPath.length) {
                let match = true;
                for (let j = 0; j < path.length; j++) {
                    if (path[j] !== targetPath[j]) { match = false; break; }
                }
                if (match) return mbx;
            }
        }
    } catch(e) {}
    
    return null;
}

// Parse mailboxPath from JSON
const mailboxPath = JSON.parse('["[Google Mail]", "Alle Nachrichten"]');

// Robustly find target mailbox
const targetMailbox = findMailboxByPath(account, mailboxPath);
```

**Why This Works:**
- JXA supports `whose()` queries which can sometimes find mailboxes where bracket syntax fails.
- The fallback correctly discovers nested folders hiding under invisible IMAP container names (like `[Google Mail]`).
- Bypasses issues with local vs. remote mailbox translations on macOS.

### Message Lookup with whose()

Once you have the target mailbox, use `whose()` for fast message lookup:

```javascript
const matchingMessages = targetMailbox.messages.whose({
  id: messageId,
})();
```

**Performance:**
- `whose()` runs in constant time (~0.3ms regardless of mailbox size)
- Searches only the specific mailbox (not recursive)
- 150x+ faster than manual iteration

### API Pattern

**get_selected_messages** output:
```json
{
  "mailbox": "GitHub",
  "mailboxPath": ["Inbox", "GitHub"],
  "account": "Exchange"
}
```

**get_message_content** input:
```bash
./get_message_content.js "Exchange" '["Inbox","GitHub"]' 66823
```

The `mailboxPath` is passed as a JSON array string.

## JXA Best Practices

Based on extensive testing and the [JXA documentation by Christian Kirsch](https://github.com/JXA-Cookbook/JXA-Cookbook):

### Object Specifier Dereferencing

In JXA, property access returns an **Object Specifier** (opaque pointer), not the actual value.

```javascript
// ❌ Wrong - returns Object Specifier
const name = mailbox.name;

// ✅ Correct - returns JavaScript string
const name = mailbox.name();

// Alternative syntax (equivalent)
const name = mailbox.name.get();
```

**Key Rule:** Always call properties as functions with `()` to get values.

### The .length Exception

The `.length` property works on **both** Object Specifiers and JavaScript arrays:

```javascript
// ✅ Both work
const count1 = mailbox.messages.length;
const count2 = mailbox.messages().length;
```

This is the **only** property that doesn't require dereferencing.

### Name Lookup Syntax

Access objects by name using bracket notation:

```javascript
// ✅ Direct name lookup (preferred)
const inbox = account.mailboxes["Inbox"];
const github = inbox.mailboxes["GitHub"];

// Also works without brackets (if no spaces/special chars)
const inbox = account.mailboxes.Inbox;

// ❌ Don't use loops for name-based lookup
const mailboxes = account.mailboxes();
for (let i = 0; i < mailboxes.length; i++) {
  if (mailboxes[i].name() === "Inbox") {
    // Found it (inefficient!)
  }
}
```

### Filtering with whose()

The `whose()` method provides SQL-like filtering with **constant-time performance**:

```javascript
// ✅ Fast filtering (constant time ~0.3ms)
const matches = mailbox.messages.whose({
  id: messageId,
})();

// With logical operators
const filtered = app.reminders.whose({
  _and: [
    {completed: true},
    {completionDate: {'>': yesterday}}
  ]
})();

// ❌ Slow filtering (linear time)
const messages = mailbox.messages();
for (let i = 0; i < messages.length; i++) {
  if (messages[i].id() === messageId) {
    // Found it (150x slower!)
  }
}
```

**Performance:**
- `whose()`: ~0.00064 ms per record (constant time)
- JavaScript filter: ~0.1 ms per record (linear time)
- Traditional loop: ~2 ms per record (linear time)

**For 500 records:**
- `whose()`: 0.32 ms
- JavaScript filter: 50 ms (150x slower)
- Traditional loop: 1000 ms (3000x slower)

### Converting Elements to Arrays

Elements (lists of objects) are **not** JavaScript arrays:

```javascript
// ❌ This fails - acc is an Object Specifier, not an array
const acc = app.accounts;
acc.forEach(a => console.log(a.name()));

// ✅ Dereference first to get JavaScript array
const acc = app.accounts();
acc.forEach(a => console.log(a.name()));
```

**Exception:** The `.length` property works on both.

### Enumeration Property Filtering

Filtering by enumeration properties requires special syntax:

```javascript
// ❌ This fails with "Types cannot be converted"
const pwAccounts = app.accounts.whose({
  authentication: "password"
})();

// ✅ Use _match with ObjectSpecifier()
const pwAccounts = app.accounts.whose({
  _match: [ObjectSpecifier().authentication, "password"]
})();
```

### Modern JavaScript Patterns

Use modern JavaScript in JXA scripts:

```javascript
// ✅ Use const/let instead of var
const Mail = Application("Mail");
const messages = mailbox.messages();

// ✅ Use template literals
const error = `Message ${messageId} not found in ${mailboxName}`;

// ✅ Use arrow functions
messages.forEach(msg => console.log(msg.subject()));

// ✅ Use for...of loops
for (const msg of messages) {
  console.log(msg.subject());
}
```

### Error Handling

```javascript
// ✅ Validate all arguments at the start
if (!accountName) {
  return JSON.stringify({
    success: false,
    error: "Account name is required"
  });
}

// ✅ Return structured JSON errors
try {
  const result = doSomething();
  return JSON.stringify({
    success: true,
    data: result
  });
} catch (e) {
  return JSON.stringify({
    success: false,
    error: e.toString()
  });
}

// ❌ Don't use empty catch blocks
try {
  const count = messages.length;
} catch (e) {} // Bad practice
```

### Date Handling

```javascript
// ✅ Use ISO format
const dateStr = message.dateReceived().toISOString();

// ❌ Avoid locale-specific formats
const dateStr = message.dateReceived().toLocaleString(); // Inconsistent
```

## Alternatives to JXA

The following alternatives were evaluated for interacting with Mail.app drafts.
None provide a strictly superior solution; each involves significant trade-offs.

### AppleScript

AppleScript uses the identical Apple Event bridge and Mail.app scripting
dictionary as JXA. The `content` property behaves exactly the same way — the
two-layer disconnect is a Mail.app architecture issue, not a JXA issue.
**No advantage over JXA.**

### IMAP APPEND

Bypass Mail.app entirely and push a fully-formed MIME message directly to the
Drafts folder via IMAP:

1. Construct a `multipart/alternative` MIME message with `text/plain` and
   `text/html` parts, including full rich text with styled `<blockquote>`.
2. Add proper threading headers (`In-Reply-To`, `References`).
3. IMAP `APPEND` the message to the Drafts folder.
4. Mail.app syncs and displays it as a draft.

| Aspect | Assessment |
|---|---|
| Rich text | ✅ Full HTML control |
| Threading | ✅ Manual headers |
| Credentials | ❌ Requires IMAP credentials or OAuth tokens |
| Exchange | ❌ Uses EWS/Graph API, not IMAP |
| Sync delay | ⚠️ User must wait for Mail.app to sync |
| Complexity | ⚠️ MIME construction is non-trivial |
| Reliability | ✅ High once working |

### Accessibility API + Clipboard Paste

The most promising approach for true rich text editing. Uses the JXA ObjC
bridge to put HTML on the clipboard, then sends Cmd+V via `CGEvent` to paste
into the compose window.

See [ObjC Bridge from JXA](#objc-bridge-from-jxa) below for implementation
details.

| Aspect | Assessment |
|---|---|
| Rich text | ✅ Full HTML via clipboard paste |
| Threading | ✅ Via `reply()` |
| Credentials | None required |
| Permissions | ❌ mail-mcp needs Accessibility access |
| Reliability | ⚠️ Medium — depends on timing, frontmost app, cursor position |
| Setup burden | ⚠️ Each MCP host app needs separate Accessibility grant |

### MailKit / Mail Extensions

Apple's official replacement for Mail plugins (deprecated in macOS 12+).
The `MEComposeSessionHandler` protocol provides:

- `additionalHeaders(for:)` — add custom headers
- Recipient annotation

**It does NOT provide** any API to read or modify the compose message body.
**Not useful for content editing.**

### `.eml` File + `open` Command

Construct a complete `.eml` file with HTML content and threading headers,
then `open -a Mail draft.eml`.

**Result:** Mail.app opens `.eml` files in **read-only viewer mode**, not as
an editable draft. **Not viable.**

### `mailto:` URL Scheme

```
mailto:user@example.com?subject=Re:%20Hello&body=My%20reply%20text
```

**Limitations:**
- Body is plain text only (URL-encoded)
- No HTML support
- No threading headers (`In-Reply-To`, `References`)
- Not a reply, just a new message

**Not viable for replies.**

### Summary Matrix

| Approach | Rich Text | Threading | Credentials | Fragility | Feasibility |
|---|---|---|---|---|---|
| **JXA + styled paragraphs** | ⚠️ Limited | ✅ via `reply()` | None | Low | ✅ Production-ready |
| **JXA + plain `>` quoting** | ❌ Plain only | ✅ via `reply()` | None | Low | ✅ Production-ready |
| **IMAP APPEND** | ✅ Full HTML | ✅ Manual headers | ❌ Required | Low | ⚠️ Moderate effort |
| **Accessibility-based reply** | ✅ Full HTML | ✅ via `reply()` | None | 🔴 High | ✅ Supported |
| **MailKit Extension** | ❌ No body API | N/A | None | Low | ❌ Not possible |
| **mailto: URL** | ❌ Plain only | ❌ No threading | None | Low | ❌ Wrong tool |
| **AppleScript** | ❌ Same as JXA | ✅ | None | Low | ❌ No advantage |

## Accessibility-based Reply Tools

The `create_reply_draft` and `replace_reply_draft` tools implement the [Clipboard HTML paste strategy](#clipboard-html-paste-strategy).

### How it works

1. It calls `targetMessage.reply({openingWindow: true})`. This preserves the native HTML quote and threading.
2. It waits for the window to appear and focus.
3. It puts the reply content on the system clipboard (NSPasteboard).
4. It uses the `CoreGraphics` API (`CGEvent`) to simulate a `Cmd+V` keystroke.

### Requirements

- **Accessibility Permissions:** The **mail-mcp** binary must be granted Accessibility access in **System Settings -> Privacy & Security -> Accessibility**. Granting access to the binary directly is recommended for better security.
- **Interactive:** A Mail.app window will briefly pop up and focus during the operation.

## ObjC Bridge from JXA

JXA has a built-in Objective-C bridge via `ObjC.import()` that provides access
to Objective-C classes and C functions — **but only within the `osascript`
process**, not in Mail.app's process.

### What the ObjC bridge CAN do

| Capability | API | Notes |
|---|---|---|
| Clipboard with UTIs | `NSPasteboard` | Put HTML, RTF, images on clipboard |
| Data manipulation | `NSData`, `NSString` | Construct MIME, encode HTML |
| Subprocesses | `NSTask` | Shell out to helper tools |
| Synthetic input | `CGEvent` | Keystrokes, mouse clicks (requires Accessibility) |
| App management | `NSWorkspace` | Activate Mail.app, open URLs |
| File operations | `NSFileManager` | Read/write local files |
| UI inspection | `AXUIElement` | Walk other apps' UI trees (requires Accessibility) |

### What the ObjC bridge CANNOT do

**Access objects in another process's memory.** You cannot:

- Get a reference to Mail.app's `WKWebView` instance
- Call `evaluateJavaScript:` on Mail.app's WebView
- Read Mail.app's internal data structures

Cross-process interaction is limited to Apple Events (JXA scripting),
Accessibility API (`AXUIElement`), pasteboard, and `CGEvent`.

### Clipboard HTML paste strategy

The most promising use of the ObjC bridge for rich text editing:

```javascript
ObjC.import('AppKit');
ObjC.import('CoreGraphics');

// 1. Check Accessibility permissions
ObjC.import('ApplicationServices');
if (!$.AXIsProcessTrusted()) {
  // Fail gracefully with instructions
}

// 2. Put HTML on the clipboard
var pb = $.NSPasteboard.generalPasteboard;
pb.clearContents;
pb.setStringForType(
  $('<div><p>Reply text</p><blockquote>Quoted text</blockquote></div>'),
  $('public.html')
);

// 3. Send Cmd+V via CGEvent (keycode 9 = 'v')
var keyDown = $.CGEventCreateKeyboardEvent(null, 9, true);
var keyUp = $.CGEventCreateKeyboardEvent(null, 9, false);
$.CGEventSetFlags(keyDown, $.kCGEventFlagMaskCommand);
$.CGEventSetFlags(keyUp, $.kCGEventFlagMaskCommand);
$.CGEventPost($.kCGHIDEventTap, keyDown);
$.CGEventPost($.kCGHIDEventTap, keyUp);
```

**Prerequisite:** The **mail-mcp** binary
must be granted Accessibility access in **System Settings → Privacy & Security
→ Accessibility**. Granting access to the binary directly is recommended for better security. This is a manual setup step.

### AXUIElement for UI inspection

`AXUIElement` can walk Mail.app's UI element tree, but for WebView elements
it only exposes:

- `AXRole: AXWebArea`
- `AXValue`: **plain text** rendering (not HTML)
- `AXSelectedText`, `AXSelectedTextRange`: cursor/selection info

It does **not** expose the HTML DOM. Even with Accessibility permissions, you
cannot inject HTML through `AXUIElement` — only the clipboard paste strategy
provides that capability.

### Keystroke mechanisms compared

All three mechanisms for sending synthetic input require the same Accessibility
permission:

| Mechanism | API Layer | Works from `osascript`? |
|---|---|---|
| System Events `keystroke` | AppleScript/JXA | ⚠️ Only if host process is trusted |
| `CGEventPost` | CoreGraphics (C) | ⚠️ Same — process must be trusted |
| `NSEvent` | AppKit (ObjC) | ❌ Per-process only — cannot target Mail.app |

### Compiled Swift helper approach

An alternative to doing everything in JXA is to bundle a small Swift CLI tool:

```swift
import AppKit
import ApplicationServices

let html = readLine()!
let pb = NSPasteboard.general
pb.clearContents()
pb.setString(html, forType: .html)

let mail = NSRunningApplication.runningApplications(
    withBundleIdentifier: "com.apple.mail"
).first!
mail.activate()

let src = CGEventSource(stateID: .hidSystemState)
let keyDown = CGEvent(keyboardEventSource: src, virtualKey: 0x09, keyDown: true)!
keyDown.flags = .maskCommand
keyDown.post(tap: .cghidEventTap)
```

**Advantage:** A code-signed binary can be granted Accessibility permissions
independently of the MCP host application.

**Disadvantage:** Adds a compiled binary dependency, build complexity, and the
same timing/focus fragility as the pure JXA approach.

## Useful Properties & Helpers

| Expression | Returns | Notes |
|---|---|---|
| `Mail.draftsMailbox()` | Mailbox | Top-level Drafts mailbox, locale-independent |
| `Mail.inbox()` | Mailbox | Top-level Inbox |
| `Mail.outgoingMessages()` | Array | All currently open `OutgoingMessage` objects |
| `targetMessage.content()` | string | Plain text content of a received message |
| `targetMessage.source()` | string | Full MIME source (headers + HTML + attachments) |
| `replyMessage.toRecipients()` | Array | Recipient objects on an outgoing message |
| `replyMessage.subject()` | string | Subject line |
| `replyMessage.visible()` | boolean | Whether the compose window is shown |
| `replyMessage.close({saving: "yes"})` | — | Save draft and close compose window |
| `msg.content.paragraphs[n].font()` | string | Font PostScript name of paragraph n |
| `msg.content.paragraphs[n].size()` | number | Point size of paragraph n |
| `msg.content.paragraphs[n].color()` | Array | RGB colour (normalised 0–1 on read) |

## Implications for New Message Composition

The same `Mail.make paragraph at content` pattern applies when creating new
messages from scratch:

```javascript
const msg = Mail.make({
  new: "outgoingMessage",
  withProperties: {
    subject: "Hello",

  },
});

// Add a recipient
Mail.make({
  new: "toRecipient",
  withProperties: { address: "user@example.com" },
  at: msg.toRecipients,
});

// Set styled content
Mail.make({
  new: "paragraph",
  withData: "Message body here.\n",
  withProperties: {
    font: "Helvetica",
    size: 13,
    color: [0, 0, 0],
  },
  at: msg.content,
});
```

The same constraints apply: `content` cannot be set via direct assignment, and
any HTML content added by Mail.app (e.g. signatures rendered via the Signature
setting) may be overwritten by the first write to `content`.

## Recommendations

1. **Always use `Mail.make({new: "paragraph", …, at: msg.content})`** to set
   content on an `OutgoingMessage`. No other method works reliably. This
   always **appends** — there is no working insertion point for prepending
   (see [Prepending: no working insertion point](#prepending-no-working-insertion-point)).

2. **Build content in the correct order** — reply body first, attribution
   line second, quoted lines last. Because only appending works, the order
   of `Mail.make` calls determines the final paragraph order. You cannot
   rearrange paragraphs after insertion.

3. **Use `withProperties: {font, size, color}`** when creating paragraphs to
   get styled rich text. This is significantly better than plain text quoting
   and avoids the need for Accessibility permissions or external tools.

4. **Use post-creation styling** (`paragraphs[n].font = …`) for granular
   control at the word or character level after initial paragraph creation.

5. **Always use `msg.content()` (with parentheses)** when reading content.
   Without parentheses you get the RichText specifier, which throws
   `Can't convert types` when used as a string.

6. **Construct quotes manually** from `targetMessage.content()`. The
   auto-generated rich text quote is inaccessible and will be destroyed by any
   content write.

7. **Use `reply()` / `forward()` for threading**, even when constructing
   content yourself. These methods set the correct `In-Reply-To`, `References`,
   and recipient headers that maintain the email thread.

8. **Use `Mail.draftsMailbox()`** to find drafts. Do not search by mailbox
   name — it varies by locale. Remember that `draftsMailbox()` is a top-level `Application` property, not an `Account` property.

9. **Do not attempt to modify `Message.content`** (saved messages in
   mailboxes). It is read-only. API calls may appear to succeed but have no
   effect.

10. **Avoid template literals in JXA scripts.** Use string concatenation (`+`)
   instead — template literals can behave unexpectedly in some JXA contexts.

11. **Prefer JXA RichText styling over Accessibility-based approaches** for
    production use. The clipboard paste strategy (ObjC bridge + CGEvent)
    provides full HTML support but introduces fragility, timing dependencies,
    and a mandatory Accessibility permission grant that varies per MCP host.

---

## References

- [Mail.sdef.md](../Mail.sdef.md) — Complete Mail.app scripting dictionary
- [RICH_TEXT_HANDLING.md](RICH_TEXT_HANDLING.md) — RichText API details and
  common pitfalls
- [create_reply_draft.js](../internal/tools/scripts/create_reply_draft.js) —
  Implementation for creating replies via Accessibility API
- [replace_reply_draft.js](../internal/tools/scripts/replace_reply_draft.js) —
  Implementation for updating reply drafts via Accessibility API
- [Apple JXA Release Notes](https://developer.apple.com/library/archive/releasenotes/InterapplicationCommunication/RN-JavaScriptForAutomation/)
- [Mac Automation Scripting Guide](https://developer.apple.com/library/archive/documentation/LanguagesUtilities/Conceptual/MacAutomationScriptingGuide/)
