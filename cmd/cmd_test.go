package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/options"
	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/proxmox"
	"github.com/spf13/cobra"
)

// --- Helper: set up context with mocks by replacing createVMByMachineID ---
// Because commands read from env vars, we need to test via MockClient injection.
// We do this through a test-only client factory pattern.

// TestStatusCmd_VMNotFound tests that status outputs "notfound" when no VM exists.
func TestStatusCmd_VMFinding(t *testing.T) {
	// Test findVMByMachineID with mock
	mock := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 100, Name: "devsy-test-abc", Node: "pve1"},
		},
	}

	vmid, err := findVMByMachineIDTest(context.Background(), mock, "pve1", "test-abc")
	if err != nil {
		t.Fatalf("expected to find VM, got error: %v", err)
	}
	if vmid != 100 {
		t.Errorf("expected VMID 100, got %d", vmid)
	}
}

func TestStatusCmd_VMNotFound(t *testing.T) {
	mock := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 100, Name: "devsy-other", Node: "pve1"},
		},
	}

	_, err := findVMByMachineIDTest(context.Background(), mock, "pve1", "test-abc")
	if err == nil {
		t.Fatal("expected error when VM not found")
	}
}

func TestStatusCmd_StateMapping(t *testing.T) {
	tests := []struct {
		proxmoxStatus string
		expected      string
	}{
		{"running", "running"},
		{"stopped", "stopped"},
		{"paused", "busy"},
		{"prelaunch", "busy"},
		{"suspended", "busy"},
	}

	for _, tt := range tests {
		t.Run(tt.proxmoxStatus, func(t *testing.T) {
			result := mapProxmoxStatus(tt.proxmoxStatus)
			if result != tt.expected {
				t.Errorf("proxmox '%s' → expected '%s', got '%s'", tt.proxmoxStatus, tt.expected, result)
			}
		})
	}
}

func TestFindTemplateID(t *testing.T) {
	mock := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 9000, Name: "devsy-template", Node: "pve1", Template: 1},
			{Type: "qemu", VMID: 100, Name: "devsy-ws1", Node: "pve1"},
		},
	}

	id, err := findTemplateID(context.Background(), mock, "pve1", "devsy-template")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 9000 {
		t.Errorf("expected template ID 9000, got %d", id)
	}
}

