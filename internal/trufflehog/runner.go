package trufflehog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

var ErrNotInstalled = errors.New("trufflehog not found in PATH (brew install trufflehog)")

type Finding struct {
	DetectorName string
	Raw          string
	Verified     bool
	File         string
	Line         int64
}

type rawOutput struct {
	DetectorName   string `json:"DetectorName"`
	Raw            string `json:"Raw"`
	Verified       bool   `json:"Verified"`
	SourceMetadata struct {
		Data struct {
			Filesystem struct {
				File string `json:"file"`
				Line int64  `json:"line"`
			} `json:"Filesystem"`
		} `json:"Data"`
	} `json:"SourceMetadata"`
}

// Scan runs `trufflehog filesystem <path> --json --no-update` and returns
// every finding emitted on stdout. exit code 183 (secrets found) is treated
// as success; any other non-zero exit is wrapped and returned.
//
// --only-verified is intentionally omitted so demo/dummy tokens still surface.
func Scan(path string) ([]Finding, error) {
	if _, err := exec.LookPath("trufflehog"); err != nil {
		return nil, ErrNotInstalled
	}

	cmd := exec.Command("trufflehog", "filesystem", path, "--json", "--no-update")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("trufflehog start: %w", err)
	}

	var findings []Finding
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		var r rawOutput
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			continue
		}
		if r.DetectorName == "" {
			continue
		}
		findings = append(findings, Finding{
			DetectorName: r.DetectorName,
			Raw:          r.Raw,
			Verified:     r.Verified,
			File:         r.SourceMetadata.Data.Filesystem.File,
			Line:         r.SourceMetadata.Data.Filesystem.Line,
		})
	}

	if err := cmd.Wait(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 183 {
			return findings, nil
		}
		return findings, fmt.Errorf("trufflehog exited: %w", err)
	}
	return findings, nil
}

// MaskSecret keeps the first 8 characters and replaces the rest with ****.
// Short secrets are returned with the suffix appended so output is consistent.
func MaskSecret(s string) string {
	const prefix = 8
	if len(s) <= prefix {
		return s + "****"
	}
	return s[:prefix] + "****"
}
