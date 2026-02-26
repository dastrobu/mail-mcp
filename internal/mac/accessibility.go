//go:build darwin

package mac

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework Cocoa

#include <stdlib.h>
#include <unistd.h>
#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <Cocoa/Cocoa.h>

// IsProcessTrustedAndPrompt checks if the process is trusted and will prompt the
// user with the system dialog if the permission has not yet been granted.
static bool IsProcessTrustedAndPrompt() {
    const void* keys[] = {kAXTrustedCheckOptionPrompt};
    const void* values[] = {kCFBooleanTrue};
    CFDictionaryRef options = CFDictionaryCreate(NULL, keys, values, 1, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
    if (!options) return false;

    bool trusted = AXIsProcessTrustedWithOptions(options);
    CFRelease(options);
    return trusted;
}

// Get PID of running application by bundle ID
static pid_t GetPIDForBundleID(const char* bundleIDString) {
    @autoreleasepool {
        NSString *bundleID = [NSString stringWithUTF8String:bundleIDString];
        NSArray *apps = [NSRunningApplication runningApplicationsWithBundleIdentifier:bundleID];
        if ([apps count] == 0) {
            return 0;
        }
        NSRunningApplication *app = [apps firstObject];
        return [app processIdentifier];
    }
}

// Activate application by PID
static void ActivateApp(pid_t pid) {
    @autoreleasepool {
        NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
        if (app) {
            [app activateWithOptions:NSApplicationActivateAllWindows];
        }
    }
}

// Set clipboard content with both HTML and plain text representations.
static bool SetClipboardContent(const char* htmlStr, const char* plainStr) {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];

        bool success = false;
        if (htmlStr) {
            NSString *html = [NSString stringWithUTF8String:htmlStr];
            if (html) {
                [pb setString:html forType:NSPasteboardTypeHTML];
				[pb setString:html forType:@"Apple HTML pasteboard type"];
                success = true;
            }
        }
        if (plainStr) {
            NSString *plain = [NSString stringWithUTF8String:plainStr];
            if (plain) {
                [pb setString:plain forType:NSPasteboardTypeString];
                success = true;
            }
        }
        return success;
    }
}

// Simulate Cmd+V keystroke
static void SimulateCmdV() {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!source) return;

    CGKeyCode vKey = 9; // 'v' key
    CGEventRef keyDown = CGEventCreateKeyboardEvent(source, vKey, true);
    if (keyDown) {
        CGEventSetFlags(keyDown, kCGEventFlagMaskCommand);
        CGEventPost(kCGHIDEventTap, keyDown);
        CFRelease(keyDown);
    }

    CGEventRef keyUp = CGEventCreateKeyboardEvent(source, vKey, false);
    if (keyUp) {
        CGEventSetFlags(keyUp, kCGEventFlagMaskCommand);
        CGEventPost(kCGHIDEventTap, keyUp);
        CFRelease(keyUp);
    }
    CFRelease(source);
}

// Recursively find and focus the message body (AXWebArea)
static bool FocusMessageBodyInElement(AXUIElementRef element, int depth) {
    if (depth > 10) return false;

    CFArrayRef children = NULL;
    if (AXUIElementCopyAttributeValue(element, kAXChildrenAttribute, (CFTypeRef *)&children) != kAXErrorSuccess || !children) {
        return false;
    }

    CFIndex count = CFArrayGetCount(children);
    bool found = false;
    for (CFIndex i = 0; i < count; i++) {
        AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(children, i);
        CFStringRef role = NULL;

        if (AXUIElementCopyAttributeValue(child, kAXRoleAttribute, (CFTypeRef *)&role) == kAXErrorSuccess && role) {
            if (CFStringCompare(role, CFSTR("AXWebArea"), 0) == kCFCompareEqualTo) {
                AXUIElementSetAttributeValue(child, kAXFocusedAttribute, kCFBooleanTrue);
                found = true;
            }
            CFRelease(role);
        }
        if (found) break;
        if (FocusMessageBodyInElement(child,
 depth + 1)) {
            found = true;
            break;
        }
    }
    CFRelease(children);
    return found;
}

