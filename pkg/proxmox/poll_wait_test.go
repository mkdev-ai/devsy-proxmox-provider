package proxmox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/proxmox"
)

// TestWaitForAgent_Alive verifies the readiness probe returns once the agent
// responds to a GetNetworkInterfaces call within the bounded timeout.
func TestWaitForAgent_Alive(t *testing.T) {
	mock := &proxmox.MockClient{
		Ifaces: []proxmox.NetworkInterface{
			{
				Name:        "eth0",
				IPAddresses: []proxmox.IPAddr{{IPAddress: "10.0.0.5", Type: "ipv4", Prefix: 24}},
			},
		},
	}

	if err := proxmox.WaitForAgent(context.Background(), mock, "pve1", 101, 5*time.Second); err != nil {
		t.Fatalf("agent did not become alive: %v", err)
	}
}

// TestWaitForAgent_BoundRetryOnTransientError verifies the readiness probe
// retries transient agent errors until the bounded timeout expires, then
// fails with DeadlineExceeded rather than hanging.
func TestWaitForAgent_BoundRetryOnTransientError(t *testing.T) {
	mock := &proxmox.MockClient{
		IfacesErr: errors.New("agent not ready"),
	}

	start := time.Now()
	err := proxmox.WaitForAgent(context.Background(), mock, "pve1", 101, 200*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a bounded timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
	if elapsed > 250*time.Millisecond {
		t.Errorf("probe hung past the 200ms bound: took %v", elapsed)
	}
}

// TestWaitForIP_RequiresAliveAgent verifies the new contract: a guest whose
// agent is alive but slow (loopback-only IP at first, then a real address)
// must be allowed to report the IP instead of being treated as a hard
// failure mid-probe.
func TestWaitForIP_RequiresAliveAgent(t *testing.T) {
	tiered := []proxmox.NetworkInterface{
		{
			Name:        "lo",
			IPAddresses: []proxmox.IPAddr{{IPAddress: "127.0.0.1", Type: "ipv4", Prefix: 8}},
		},
		{
			Name:        "eth0",
			IPAddresses: []proxmox.IPAddr{{IPAddress: "10.0.0.5", Type: "ipv4", Prefix: 24}},
		},
	}
	mock := &proxmox.MockClient{Ifaces: tiered}

	ip, err := proxmox.WaitForIP(context.Background(), mock, "pve1", 101, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.0.0.5" {
		t.Errorf("expected IP 10.0.0.5, got %q", ip)
	}
}

// TestWaitForIP_ActionableTimeout verifies that when the agent is alive but no
// non-loopback IPv4 appears, the error is actionable: it names the waited
// duration and the alive-agent condition instead of a bare "timed out".
func TestWaitForIP_ActionableTimeout(t *testing.T) {
	mock := &proxmox.MockClient{
		Ifaces: []proxmox.NetworkInterface{
			{
				Name:        "lo",
				IPAddresses: []proxmox.IPAddr{{IPAddress: "127.0.0.1", Type: "ipv4", Prefix: 8}},
			},
		},
	}

	_, err := proxmox.WaitForIP(context.Background(), mock, "pve1", 101, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	e := err.Error()
	if !containsAny(e, "timed out") {
		t.Errorf("expected 'timed out' in error; got: %s", e)
	}
	if !containsAny(e, "10m") && !containsAny(e, "agent alive") {
		t.Errorf("expected actionable message naming duration/agent-alive condition; got: %s", e)
	}
}

func containsAny(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
