package keychain

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const Account = "credmux"

func Service(profile, platform string) string {
	return fmt.Sprintf("credmux-%s-%s", profile, platform)
}

// Add stores a secret in the macOS login keychain. Existing entries with the
// same service+account are removed first so the call is idempotent.
func Add(profile, platform, secret string) error {
	_ = Delete(profile, platform)
	svc := Service(profile, platform)
	cmd := exec.Command("security", "add-generic-password",
		"-s", svc, "-a", Account, "-w", secret)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain add %s: %w (%s)", svc, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Find retrieves the plaintext secret for the given profile/platform pair.
func Find(profile, platform string) (string, error) {
	svc := Service(profile, platform)
	cmd := exec.Command("security", "find-generic-password",
		"-s", svc, "-a", Account, "-w")
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("keychain find %s: %s", svc, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("keychain find %s: %w", svc, err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// Delete removes a keychain entry. Missing entries return an error which the
// caller can ignore; Add() does so to stay idempotent.
func Delete(profile, platform string) error {
	svc := Service(profile, platform)
	cmd := exec.Command("security", "delete-generic-password",
		"-s", svc, "-a", Account)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keychain delete %s: %w", svc, err)
	}
	return nil
}
