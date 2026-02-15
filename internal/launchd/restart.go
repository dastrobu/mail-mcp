package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

// Restart uses 'launchctl kickstart' to stop and start the launchd service
// without modifying the existing plist file. This is the recommended, atomic way
// to restart a service while respecting any manual user configuration.
func Restart() error {
	// First, check if the service plist even exists. If not, we can't restart it.
	plistPath := PlistPath()
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("❌ launchd service file not found at %s. Cannot restart. Please run 'launchd create' first", plistPath)
	}

	fmt.Println("♻️  Restarting launchd service using existing configuration...")

	// Get the current user's UID to construct the correct launchd target service specifier.
	// For user agents, the format is 'gui/<uid>/<service-label>'.
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("❌ failed to get current user: %w", err)
	}
	uid := currentUser.Uid
	serviceTarget := fmt.Sprintf("gui/%s/%s", uid, Label)

	// The 'launchctl kickstart -k' command is the modern, correct way to stop and
	// immediately restart a service. It's atomic and ensures the service is
	// killed if running before starting a new instance from the existing plist.
	cmd := exec.Command("launchctl", "kickstart", "-k", serviceTarget)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ failed to restart launchd service with kickstart: %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✅  Launchd service restarted successfully.")
	return nil
}
