package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iamveen/devpod-proxmox-provider/pkg/options"
	"github.com/iamveen/devpod-proxmox-provider/pkg/proxmox"
)

func TestRunSSHCommand_RelaysStreamsAndExitStatus(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("test-key"), 0600); err != nil {
		t.Fatal(err)
	}

	argsPath := filepath.Join(tempDir, "ssh-args")
	sshPath := filepath.Join(tempDir, "ssh")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + shellQuote(argsPath) + "\ncat\nprintf 'ssh stderr' >&2\nexit 23\n"
	if err := os.WriteFile(sshPath, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}

	client := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 19, Name: "devpod-machine-123", Node: "pve1", Status: "running"},
			{Type: "qemu", VMID: 42, Name: "devsy-machine-123", Node: "pve1", Status: "running"},
		},
		Ifaces: []proxmox.NetworkInterface{{
			Name:        "eth0",
			IPAddresses: []proxmox.IPAddr{{IPAddress: "192.0.2.42", Type: "ipv4"}},
		}},
	}
	opts := options.Options{Node: "pve1", MachineID: "machine-123", MachineFolder: tempDir}
	input := []byte("agent protocol input\x00with raw bytes")
	var stdout, stderr bytes.Buffer
	var exitCode int
	var exitCalls int

	err := runSSHCommand(context.Background(), opts, client, []string{"sh -c 'cat /dev/stdin'"}, sshPath, bytes.NewReader(input), &stdout, &stderr, func(code int) {
		exitCode = code
		exitCalls++
	})
	if err != nil {
		t.Fatalf("runSSHCommand() error = %v", err)
	}
	if !bytes.Equal(stdout.Bytes(), input) {
		t.Fatalf("stdout = %q, want exact input %q", stdout.Bytes(), input)
	}
	if got, want := stderr.String(), "ssh stderr"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
	if exitCalls != 1 || exitCode != 23 {
		t.Fatalf("exit callback = (%d calls, code %d), want (1 call, code 23)", exitCalls, exitCode)
	}
	if !client.HasCall("GetClusterResources") || !client.HasCall("GetNetworkInterfaces") {
		t.Fatalf("expected Proxmox VM and interface lookups, got calls: %+v", client.Calls)
	}

	sshArgs, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	joinedArgs := string(sshArgs)
	for _, expected := range []string{"-T\n", keyPath + "\n", "IdentitiesOnly=yes\n", "192.0.2.42\n", "devsy@192.0.2.42\n", "sh -c 'cat /dev/stdin'\n"} {
		if !strings.Contains(joinedArgs, expected) {
			t.Errorf("SSH args %q do not contain %q", joinedArgs, expected)
		}
	}
}

func TestRunSSHCommand_DoesNotUseLegacyVM(t *testing.T) {
	client := &proxmox.MockClient{Resources: []proxmox.Resource{{
		Type: "qemu", VMID: 19, Name: "devpod-machine-123", Node: "pve1", Status: "running",
	}}}
	var stdout, stderr bytes.Buffer
	err := runSSHCommand(context.Background(), options.Options{
		Node: "pve1", MachineID: "machine-123", MachineFolder: t.TempDir(),
	}, client, []string{"true"}, "ssh", strings.NewReader(""), &stdout, &stderr, func(int) {})
	if err == nil || !strings.Contains(err.Error(), "Devsy VM not found") {
		t.Fatalf("error = %v, want Devsy VM not found", err)
	}
	if client.HasCall("GetNetworkInterfaces") {
		t.Fatal("looked up a legacy VM's network interfaces")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout contains provider diagnostics: %q", stdout.String())
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestCommandHelpUsesDevsyTerminology(t *testing.T) {
	if strings.Contains(strings.ToLower(commandCmd.Short), "devpod") {
		t.Fatalf("command help contains obsolete DevPod terminology: %q", commandCmd.Short)
	}
	if !strings.Contains(commandCmd.Short, "Devsy") {
		t.Fatalf("command help does not mention Devsy: %q", commandCmd.Short)
	}
}
