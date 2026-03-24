package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmDangerous prompts the user to type the resource name to confirm deletion.
func ConfirmDangerous(resourceType, resourceName string, confirmed bool) bool {
	if confirmed {
		return true
	}

	fmt.Fprintf(os.Stderr, "\n⚠ This will permanently delete %s '%s'. This cannot be undone.\n", resourceType, resourceName)
	fmt.Fprintf(os.Stderr, "Type the %s name to confirm: ", resourceType)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == resourceName
	}
	return false
}

// ConfirmDestructive prompts for destructive operations (project/repo delete).
func ConfirmDestructive(resourceType, resourceName string, confirmed, understands bool) bool {
	if confirmed && understands {
		return true
	}

	fmt.Fprintf(os.Stderr, "\n🛑 DESTRUCTIVE: This will permanently delete %s '%s' and ALL its contents.\n", resourceType, resourceName)
	fmt.Fprintf(os.Stderr, "   This CANNOT be undone.\n")
	fmt.Fprintf(os.Stderr, "Type '%s' to confirm: ", resourceName)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == resourceName
	}
	return false
}
