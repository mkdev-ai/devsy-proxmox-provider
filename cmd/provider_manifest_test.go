package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProviderManifest_DevsyContract(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate manifest relative to test source")
	}
	manifestPath := filepath.Join(filepath.Dir(sourceFile), "..", "provider.yaml.tmpl")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("reading provider manifest: %v", err)
	}
	manifest := string(manifestBytes)

	// No literal tilde is allowed: Devsy's daemon installer rejects a
	// non-expanded "~" as a non-absolute state root. The agent dataPath
	// must be absolute, or omitted so Devsy uses its own absolute default.
	for _, obsolete := range []string{"DEVPOD", "~/.devpod", "helper sh -c"} {
		if strings.Contains(manifest, obsolete) {
			t.Errorf("manifest contains obsolete CLI/agent syntax %q", obsolete)
		}
	}
	for _, required := range []string{
		"path: ${DEVSY}",
		"command: ${PROXMOX_PROVIDER} command \"${COMMAND}\"",
		"init: ${PROXMOX_PROVIDER} init",
		"create: ${PROXMOX_PROVIDER} create",
		"delete: ${PROXMOX_PROVIDER} delete",
		"start: ${PROXMOX_PROVIDER} start",
		"stop: ${PROXMOX_PROVIDER} stop",
		"status: ${PROXMOX_PROVIDER} status",
	} {
		if !strings.Contains(manifest, required) {
			t.Errorf("manifest missing required Devsy contract entry %q", required)
		}
	}
}
