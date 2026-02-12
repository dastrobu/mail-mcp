# Apple Mail Scripting Dictionary

This document describes the scripting interface for Apple Mail on macOS, generated from Mail.sdef.

---

## Table of Contents

- [Standard Suite](#standard-suite)
- [Text Suite](#text-suite)
- [Mail Suite](#mail-suite)
- [Mail Framework](#mail-framework)

---

## Standard Suite

Common classes and commands for all applications.

### Methods

#### `open`
Open a document.

**Parameters:**
- `file or list of file` — The file(s) to be opened

**Returns:** `Document or list of Document` — The opened document(s)

#### `close`
Close a document.

**Parameters:**
- `specifier` — The document(s) or window(s) to close
- `[saving]` — Should changes be saved before closing? (`"yes"` / `"no"` / `"ask"`)
- `[savingIn]` — The file in which to save the document, if so

#### `save`
Save a document.

**Parameters:**
- `specifier` — The document(s) or window(s) to save
- `[in]` — The file in which to save the document
- `[as]` — The file format to use (`"native format"`)

#### `print`
Print a document.

**Parameters:**
- `list of file or specifier` — The file(s), document(s), or window(s) to be printed
- `[withProperties]` — The print settings to use
- `[printDialog]` — Should the application show the print dialog?

#### `quit`
Quit the application.

**Parameters:**
- `[saving]` — Should changes be saved before quitting? (`"yes"` / `"no"` / `"ask"`)

#### `count`
Return the number of elements of a particular class within an object.

**Parameters:**
- `specifier` — The objects to be counted

**Returns:** `integer` — The count

#### `exists`
Verify that an object exists.

**Parameters:**
- `any` — The object(s) to check

**Returns:** `boolean` — Did the object(s) exist?

#### `make`
Create a new object.

**Parameters:**
- `new` — The class of the new object
- `[at]` — The location at which to insert the object
- `[withData]` — The initial contents of the object
- `[withProperties]` — The initial values for properties of the object

**Returns:** `specifier` — The new object

#### `delete`
Delete an object.

**Parameters:**
- `specifier` — The object(s) to delete

#### `duplicate`
Copy an object.

**Parameters:**
- `specifier` — The object(s) to copy
- `[to]` — The location for the new copy or copies
- `[withProperties]` — Properties to set in the new copy or copies right away

#### `move`
Move an object to a new location.

**Parameters:**
- `specifier` — The object(s) to move
- `to` — The new location for the object(s)

### Objects

#### `Application`
The application's top-level scripting object. [see also Mail]

**Elements:**
- `documents`
- `windows`

**Properties:**
- `name` (text, r/o) — The name of the application
- `frontmost` (boolean, r/o) — Is this the active application?
- `version` (text, r/o) — The version number of the application

**Methods:** `open`, `print`, `quit`

#### `Document`
A document.

**Contained by:** `application`

**Properties:**
- `name` (text, r/o) — Its name
- `modified` (boolean, r/o) — Has it been modified since the last save?
- `file` (file, r/o) — Its location on disk, if it has one

**Methods:** `close`, `print`, `save`

#### `Window`
A window.

**Contained by:** `application`

**Properties:**
- `name` (text, r/o) — The title of the window
- `id` (integer, r/o) — The unique identifier of the window
- `index` (integer) — The index of the window, ordered front to back
- `bounds` (rectangle) — The bounding rectangle of the window
- `closeable` (boolean, r/o) — Does the window have a close button?
- `miniaturizable` (boolean, r/o) — Does the window have a minimize button?
- `miniaturized` (boolean) — Is the window minimized right now?
- `resizable` (boolean, r/o) — Can the window be resized?
- `visible` (boolean) — Is the window visible right now?
- `zoomable` (boolean, r/o) — Does the window have a zoom button?
- `zoomed` (boolean) — Is the window zoomed right now?
- `document` (Document, r/o) — The document whose contents are displayed in the window

**Methods:** `close`, `print`, `save`

#### `PrintSettings`
Print settings object.

**Properties:**
- `copies` (integer) — The number of copies of a document to be printed
- `collating` (boolean) — Should printed copies be collated?
- `startingPage` (integer) — The first page of the document to be printed
- `endingPage` (integer) — The last page of the document to be printed
- `pagesAcross` (integer) — Number of logical pages laid across a physical page
- `pagesDown` (integer) — Number of logical pages laid out down a physical page
- `requestedPrintTime` (date) — The time at which the desktop printer should print the document
- `errorHandling` (`"standard"` / `"detailed"`) — How errors are handled
- `faxNumber` (text) — For fax number
- `targetPrinter` (text) — For target printer

### Enumerations

#### `save options`
- `"yes"` — Save the file
- `"no"` — Do not save the file
- `"ask"` — Ask the user whether or not to save the file

#### `printing error handling`
- `"standard"` — Standard PostScript error handling
- `"detailed"` — Print a detailed report of PostScript errors

#### `saveable file format`
- `"native format"` — Native format

---

## Text Suite

A set of basic classes for text processing.

### Objects

#### `RGBColor`
RGB color object.

#### `RichText`
Rich (styled) text.

**Elements:**
- `paragraphs`
- `words`
- `characters`
- `attributeRuns`
- `attachments`

**Properties:**
- `color` (RGBColor) — The color of the first character
- `font` (text) — The name of the font of the first character
- `size` (number) — The size in points of the first character

**Note:** To set RichText content programmatically in JXA, insert paragraphs using `Mail.make({new: "paragraph", withData: "text", at: message.content})`. You cannot directly assign a string to a RichText property (fails with "Can't convert types").

#### `Attachment` [inherits from RichText]
Represents an inline text attachment. This class is used mainly for make commands.

**Contained by:** `richText`, `paragraphs`, `words`, `characters`, `attributeRuns`

**Properties:**
- `fileName` (file) — The file for the attachment

#### `Paragraph`
This subdivides the text into paragraphs.

**Contains:** `words`, `characters`, `attributeRuns`, `attachments`

**Contained by:** `richText`, `attributeRuns`

**Properties:**
- `color` (RGBColor) — The color of the first character
- `font` (text) — The name of the font of the first character
- `size` (number) — The size in points of the first character

#### `Word`
This subdivides the text into words.

**Contains:** `characters`, `attributeRuns`, `attachments`

**Contained by:** `richText`, `paragraphs`, `attributeRuns`

**Properties:**
- `color` (RGBColor) — The color of the first character
- `font` (text) — The name of the font of the first character
- `size` (number) — The size in points of the first character

#### `Character`
This subdivides the text into characters.

**Contains:** `attributeRuns`, `attachments`

**Contained by:** `richText`, `paragraphs`, `words`, `attributeRuns`

**Properties:**
- `color` (RGBColor) — The color of the character
- `font` (text) — The name of the font of the character
- `size` (number) — The size in points of the character

#### `AttributeRun`
This subdivides the text into chunks that all have the same attributes.

**Contains:** `paragraphs`, `words`, `characters`, `attachments`

**Contained by:** `richText`, `paragraphs`, `words`, `characters`

**Properties:**
- `color` (RGBColor) — The color of the first character
- `font` (text) — The name of the font of the first character
- `size` (number) — The size in points of the first character

---

## Mail Suite

Classes and commands for the Mail application.

### Methods

#### `checkForNewMail`
Triggers a check for email.

**Parameters:**
- `[for]` — Specify the account that you wish to check for mail

#### `extractNameFrom`
Command to get the full name out of a fully specified email address.  
Example: `"John Doe <jdoe@example.com>"` returns `"John Doe"`

**Parameters:**
- `text` — Fully formatted email address

**Returns:** `text` — The full name

#### `extractAddressFrom`
Command to get just the email address of a fully specified email address.  
Example: `"John Doe <jdoe@example.com>"` returns `"jdoe@example.com"`

**Parameters:**
- `text` — Fully formatted email address

**Returns:** `text` — The email address

#### `forward`
Creates a forwarded message.

**Parameters:**
- `Message` — The message to forward
- `[openingWindow]` — Whether the window for the forwarded message is shown. Default is to not show the window.

**Returns:** `OutgoingMessage` — The message to be forwarded

#### `geturl`
Opens a mailto URL.

**Parameters:**
- `text` — The mailto URL

#### `importMailMailbox`
Imports a mailbox created by Mail.

**Parameters:**
- `at` — The mailbox or folder of mailboxes to import

#### `mailto`
Opens a mailto URL.

**Parameters:**
- `text` — The mailto URL

#### `performMailActionWithMessages`
Script handler invoked by rules and menus that execute AppleScripts. The direct parameter of this handler is a list of messages being acted upon.

**Parameters:**
- `list of Message` — The message being acted upon
- `[inMailboxes]` — If the script is being executed by the user selecting an item in the scripts menu, this argument will specify the mailboxes that are currently selected
- `[forRule]` — If the script is being executed by a rule action, this argument will be the rule being invoked

#### `redirect`
Creates a redirected message.

**Parameters:**
- `Message` — The message to redirect
- `[openingWindow]` — Whether the window for the redirected message is shown. Default is to not show the window.

**Returns:** `OutgoingMessage` — The redirected message

#### `reply`
Creates a reply message.

**Parameters:**
- `Message` — The message to reply to
- `[openingWindow]` — Whether the window for the reply message is shown. Default is to not show the window.
- `[replyToAll]` — Whether to reply to all recipients. Default is to reply to the sender only.

**Returns:** `OutgoingMessage` — The reply message

#### `send`
Sends a message.

**Parameters:**
- `OutgoingMessage` — The message to send

**Returns:** `boolean` — True if sending was successful, false if not

#### `synchronize`
Command to trigger synchronizing of an IMAP account with the server.

**Parameters:**
- `with` — The account to synchronize

### Objects

#### `OutgoingMessage`
A new email message.

**Contains:** `bccRecipients`, `ccRecipients`, `recipients`, `toRecipients`

**Contained by:** `application`

**Properties:**
- `sender` (text) — The sender of the message
- `subject` (text) — The subject of the message
- `content` (RichText) — The contents of the message. **IMPORTANT:** This is a RichText object, not a plain string. To set content, you must insert paragraphs using `Mail.make({new: "paragraph", withData: "your text", at: message.content})`. Direct string assignment will fail with "Can't convert types" error. Note: any write to this property destroys any auto-generated HTML content (e.g. quoted original in replies).
- `visible` (boolean) — Controls whether the message window is shown on the screen. The default is false
- `messageSignature` (Signature or missing value) — The signature of the message
- `id` (integer, r/o) — The unique identifier of the message

**Methods:** `save`, `close`, `send`

#### `Application` [see also Standard Suite]
Mail's top level scripting object.

**Elements:**
- `accounts`
- `popAccounts`
- `imapAccounts`
- `icloudAccounts`
- `smtpServers`
- `outgoingMessages`
- `mailboxes`
- `messageViewers`
- `rules`
- `signatures`

**Properties:**
- `alwaysBccMyself` (boolean) — Indicates whether you will be included in the Bcc: field of messages which you are composing
- `alwaysCcMyself` (boolean) — Indicates whether you will be included in the Cc: field of messages which you are composing
- `selection` (list of Message, r/o) — List of messages that the user has selected
- `applicationVersion` (text, r/o) — The build number of the application
- `fetchInterval` (integer) — The interval (in minutes) between automatic fetches of new mail, -1 means to use an automatically determined interval
- `backgroundActivityCount` (integer, r/o) — Number of background activities currently running in Mail, according to the Activity window
- `chooseSignatureWhenComposing` (boolean) — Indicates whether user can choose a signature directly in a new compose window
- `colorQuotedText` (boolean) — Indicates whether quoted text should be colored
- `defaultMessageFormat` (`"plain format"` / `"rich format"`) — Default format for messages being composed or message replies
- `downloadHtmlAttachments` (boolean) — Indicates whether images and attachments in HTML messages should be downloaded and displayed
- `draftsMailbox` (Mailbox, r/o) — The top level Drafts mailbox
- `expandGroupAddresses` (boolean) — Indicates whether group addresses will be expanded when entered into the address fields of a new compose message
- `fixedWidthFont` (text) — Font for plain text messages, only used if 'use fixed width font' is set to true
- `fixedWidthFontSize` (real) — Font size for plain text messages, only used if 'use fixed width font' is set to true
- `inbox` (Mailbox, r/o) — The top level In mailbox
- `includeAllOriginalMessageText` (boolean) — Indicates whether all of the original message will be quoted or only the text you have selected (if any)
- `quoteOriginalMessage` (boolean) — Indicates whether the text of the original message will be included in replies
- `checkSpellingWhileTyping` (boolean) — Indicates whether spelling will be checked automatically in messages being composed
- `junkMailbox` (Mailbox, r/o) — The top level Junk mailbox
- `levelOneQuotingColor` (`"blue"` / `"green"` / `"orange"` / `"other"` / `"purple"` / `"red"` / `"yellow"`) — Color for quoted text with one level of indentation
- `levelTwoQuotingColor` (`"blue"` / `"green"` / `"orange"` / `"other"` / `"purple"` / `"red"` / `"yellow"`) — Color for quoted text with two levels of indentation
- `levelThreeQuotingColor` (`"blue"` / `"green"` / `"orange"` / `"other"` / `"purple"` / `"red"` / `"yellow"`) — Color for quoted text with three levels of indentation
- `messageFont` (text) — Font for messages (proportional font)
- `messageFontSize` (real) — Font size for messages (proportional font)
- `messageListFont` (text) — Font for message list
- `messageListFontSize` (real) — Font size for message list
- `newMailSound` (text) — Name of new mail sound or 'None' if no sound is selected
- `outbox` (Mailbox, r/o) — The top level Out mailbox
- `shouldPlayOtherMailSounds` (boolean) — Indicates whether sounds will be played for various things such as when a messages is sent or if no mail is found when manually checking for new mail or if there is a fetch error
- `sameReplyFormat` (boolean) — Indicates whether replies will be in the same text format as the message to which you are replying
- `selectedSignature` (text) — Name of current selected signature (or 'randomly', 'sequentially', or 'none')
- `sentMailbox` (Mailbox, r/o) — The top level Sent mailbox
- `fetchesAutomatically` (boolean) — Indicates whether mail will automatically be fetched at a specific interval
- `highlightSelectedConversation` (boolean) — Indicates whether messages in conversations should be highlighted in the Mail viewer window when not grouped
- `trashMailbox` (Mailbox, r/o) — The top level Trash mailbox
- `useAddressCompletion` (boolean) — This always returns true, and setting it doesn't do anything (deprecated)
- `useFixedWidthFont` (boolean) — Should fixed-width font be used for plain text messages?
- `primaryEmail` (text, r/o) — The user's primary email address

**Methods:** `checkForNewMail`, `importMailMailbox`, `synchronize`

#### `MessageViewer`
Represents the object responsible for managing a viewer window.

**Contains:** `messages`

**Contained by:** `application`

**Properties:**
- `draftsMailbox` (Mailbox, r/o) — The top level Drafts mailbox
- `inbox` (Mailbox, r/o) — The top level In mailbox
- `junkMailbox` (Mailbox, r/o) — The top level Junk mailbox
- `outbox` (Mailbox, r/o) — The top level Out mailbox
- `sentMailbox` (Mailbox, r/o) — The top level Sent mailbox
- `trashMailbox` (Mailbox, r/o) — The top level Trash mailbox
- `sortColumn` (ViewerColumns) — The column that is currently sorted in the viewer
- `sortedAscending` (boolean) — Whether the viewer is sorted ascending or not
- `mailboxListVisible` (boolean) — Controls whether the list of mailboxes is visible or not
- `previewPaneIsVisible` (boolean) — Controls whether the preview pane of the message viewer window is visible or not
- `visibleColumns` (list of ViewerColumns) — List of columns that are visible. The subject column and the message status column will always be visible
- `id` (integer, r/o) — The unique identifier of the message viewer
- `visibleMessages` (list of Message) — List of messages currently being displayed in the viewer
- `selectedMessages` (list of Message) — List of messages currently selected
- `selectedMailboxes` (list of Mailbox) — List of mailboxes currently selected in the list of mailboxes
- `window` (Window) — The window for the message viewer

#### `Signature`
Email signatures.

**Contained by:** `application`

**Properties:**
- `content` (text) — Contents of email signature. If there is a version with fonts and/or styles, that will be returned over the plain text version
- `name` (text) — Name of the signature

### Enumerations

#### `DefaultMessageFormat`
- `"plain format"` — Plain Text
- `"rich format"` — Rich Text

#### `QuotingColor`
- `"blue"` — Blue
- `"green"` — Green
- `"orange"` — Orange
- `"other"` — Other
- `"purple"` — Purple
- `"red"` — Red
- `"yellow"` — Yellow

#### `ViewerColumns`
- `"attachments column"` — Column containing the number of attachments a message contains
- `"message color"` — Used to indicate sorting should be done by color
- `"date received column"` — Column containing the date a message was received
- `"date sent column"` — Column containing the date a message was sent
- `"flags column"` — Column containing the flags of a message
- `"from column"` — Column containing the sender's name
- `"mailbox column"` — Column containing the name of the mailbox or account a message is in
- `"message status column"` — Column indicating a messages status (read, unread, replied to, forwarded, etc)
- `"number column"` — Column containing the number of a message in a mailbox
- `"size column"` — Column containing the size of a message
- `"subject column"` — Column containing the subject of a message
- `"to column"` — Column containing the recipients of a message
- `"date last saved column"` — Column containing the date a draft message was saved

---

## Mail Framework

Classes and commands for the Mail framework.

### Objects

#### `Message`
An email message.

**Contains:** `bccRecipients`, `ccRecipients`, `recipients`, `toRecipients`, `headers`, `mailAttachments`

**Contained by:** `messageViewers`, `mailboxes`

**Properties:**
- `id` (integer, r/o) — The unique identifier of the message
- `allHeaders` (text, r/o) — All the headers of the message
- `backgroundColor` (HighlightColors) — The background color of the message
- `mailbox` (Mailbox) — The mailbox in which this message is filed
- `content` (RichText, r/o) — Contents of an email message
- `dateReceived` (date, r/o) — The date a message was received
- `dateSent` (date, r/o) — The date a message was sent
- `deletedStatus` (boolean) — Indicates whether the message is deleted or not
- `flaggedStatus` (boolean) — Indicates whether the message is flagged or not
- `flagIndex` (integer) — The flag on the message, or -1 if the message is not flagged
- `junkMailStatus` (boolean) — Indicates whether the message has been marked junk or evaluated to be junk by the junk mail filter
- `readStatus` (boolean) — Indicates whether the message is read or not
- `messageId` (text, r/o) — The unique message ID string
- `source` (text, r/o) — Raw source of the message
- `replyTo` (text, r/o) — The address that replies should be sent to
- `messageSize` (integer, r/o) — The size (in bytes) of a message
- `sender` (text, r/o) — The sender of the message
- `subject` (text, r/o) — The subject of the message
- `wasForwarded` (boolean, r/o) — Indicates whether the message was forwarded or not
- `wasRedirected` (boolean, r/o) — Indicates whether the message was redirected or not
- `wasRepliedTo` (boolean, r/o) — Indicates whether the message was replied to or not

**Methods:** `open`, `bounce`, `forward`, `redirect`, `reply`

#### `Recipient`
An email recipient.

**Contains:** `bccRecipients`, `ccRecipients`, `recipients`, `toRecipients`

**Contained by:** `messages`, `outgoingMessages`

**Properties:**
- `address` (text, r/o) — The email address
- `name` (text, r/o) — The full name

#### `BccRecipient` [inherits from Recipient]
A Bcc recipient.

#### `CcRecipient` [inherits from Recipient]
A Cc recipient.

#### `ToRecipient` [inherits from Recipient]
A To recipient.

#### `Container`
A mailbox container.

**Contains:** `mailboxes`

**Contained by:** `application`, `accounts`, `mailboxes`

**Properties:**
- `name` (text, r/o) — The name of the container

#### `Header`
An email header.

**Contained by:** `messages`

**Properties:**
- `content` (text, r/o) — Contents of the header
- `name` (text, r/o) — Name of the header value

#### `MailAttachment`
An email attachment.

**Contained by:** `messages`

**Properties:**
- `name` (text, r/o) — Name of the attachment
- `mimeType` (text, r/o) — MIME type of the attachment E.g. text/plain
- `fileSize` (integer, r/o) — Approximate size in bytes
- `downloaded` (boolean, r/o) — Indicates whether the attachment has been downloaded
- `id` (text, r/o) — The unique identifier of the attachment

#### `Account`
A Mail account for receiving messages (POP/IMAP). To create a new receiving account, use the 'pop account', 'imap account', and 'iCloud account' objects.

**Contains:** `mailboxes`

**Contained by:** `application`

**Properties:**
- `deliveryAccount` (SmtpServer or missing value) — The delivery account used when sending mail from this account
- `name` (text) — The name of an account
- `id` (text, r/o) — The unique identifier of the account
- `password` (text) — Password for this account. Can be set, but not read via scripting
- `authentication` (AuthenticationMethod) — Preferred authentication scheme for account
- `accountType` (AccountType, r/o) — The type of an account
- `emailAddresses` (list of text) — The list of email addresses configured for an account
- `fullName` (text) — The users full name configured for an account
- `emptyJunkMessagesFrequency` (integer) — Number of days before junk messages are deleted (0 = delete on quit, -1 = never delete)
- `emptyTrashFrequency` (integer) — Number of days before messages in the trash are permanently deleted (0 = delete on quit, -1 = never delete)
- `emptyJunkMessagesOnQuit` (boolean) — Indicates whether the messages in the junk messages mailboxes will be deleted on quit
- `emptyTrashOnQuit` (boolean) — Indicates whether the messages in deleted messages mailboxes will be permanently deleted on quit
- `enabled` (boolean) — Indicates whether the account is enabled or not
- `userName` (text) — The user name used to connect to an account
- `accountDirectory` (file, r/o) — The directory where the account stores things on disk
- `port` (integer) — The port used to connect to an account
- `serverName` (text) — The host name used to connect to an account
- `moveDeletedMessagesToTrash` (boolean) — Indicates whether messages that are deleted will be moved to the trash mailbox
- `usesSsl` (boolean) — Indicates whether SSL is enabled for this receiving account

#### `ImapAccount` [inherits from Account]
An IMAP email account.

**Contained by:** `application`

**Properties:**
- `compactMailboxesWhenClosing` (boolean) — Indicates whether an IMAP mailbox is automatically compacted when you quit Mail or switch to another mailbox
- `messageCaching` (MessageCachingPolicy) — Message caching setting for this account
- `storeDraftsOnServer` (boolean) — Indicates whether drafts will be stored on the IMAP server
- `storeJunkMailOnServer` (boolean) — Indicates whether junk mail will be stored on the IMAP server
- `storeSentMessagesOnServer` (boolean) — Indicates whether sent messages will be stored on the IMAP server
- `storeDeletedMessagesOnServer` (boolean) — Indicates whether deleted messages will be stored on the IMAP server

#### `ICloudAccount` [inherits from ImapAccount > Account]
An iCloud or MobileMe email account.

**Synonyms:** MacAccount, MobileMeAccount

**Contained by:** `application`

#### `PopAccount` [inherits from Account]
A POP email account.

**Contained by:** `application`

**Properties:**
- `bigMessageWarningSize` (integer) — If message size (in bytes) is over this amount, Mail will prompt you asking whether you want to download the message (-1 = do not prompt)
- `delayedMessageDeletionInterval` (integer) — Number of days before messages that have been downloaded will be deleted from the server (0 = delete immediately after downloading)
- `deleteMailOnServer` (boolean) — Indicates whether POP account deletes messages on the server after downloading
- `deleteMessagesWhenMovedFromInbox` (boolean) — Indicates whether messages will be deleted from the server when moved from your POP inbox

#### `SmtpServer`
An SMTP account (for sending email).

**Contained by:** `application`

**Properties:**
- `name` (text, r/o) — The name of an account
- `password` (text) — Password for this account. Can be set, but not read via scripting
- `accountType` (AccountType, r/o) — The type of an account
- `authentication` (AuthenticationMethod) — Preferred authentication scheme for account
- `enabled` (boolean) — Indicates whether the account is enabled or not
- `userName` (text) — The user name used to connect to an account
- `port` (integer) — The port used to connect to an account
- `serverName` (text) — The host name used to connect to an account
- `usesSsl` (boolean) — Indicates whether SSL is enabled for this receiving account

#### `Mailbox`
A mailbox that holds messages.

**Contains:** `mailboxes`, `messages`

**Contained by:** `application`, `accounts`, `mailboxes`

**Properties:**
- `name` (text) — The name of a mailbox
- `unreadCount` (integer, r/o) — The number of unread messages in the mailbox
- `account` (Account, r/o)
- `container` (Mailbox, r/o)

#### `Rule`
Class for message rules.

**Contains:** `ruleConditions`

**Contained by:** `application`

**Properties:**
- `colorMessage` (HighlightColors) — If rule matches, apply this color
- `deleteMessage` (boolean) — If rule matches, delete message
- `forwardText` (text) — If rule matches, prepend this text to the forwarded message. Set to empty string to include no prepended text
- `forwardMessage` (text) — If rule matches, forward message to this address, or multiple addresses, separated by commas. Set to empty string to disable this action
- `markFlagged` (boolean) — If rule matches, mark message as flagged
- `markFlagIndex` (integer) — If rule matches, mark message with the specified flag. Set to -1 to disable this action
- `markRead` (boolean) — If rule matches, mark message as read
- `playSound` (text) — If rule matches, play this sound (specify name of sound or path to sound)
- `redirectMessage` (text) — If rule matches, redirect message to this address or multiple addresses, separate by commas. Set to empty string to disable this action
- `replyText` (text) — If rule matches, reply to message and prepend with this text. Set to empty string to disable this action
- `runScript` (file or missing value) — If rule matches, run this compiled AppleScript file. Set to empty string to disable this action
- `allConditionsMustBeMet` (boolean) — Indicates whether all conditions must be met for rule to execute
- `copyMessage` (Mailbox) — If rule matches, copy to this mailbox
- `moveMessage` (Mailbox) — If rule matches, move to this mailbox
- `highlightTextUsingColor` (boolean) — Indicates whether the color will be used to highlight the text or background of a message in the message list
- `enabled` (boolean) — Indicates whether the rule is enabled
- `name` (text) — Name of rule
- `shouldCopyMessage` (boolean) — Indicates whether the rule has a copy action
- `shouldMoveMessage` (boolean) — Indicates whether the rule has a move action
- `stopEvaluatingRules` (boolean) — If rule matches, stop rule evaluation for this message

#### `RuleCondition`
Class for conditions that can be attached to a single rule.

**Contained by:** `rules`

**Properties:**
- `expression` (text) — Rule expression field
- `header` (text) — Rule header key
- `qualifier` (RuleQualifier) — Rule qualifier
- `ruleType` (RuleType) — Rule type

### Enumerations

#### `AccountType`
- `"pop"` — POP account
- `"smtp"` — SMTP account
- `"imap"` — IMAP account
- `"iCloud"` — iCloud account

#### `AuthenticationMethod`
- `"password"` — Clear text password
- `"apop"` — APOP
- `"kerberos 5"` — Kerberos V5 (GSSAPI)
- `"ntlm"` — NTLM
- `"md5"` — CRAM-MD5
- `"external"` — External authentication
- `"Apple token"` — Apple token
- `"none"` — None

#### `HighlightColors`
- `"blue"` — Blue
- `"gray"` — Gray
- `"green"` — Green
- `"none"` — None
- `"orange"` — Orange
- `"other"` — Other
- `"purple"` — Purple
- `"red"` — Red
- `"yellow"` — Yellow

#### `MessageCachingPolicy`
- `"all messages but omit attachments"` — All messages but omit attachments
- `"all messages and their attachments"` — All messages and their attachments

#### `RuleQualifier`
- `"begins with value"` — Begins with value
- `"does contain value"` — Does contain value
- `"does not contain value"` — Does not contain value
- `"ends with value"` — Ends with value
- `"equal to value"` — Equal to value
- `"less than value"` — Less than value
- `"greater than value"` — Greater than value
- `"none"` — None

#### `RuleType`
- `"account"` — Account
- `"any recipient"` — Any recipient
- `"cc header"` — Cc header
- `"header key"` — An arbitrary header key
- `"message content"` — Message content
- `"message is junk mail"` — Message is or is not junk mail
- `"sender"` — Sender
- `"subject header"` — Subject header
- `"to header"` — To header
- `"to or cc header"` — To or Cc header
- `"sender is in my contacts"` — Sender is or is not in my contacts
- `"sender is vip"` — Sender is or is not VIP
- `"sender is member of group"` — Sender is or is not member of group
- `"attachment type"` — Attachment type

---

## Notes

- Properties marked with `r/o` are read-only
- Optional parameters are shown in square brackets `[parameter]`
- Enum values are shown in quotes with forward slashes: `"value1"` / `"value2"`
- Inheritance is indicated with `[inh. Parent]`

---

*This documentation was generated from Mail.sdef*
