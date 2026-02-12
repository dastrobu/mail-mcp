package jxa

import (
	"context"
	"fmt"
	"time"
)

// startupCheckScript is a minimal JXA script that verifies Mail.app is accessible
const startupCheckScript = `
function run(argv) {
    try {
        const Mail = Application('Mail');
        Mail.includeStandardAdditions = true;

        // Check if Mail is running
        if (!Mail.running()) {
            return JSON.stringify({
                success: false,
                error: 'Mail.app is not running. Please start Mail.app and try again.'
            });
        }

        // Try to access accounts to verify we can interact with Mail
        const accounts = Mail.accounts();
        const accountCount = accounts.length;

        // Retrieve all Mail.app properties
        // Some properties may fail on certain macOS/Mail versions, so wrap each in try-catch
        const properties = {};

        // Helper function to safely get property
        function safeGet(name, getter) {
            try {
                properties[name] = getter();
            } catch (e) {
                properties[name] = null; // Property not available
            }
        }

        safeGet('alwaysBccMyself', () => Mail.alwaysBccMyself());
        safeGet('alwaysCcMyself', () => Mail.alwaysCcMyself());
        safeGet('applicationVersion', () => Mail.applicationVersion());
        safeGet('fetchInterval', () => Mail.fetchInterval());
        safeGet('backgroundActivityCount', () => Mail.backgroundActivityCount());
        safeGet('chooseSignatureWhenComposing', () => Mail.chooseSignatureWhenComposing());
        safeGet('colorQuotedText', () => Mail.colorQuotedText());
        safeGet('defaultMessageFormat', () => Mail.defaultMessageFormat());
        safeGet('downloadHtmlAttachments', () => Mail.downloadHtmlAttachments());
        safeGet('expandGroupAddresses', () => Mail.expandGroupAddresses());
        safeGet('fixedWidthFont', () => Mail.fixedWidthFont());
        safeGet('fixedWidthFontSize', () => Mail.fixedWidthFontSize());
        safeGet('includeAllOriginalMessageText', () => Mail.includeAllOriginalMessageText());
        safeGet('quoteOriginalMessage', () => Mail.quoteOriginalMessage());
        safeGet('checkSpellingWhileTyping', () => Mail.checkSpellingWhileTyping());
        safeGet('levelOneQuotingColor', () => Mail.levelOneQuotingColor());
        safeGet('levelTwoQuotingColor', () => Mail.levelTwoQuotingColor());
        safeGet('levelThreeQuotingColor', () => Mail.levelThreeQuotingColor());
        safeGet('messageFont', () => Mail.messageFont());
        safeGet('messageFontSize', () => Mail.messageFontSize());
        safeGet('messageListFont', () => Mail.messageListFont());
        safeGet('messageListFontSize', () => Mail.messageListFontSize());
        safeGet('newMailSound', () => Mail.newMailSound());
        safeGet('shouldPlayOtherMailSounds', () => Mail.shouldPlayOtherMailSounds());
        safeGet('sameReplyFormat', () => Mail.sameReplyFormat());
        safeGet('selectedSignature', () => Mail.selectedSignature());
        safeGet('fetchesAutomatically', () => Mail.fetchesAutomatically());
        safeGet('highlightSelectedConversation', () => Mail.highlightSelectedConversation());
        safeGet('useAddressCompletion', () => Mail.useAddressCompletion());
        safeGet('useFixedWidthFont', () => Mail.useFixedWidthFont());
        safeGet('primaryEmail', () => Mail.primaryEmail());

        return JSON.stringify({
            success: true,
            data: {
                running: true,
                accountCount: accountCount,
                version: Mail.version(),
                properties: properties
            }
        });
    } catch (e) {
        return JSON.stringify({
            success: false,
            error: 'Failed to access Mail.app: ' + e.toString()
        });
    }
}
`

// StartupCheck performs a basic connectivity check with Mail.app
// It verifies that Mail.app is running and accessible via JXA
// Returns the Mail.app properties if successful
//
// Common errors:
// - "signal: killed" - macOS is blocking automation; grant permissions in System Settings
// - "Mail.app is not running" - Start Mail.app before running the server
func StartupCheck(ctx context.Context) (map[string]any, error) {
	// Use a timeout for the startup check
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := Execute(ctx, startupCheckScript)
	if err != nil {
		return nil, fmt.Errorf("Mail.app startup check failed: %w", err)
	}

	// Verify we got valid data back
	data, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("startup check returned unexpected data type: %T", result)
	}

	running, ok := data["running"].(bool)
	if !ok || !running {
		return nil, fmt.Errorf("Mail.app is not properly accessible")
	}

	return data, nil
}
