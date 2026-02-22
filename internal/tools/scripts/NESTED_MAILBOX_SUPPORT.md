# Nested Mailbox Support

All tools must support hierarchical mailboxes (e.g., `Inbox > GitHub`).

## Mailbox Path Representation

Use JSON arrays for mailbox paths:
- Top-level: `["Inbox"]`
- Nested: `["Inbox", "GitHub"]`
- Deeply nested: `["Archive", "2024", "Q1"]`

## Building Mailbox Paths (JXA)

```javascript
function getMailboxPath(mailbox, accountName) {
  const path = [];
  let current = mailbox;

  while (current) {
    const name = current.name();
    if (name === accountName) break;
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

## Navigating to Nested Mailboxes (JXA)

```javascript
// Parse mailboxPath from JSON
const mailboxPath = JSON.parse(mailboxPathStr);

// Navigate using chained lookups
let targetMailbox = account.mailboxes[mailboxPath[0]];
for (let i = 1; i < mailboxPath.length; i++) {
  targetMailbox = targetMailbox.mailboxes[mailboxPath[i]];
}
```

## Go Integration

**Input Type:**
```go
type ToolInput struct {
    Account     string   `json:"account" jsonschema:"Account name"`
    MailboxPath []string `json:"mailboxPath" jsonschema:"Mailbox path array"`
    ID          int      `json:"id" jsonschema:"Message ID"`
}
```

**Execution:**
```go
mailboxPathJSON, err := json.Marshal(input.MailboxPath)
if err != nil {
    return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
}

data, err := jxa.Execute(ctx, script,
    input.Account,
    string(mailboxPathJSON),
    fmt.Sprintf("%d", input.ID))
```
