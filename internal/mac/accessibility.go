package mac

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework Cocoa

#include <stdlib.h>
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

// Get focused window title for a PID
// Returns newly allocated C string that must be freed by caller, or NULL on failure
static char* GetFocusedWindowTitle(pid_t pid) {
    AXUIElementRef app = AXUIElementCreateApplication(pid);
    if (!app) return NULL;

    AXUIElementRef window = NULL;
    AXError err = AXUIElementCopyAttributeValue(app, kAXFocusedWindowAttribute, (CFTypeRef *)&window);
    CFRelease(app);

    if (err != kAXErrorSuccess || !window) {
        return NULL;
    }

    CFStringRef title = NULL;
    err = AXUIElementCopyAttributeValue(window, kAXTitleAttribute, (CFTypeRef *)&title);
    CFRelease(window);

    if (err != kAXErrorSuccess || !title) {
        return NULL;
    }

    // Convert CFString to C string
    CFIndex length = CFStringGetLength(title);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char *buffer = (char *)malloc(maxSize);
    if (buffer) {
        if (!CFStringGetCString(title, buffer, maxSize, kCFStringEncodingUTF8)) {
            free(buffer);
            buffer = NULL;
        }
    }

    CFRelease(title);
    return buffer;
}

// Recursively find and focus the message body (AXWebArea)
// Returns true if found and focused
static bool FocusMessageBodyInElement(AXUIElementRef element, int depth) {
    if (depth > 10) return false; // Avoid deep recursion

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
            // Check if this is the WebArea (message body)
            if (CFStringCompare(role, CFSTR("AXWebArea"), 0) == kCFCompareEqualTo) {
                // Found it! Set focus.
                // Note: Sometimes we need to focus the group container or the web area itself.
                // Let's try focusing the WebArea.
                AXUIElementSetAttributeValue(child, kAXFocusedAttribute, kCFBooleanTrue);
                found = true;
            }
            CFRelease(role);
        }

        if (found) break;

        // Recursively search children
        if (FocusMessageBodyInElement(child, depth + 1)) {
            found = true;
            break;
        }
    }

    CFRelease(children);
    return found;
}

// Focus the message body of the frontmost window of the given PID
static bool FocusMessageBody(pid_t pid) {
    AXUIElementRef app = AXUIElementCreateApplication(pid);
    if (!app) return false;

    AXUIElementRef window = NULL;
    if (AXUIElementCopyAttributeValue(app, kAXFocusedWindowAttribute, (CFTypeRef *)&window) != kAXErrorSuccess || !window) {
        CFRelease(app);
        return false;
    }

    // Traverse window children to find AXWebArea
    bool result = FocusMessageBodyInElement(window, 0);

    CFRelease(window);
    CFRelease(app);
    return result;
}

// Set clipboard content
static bool SetClipboardContent(const char* contentStr, bool isHTML) {
    @autoreleasepool {
        if (!contentStr) return false;
        NSString *content = [NSString stringWithUTF8String:contentStr];
        if (!content) return false;

        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];

        if (isHTML) {
            [pb setString:content forType:NSPasteboardTypeHTML];
            [pb setString:content forType:NSPasteboardTypeString];
        } else {
            [pb setString:content forType:NSPasteboardTypeString];
        }
        return true;
    }
}

// Simulate Cmd+V keystroke
static void SimulateCmdV() {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!source) return;

    // key code 9 is 'v' (ANSI V)
    CGKeyCode vKey = 9;

    // Create KeyDown event
    CGEventRef keyDown = CGEventCreateKeyboardEvent(source, vKey, true);
    if (keyDown) {
        CGEventSetFlags(keyDown, kCGEventFlagMaskCommand);
        CGEventPost(kCGHIDEventTap, keyDown);
        CFRelease(keyDown);
    }

    // Create KeyUp event
    CGEventRef keyUp = CGEventCreateKeyboardEvent(source, vKey, false);
    if (keyUp) {
        CGEventSetFlags(keyUp, kCGEventFlagMaskCommand);
        CGEventPost(kCGHIDEventTap, keyUp);
        CFRelease(keyUp);
    }

    CFRelease(source);
}
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"
)