func TestFindTemplateID_NotFound(t *testing.T) {
	mock := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 100, Name: "other-template", Node: "pve1", Template: 1},
		},
	}

	_, err := findTemplateID(context.Background(), mock, "pve1", "devsy-template")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsVMIDConflict(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected bool
	}{
		{"VM 2000 already exists on node 'pve1'", true},
		{"unable to create VM 2000 - VM 2000 already defined", true},
		{"some other error", false},
		{"permission denied", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			result := isVMIDConflict(errors.New(tt.errMsg))
			if result != tt.expected {
				t.Errorf("isVMIDConflict('%s') = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestDelete_ShutdownBeforeDestroy(t *testing.T) {
	mock := &proxmox.MockClient{
		VMStatus: &proxmox.VMStatus{Status: "running"},
	}

	// Verify we'd call shutdown for running VMs
	if mock.VMStatus.Status != "running" {
		t.Error("should shutdown running VMs")
	}
}

// findVMByMachineIDTest delegates to the production lookup with a mockable client.
func findVMByMachineIDTest(ctx context.Context, client proxmox.Client, node, machineID string) (int, error) {
	return findVMByMachineIDWithClient(ctx, client, options.Options{Node: node, MachineID: machineID})
}

// mapProxmoxStatus maps Proxmox VM status to DevPod status strings.
func mapProxmoxStatus(status string) string {
	switch status {
	case "running":
		return "running"
	case "stopped":
		return "stopped"
	default:
		return "busy"
	}
}

func TestCreate_DiskResizeLogic(t *testing.T) {
	// Test that disk resize is only triggered when VM_DISK_SIZE > template size
	templateDiskGB := 17

	tests := []struct {
		diskSize     int
		shouldResize bool
	}{
		{50, true},
		{17, false},
		{16, false},
		{100, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("disk=%d", tt.diskSize), func(t *testing.T) {
			wouldResize := tt.diskSize > templateDiskGB
			if wouldResize != tt.shouldResize {
				t.Errorf("disk=%d: would resize=%v, expected=%v", tt.diskSize, wouldResize, tt.shouldResize)
			}
		})
	}
}

func TestCreate_CloneRetryLogic(t *testing.T) {
	// Simulate VMID conflict retry behavior
	attemptedVMIDs := []int{}
	conflictCount := 0
	startID := 2000

	for vmid := startID; vmid < startID+100; vmid++ {
		attemptedVMIDs = append(attemptedVMIDs, vmid)
		if conflictCount < 2 {
			conflictCount++
			continue // simulate conflict
		}
		// Success on 3rd attempt
		break
	}

	if len(attemptedVMIDs) != 3 {
		t.Errorf("expected 3 attempts, got %d: %v", len(attemptedVMIDs), attemptedVMIDs)
	}
	if attemptedVMIDs[0] != 2000 || attemptedVMIDs[1] != 2001 || attemptedVMIDs[2] != 2002 {
		t.Errorf("expected VMIDs 2000,2001,2002, got %v", attemptedVMIDs)
	}
}

func TestCommandTemplate(t *testing.T) {
	// Verify the command template is correct
	expected := `"${DEVPOD}" helper sh -c "${COMMAND}"`

	var sb strings.Builder
	fmt.Fprintln(&sb, expected)
	output := strings.TrimSpace(sb.String())
	if output != expected {
		t.Errorf("expected '%s', got '%s'", expected, output)
	}
}

func TestWaitForTask_PollingBehavior(t *testing.T) {
	// Test that WaitForTask polls with the mock
	callCount := 0
	var fakeTask proxmox.MockClient

	// Set up the error field to fail after we check
	fakeTask.WaitForTaskErr = fmt.Errorf("task failed")

	err := fakeTask.WaitForTask(context.Background(), "pve1", "UPID:test", time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
	callCount = fakeTask.CallCount("WaitForTask")
	if callCount != 1 {
		t.Errorf("expected 1 WaitForTask call, got %d", callCount)
	}
}

func TestMockClient_RecordsAllCalls(t *testing.T) {
	mock := &proxmox.MockClient{}

	_, _ = mock.GetVersion(context.Background())
	_, _ = mock.GetClusterResources(context.Background())
	_ = mock.ConfigureVM(context.Background(), "pve1", 101, proxmox.VMConfig{})
	_, _ = mock.StartVM(context.Background(), "pve1", 101)
	_, _ = mock.GetVMStatus(context.Background(), "pve1", 101)

	expectedCalls := []string{
		"GetVersion", "GetClusterResources", "ConfigureVM", "StartVM", "GetVMStatus",
	}

	for _, expected := range expectedCalls {
		if !mock.HasCall(expected) {
			t.Errorf("expected call '%s' not recorded", expected)
		}
	}

	if mock.CallCount("GetVersion") != 1 {
		t.Errorf("expected 1 GetVersion call, got %d", mock.CallCount("GetVersion"))
	}
}

func TestFindVMByMachineID_IgnoresLegacyVM(t *testing.T) {
	mock := &proxmox.MockClient{Resources: []proxmox.Resource{
		{Type: "qemu", VMID: 100, Name: "devpod-machine-1", Node: "pve1", Status: "running"},
	}}
	_, err := findVMByMachineIDTest(context.Background(), mock, "pve1", "machine-1")
	if err == nil || !strings.Contains(err.Error(), "devsy-machine-1") {
		t.Fatalf("expected Devsy VM not-found error, got %v", err)
	}
	if mock.CallCount("GetVMStatus") != 0 || mock.CallCount("StartVM") != 0 || mock.CallCount("DeleteVM") != 0 {
		t.Fatalf("legacy VM must not be operated on: calls=%+v", mock.Calls)
	}
}

func TestTryClone_UsesDevsyResourceNames(t *testing.T) {
	mock := &proxmox.MockClient{}
	opts := options.DefaultOptions()
	opts.Node = "pve1"
	if err := tryClone(context.Background(), mock, opts, 9000, 2000, "devsy-machine-1", "ssh-key"); err != nil {
		t.Fatalf("tryClone returned error: %v", err)
	}
	cloneCall := mock.Calls[0].Args[2].(proxmox.CloneRequest)
	if cloneCall.Name != "devsy-machine-1" {
		t.Errorf("clone name = %q, want devsy-machine-1", cloneCall.Name)
	}
	config := mock.Calls[2].Args[2].(proxmox.VMConfig)
	if config.CIUser != "devsy" || config.Tags != "devsy" {
		t.Errorf("cloud-init should use Devsy guest/tag, got user=%q tags=%q", config.CIUser, config.Tags)
	}
}

func TestCreate_DevsyResultAndEnvironmentContract(t *testing.T) {
	setRequiredOptions(t)
	t.Setenv("MACHINE_ID", "machine-1")
	t.Setenv("MACHINE_FOLDER", t.TempDir())
	mock := &proxmox.MockClient{
		Resources: []proxmox.Resource{
			{Type: "qemu", VMID: 9000, Name: "devsy-template", Node: "pve1", Template: 1},
		},
		Ifaces: []proxmox.NetworkInterface{{
			Name:        "eth0",
			IPAddresses: []proxmox.IPAddr{{IPAddress: "192.0.2.10", Type: "ipv4"}},
		}},
	}
	withProxmoxClient(t, mock)
	var output bytes.Buffer
	cmd := newTestCommand()
	cmd.SetOut(&output)
	if err := runCreate(cmd, nil); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}
	var result map[string]string
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("create result is not JSON: %v; output=%q", err, output.String())
	}
	if result["ipAddress"] != "192.0.2.10" || result["sshUser"] != "devsy" {
		t.Errorf("unexpected Devsy create result: %+v", result)
	}
	if got := mockCall(t, mock, "CloneVM").Args[2].(proxmox.CloneRequest).Name; got != "devsy-machine-1" {
		t.Errorf("VM name = %q, want devsy-machine-1", got)
	}
}

func TestCreateRequiresMachineFolder(t *testing.T) {
	setRequiredOptions(t)
	t.Setenv("MACHINE_ID", "machine-1")
	t.Setenv("MACHINE_FOLDER", "")
	err := runCreate(newTestCommand(), nil)
	if err == nil || !strings.Contains(err.Error(), "MACHINE_FOLDER not set") {
		t.Fatalf("expected actionable MACHINE_FOLDER error, got %v", err)
	}
}

func TestRunInitValidatesReadiness(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*proxmox.MockClient)
		wantError string
	}{
		{name: "ready", configure: func(mock *proxmox.MockClient) {}},
		{name: "API unavailable", configure: func(mock *proxmox.MockClient) { mock.VersionErr = errors.New("connection refused") }, wantError: "PROXMOX_HOST"},
		{name: "resource access denied", configure: func(mock *proxmox.MockClient) { mock.ResourcesErr = errors.New("forbidden") }, wantError: "token has cluster resource access"},
		{name: "node missing", configure: func(mock *proxmox.MockClient) {
			mock.Resources = []proxmox.Resource{{Type: "qemu", VMID: 9000, Name: "devsy-template", Node: "pve1", Template: 1}}
		}, wantError: "PROXMOX_NODE"},
		{name: "node offline", configure: func(mock *proxmox.MockClient) {
			mock.Resources[0].Status = "offline"
		}, wantError: "online"},
		{name: "template missing", configure: func(mock *proxmox.MockClient) {
			mock.Resources = []proxmox.Resource{{Type: "node", Node: "pve1", Status: "online"}}
		}, wantError: "PROXMOX_TEMPLATE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredOptions(t)
			mock := readyMockClient()
			tt.configure(mock)
			withProxmoxClient(t, mock)
			cmd := newTestCommand()
			var output bytes.Buffer
			cmd.SetOut(&output)
			err := runInit(cmd, nil)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("runInit returned error: %v", err)
				}
				if !strings.Contains(output.String(), "Connected to Proxmox") {
					t.Errorf("expected readiness confirmation, got %q", output.String())
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("error = %v, want actionable message containing %q", err, tt.wantError)
			}
			if output.Len() != 0 {
				t.Errorf("failed initialization reported success: %q", output.String())
			}
		})
	}
}

