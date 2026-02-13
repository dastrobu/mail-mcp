package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// unload unloads the service if it's currently loaded
func unload() error {
	if !IsLoaded() {
		return nil
	}

	plistPath := PlistPath()
	cmd := exec.Command("launchctl", "unload", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ failed to unload service: %w", err)
	}

	// Wait a moment for the service to unload
	time.Sleep(time.Second)
	return nil
}

// Remove removes the launchd service
func Remove() error {
	fmt.Println("Removing launchd service...")

	// Unload service
	if IsLoaded() {
		if err := unload(); err != nil {
			return err
		}
		fmt.Println("Service unloaded")
	}

	// Remove plist file
	plistPath := PlistPath()
	if _, err := os.Stat(plistPath); err == nil {
		if err := os.Remove(plistPath); err != nil {
			return fmt.Errorf("❌ failed to remove plist file: %w", err)
		}
		fmt.Printf("Removed plist file: %s\n", plistPath)
	}

	fmt.Println("✅ Service successfully removed")
	return nil
}