// EnsureAccessibility checks for accessibility permissions and prompts the user to grant them if missing.
// It triggers the native macOS permission prompt if necessary.
func EnsureAccessibility() error {
	// This will check for permission and, if not yet granted, will trigger the
	// native macOS prompt asking the user to grant it. If the user denies the
	// prompt, this will return false, and we will show our detailed error.
	if bool(C.IsProcessTrustedAndPrompt()) {
		return nil
	}

	executable, err := os.Executable()
	if err != nil {
		// Fallback to a generic name if we can't get the executable path
		executable = "apple-mail-mcp"
	}
	executableName := filepath.Base(executable)

	// macOS Permission Edge Case (Updates):
	// macOS tracks the binary's exact cdhash (Code Directory hash). When the binary is recompiled,
	// the cdhash changes, and macOS silently invalidates the old Accessibility permission.
	// A launcher script won't bypass this, as the system checks the actual compiled binary making
	// the Accessibility API calls. Thus, the user must manually remove the stale entry.
	return fmt.Errorf(`accessibility permission is required. Please follow these steps:
1. In System Settings > Privacy & Security > Accessibility, find '%s' and ensure it's enabled.
2. If it is already enabled but failing, the binary was likely updated. macOS revokes permissions when an unsigned tool is recompiled. You must select it and click the minus (-) button to remove the stale entry.
3. Run the tool again to trigger a new macOS permission prompt.
4. IMPORTANT: After granting the new permission, you MUST restart the service for it to take effect.

Execute this command:
%s launchd restart
(Or 'brew services restart apple-mail-mcp' if installed via Homebrew)`, executableName, executableName)
}

// GetMailPID returns the PID of Mail.app if running, or 0 if not found
func GetMailPID() int {
	cBundleID := C.CString("com.apple.mail")
	defer C.free(unsafe.Pointer(cBundleID))
	return int(C.GetPIDForBundleID(cBundleID))
}

// GetFocusedWindowTitle returns the title of the focused window for the given PID
func GetFocusedWindowTitle(pid int) (string, error) {
	cTitle := C.GetFocusedWindowTitle(C.pid_t(pid))
	if cTitle == nil {
		return "", fmt.Errorf("failed to get focused window title (accessibility permission or focus issue)")
	}
	defer C.free(unsafe.Pointer(cTitle))
	return C.GoString(cTitle), nil
}

// SetClipboard sets the system clipboard content.
// If isHTML is true, the content is treated as HTML/Rich Text.
func SetClipboard(content string, isHTML bool) error {
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cContent))
	success := C.SetClipboardContent(cContent, C.bool(isHTML))
	if !success {
		return fmt.Errorf("failed to set clipboard content")
	}
	return nil
}

// SimulatePaste simulates Cmd+V keystroke using CoreGraphics
func SimulatePaste() {
	C.SimulateCmdV()
}

// PasteContent sets the clipboard and simulates a Cmd+V keystroke
func PasteContent(content string, isHTML bool) error {
	if err := SetClipboard(content, isHTML); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	SimulatePaste()
	return nil
}

// FocusBody attempts to find and focus the message body (WebArea) of the frontmost window
func FocusBody(pid int) error {
	if !bool(C.FocusMessageBody(C.pid_t(pid))) {
		return fmt.Errorf("failed to find or focus message body")
	}
	return nil
}

// ActivateApp brings the application with the given PID to the foreground
func ActivateApp(pid int) {
	C.ActivateApp(C.pid_t(pid))
}

// WaitForWindowFocus waits for a window belonging to the given PID to gain focus.
// If expectedTitle is non-empty, it waits until the focused window title contains that string.
// It returns the title of the focused window or an error if the timeout is reached.
func WaitForWindowFocus(ctx context.Context, pid int, expectedTitle string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Ensure the app is activated
	ActivateApp(pid)

	var lastTitle string
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return lastTitle, fmt.Errorf("timed out waiting for window focus (last title: %q)", lastTitle)
			}

			title, err := GetFocusedWindowTitle(pid)
			if err == nil && title != "" {
				lastTitle = title
				// If no specific title is expected, any focused window title is enough
				if expectedTitle == "" {
					return title, nil
				}
				// If a title is expected, check if it matches
				if strings.Contains(title, expectedTitle) {
					return title, nil
				}
			}
		}
	}
}