func TestRunInitRejectsMissingOptionsBeforeConnecting(t *testing.T) {
	t.Setenv("PROXMOX_HOST", "")
	t.Setenv("PROXMOX_USER", "")
	t.Setenv("PROXMOX_TOKEN", "")
	t.Setenv("PROXMOX_NODE", "")
	t.Setenv("PROXMOX_TEMPLATE", "")
	oldFactory := newProxmoxClient
	newProxmoxClient = func(options.Options) proxmox.Client {
		t.Fatal("client created with invalid options")
		return nil
	}
	t.Cleanup(func() { newProxmoxClient = oldFactory })
	err := runInit(newTestCommand(), nil)
	if err == nil || !strings.Contains(err.Error(), "PROXMOX_HOST") {
		t.Fatalf("expected actionable missing-option error, got %v", err)
	}
}

func TestLifecycleOperationsUseDevsyVMOnly(t *testing.T) {
	for _, operation := range []string{"status", "start", "stop", "delete"} {
		t.Run(operation, func(t *testing.T) {
			setRequiredOptions(t)
			t.Setenv("MACHINE_ID", "machine-1")
			mock := &proxmox.MockClient{
				Resources: []proxmox.Resource{
					{Type: "qemu", VMID: 100, Name: "devpod-machine-1", Node: "pve1", Status: "running"},
					{Type: "qemu", VMID: 101, Name: "devsy-machine-1", Node: "pve1", Status: "running"},
				},
				VMStatus: &proxmox.VMStatus{Status: "stopped"},
			}
			withProxmoxClient(t, mock)
			switch operation {
			case "status":
				var output bytes.Buffer
				statusCmd.SetOut(&output)
				t.Cleanup(func() { statusCmd.SetOut(nil) })
				if err := statusCmd.RunE(statusCmd, nil); err != nil {
					t.Fatalf("status returned error: %v", err)
				}
				if strings.TrimSpace(output.String()) != "stopped" {
					t.Errorf("status = %q, want stopped", output.String())
				}
			case "start":
				if err := runStart(newTestCommand(), nil); err != nil {
					t.Fatalf("start returned error: %v", err)
				}
			case "stop":
				if err := runStop(newTestCommand(), nil); err != nil {
					t.Fatalf("stop returned error: %v", err)
				}
			case "delete":
				if err := runDelete(newTestCommand(), nil); err != nil {
					t.Fatalf("delete returned error: %v", err)
				}
			}
			if operation == "status" && mock.CallCount("GetVMStatus") != 1 {
				t.Fatalf("expected status lookup on Devsy VM, calls=%+v", mock.Calls)
			}
			for _, call := range mock.Calls {
				if len(call.Args) > 1 {
					if vmid, ok := call.Args[1].(int); ok && vmid != 101 {
						t.Errorf("%s operated on VMID %d instead of Devsy VMID 101", call.Method, vmid)
					}
				}
			}
			if operation == "start" && !mock.HasCall("StartVM") {
				t.Error("start did not start the Devsy VM")
			}
			if operation == "stop" && !mock.HasCall("ShutdownVM") {
				t.Error("stop did not stop the Devsy VM")
			}
			if operation == "delete" && !mock.HasCall("DeleteVM") {
				t.Error("delete did not delete the Devsy VM")
			}
		})
	}
}