// WaitFocusAndPaste performs an atomic operation to wait for a window, focus its body, and paste.
static bool WaitFocusAndPaste(pid_t pid, const char* expectedTitleCStr, int timeout_ms) {
    AXUIElementRef app = AXUIElementCreateApplication(pid);
    if (!app) return false;

	bool success = false;
	@autoreleasepool {
		NSString *expectedTitle = [NSString stringWithUTF8String:expectedTitleCStr];
		CFAbsoluteTime deadline = CFAbsoluteTimeGetCurrent() + (timeout_ms / 1000.0);
		while (CFAbsoluteTimeGetCurrent() < deadline) {
			AXUIElementRef window = NULL;
			if (AXUIElementCopyAttributeValue(app, kAXFocusedWindowAttribute, (CFTypeRef *)&window) == kAXErrorSuccess && window) {
				CFStringRef title = NULL;
				if (AXUIElementCopyAttributeValue(window, kAXTitleAttribute, (CFTypeRef *)&title) == kAXErrorSuccess && title) {
					if (expectedTitle.length == 0 || [(NSString *)title rangeOfString:expectedTitle].location != NSNotFound) {
						if (FocusMessageBodyInElement(window, 0)) {
							// Crucial delay to allow Mail.app's UI thread to catch up after focus before pasting.
							usleep(50 * 1000);
							SimulateCmdV();
							success = true;
						}
					}
					CFRelease(title);
				}
				CFRelease(window);
			}
			if (success) break;
			usleep(100 * 1000);
		}
	}
    CFRelease(app);
    return success;
}

*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unsafe"
)

// EnsureAccessibility checks for accessibility permissions.
func EnsureAccessibility() error {
	if bool(C.IsProcessTrustedAndPrompt()) {
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		executable = "mail-mcp"
	}
	executableName := filepath.Base(executable)
	return fmt.Errorf(`accessibility permission is required. Please follow these steps:
1. In System Settings > Privacy & Security > Accessibility, find '%s' and ensure it's enabled.
2. If it is already enabled but failing, the binary was likely updated. You must remove the stale entry.
3. Run the tool again to trigger a new macOS permission prompt.
4. IMPORTANT: After granting permission, you MUST restart the service.

If you are running the service via Homebrew, execute this command:
brew services restart %s

Otherwise, execute this command:
%s launchd restart`, executableName, executableName, executableName)
}

// GetMailPID returns the PID of Mail.app.
func GetMailPID() int {
	cBundleID := C.CString("com.apple.mail")
	defer C.free(unsafe.Pointer(cBundleID))
	return int(C.GetPIDForBundleID(cBundleID))
}

// SetClipboard sets the system clipboard content.
func SetClipboard(htmlContent *string, plainContent string) error {
	var cHTML *C.char
	if htmlContent != nil {
		cHTML = C.CString(*htmlContent)
		defer C.free(unsafe.Pointer(cHTML))
	}
	cPlain := C.CString(plainContent)
	defer C.free(unsafe.Pointer(cPlain))

	if !bool(C.SetClipboardContent(cHTML, cPlain)) {
		return fmt.Errorf("C.SetClipboardContent failed")
	}
	return nil
}

// PasteIntoWindow performs a "fire-and-forget" paste operation. It finds a window,
// focuses its body, and simulates a paste command. It assumes success if the
// commands execute without error, but does not verify the final content.
func PasteIntoWindow(ctx context.Context, pid int, expectedTitle string, timeout time.Duration, htmlContent *string, plainContent string) error {
	// 1. Set clipboard content.
	if err := SetClipboard(htmlContent, plainContent); err != nil {
		return fmt.Errorf("failed to set clipboard for pasting: %w", err)
	}
	time.Sleep(100 * time.Millisecond) // Pause to ensure clipboard is ready.

	// 2. Activate app.
	C.ActivateApp(C.pid_t(pid))

	// 3. Perform the atomic wait, focus, and paste operation.
	cExpectedTitle := C.CString(expectedTitle)
	defer C.free(unsafe.Pointer(cExpectedTitle))
	timeoutMs := C.int(timeout.Milliseconds())

	if bool(C.WaitFocusAndPaste(C.pid_t(pid), cExpectedTitle, timeoutMs)) {
		return nil // SUCCESS
	}

	return fmt.Errorf("failed to find, focus, or paste into window with title '%s' within %v", expectedTitle, timeout)
}
