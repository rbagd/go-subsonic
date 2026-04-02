//go:build noaudio

package ui

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestUINoSubsonicDeps(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, string(out))
	}

	needle := "go-subsonic/internal/subsonic"
	for _, line := range strings.Split(string(bytes.TrimSpace(out)), "\n") {
		if strings.TrimSpace(line) == needle {
			t.Fatalf("internal/ui depends on %q; UI should depend only on internal/core and friends", needle)
		}
	}
}