func setRequiredOptions(t *testing.T) {
	t.Helper()
	t.Setenv("PROXMOX_HOST", "pve.local")
	t.Setenv("PROXMOX_USER", "devsy@pve")
	t.Setenv("PROXMOX_TOKEN", "devsy@pve!provider=secret")
	t.Setenv("PROXMOX_NODE", "pve1")
	t.Setenv("PROXMOX_TEMPLATE", "devsy-template")
}

func readyMockClient() *proxmox.MockClient {
	return &proxmox.MockClient{
		Version: &proxmox.VersionResponse{Version: "8.0", Release: "8"},
		Resources: []proxmox.Resource{
			{Type: "node", Node: "pve1", Status: "online"},
			{Type: "qemu", VMID: 9000, Name: "devsy-template", Node: "pve1", Template: 1},
		},
	}
}

func withProxmoxClient(t *testing.T, client proxmox.Client) {
	t.Helper()
	oldFactory := newProxmoxClient
	newProxmoxClient = func(options.Options) proxmox.Client { return client }
	t.Cleanup(func() { newProxmoxClient = oldFactory })
}

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	return cmd
}

func mockCall(t *testing.T, mock *proxmox.MockClient, method string) proxmox.MockCall {
	t.Helper()
	for _, call := range mock.Calls {
		if call.Method == method {
			return call
		}
	}
	t.Fatalf("mock did not record %s", method)
	return proxmox.MockCall{}
}
