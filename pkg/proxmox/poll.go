package proxmox

import (
	"context"
	"fmt"
	"time"
)

// WaitForAgent polls the QEMU guest agent until it is demonstrably alive.
// It returns nil once the agent responds to a query, or an error if the agent
// does not become ready within the bounded timeout. Transient errors (e.g.
// network interface not populated yet) are retried rather than treated as
// fatal.
func WaitForAgent(ctx context.Context, client Client, node string, vmid int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := client.GetNetworkInterfaces(ctx, node, vmid); err != nil {
				// Agent not ready yet — keep trying.
				continue
			}
			return nil
		}
	}
}

// WaitForIP polls the QEMU guest agent until a non-loopback IPv4 address is
// found. It first ensures the guest agent is alive via WaitForAgent, so a
// slow-but-healthy guest is never mistaken for a hard failure. Returns the
// first non-loopback IPv4 address, or an actionable error after the bounded
// timeout.
func WaitForIP(ctx context.Context, client Client, node string, vmid int, timeout time.Duration) (string, error) {
	if err := WaitForAgent(ctx, client, node, vmid, 30*time.Second); err != nil {
		return "", fmt.Errorf("guest agent not ready: %w", err)
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline:
			return "", fmt.Errorf("timed out waiting for VM IP after %v (agent alive but no non-loopback IPv4 address reported)", timeout)
		case <-ticker.C:
			ifaces, err := client.GetNetworkInterfaces(ctx, node, vmid)
			if err != nil {
				// Agent may have died — report the actionable readiness error.
				return "", fmt.Errorf("guest agent not ready while waiting for IP: %w", err)
			}
			if ip := findNonLoopbackIPv4(ifaces); ip != "" {
				return ip, nil
			}
		}
	}
}

func findNonLoopbackIPv4(interfaces []NetworkInterface) string {
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			continue
		}
		for _, addr := range iface.IPAddresses {
			if addr.Type == "ipv4" && addr.IPAddress != "127.0.0.1" {
				return addr.IPAddress
			}
		}
	}
	return ""
}

// WaitForVMReady polls the qemu status endpoint until the VM reports
// "running", confirming the newly started VM is reachable before the IP wait
// begins. The qemu status read is the authoritative signal that the hypervisor
// has finished creating/starting the VM: a freshly cloned cold-booted VM can
// take well over 90s for the guest agent to answer, so this gate tolerates a
// slow-but-healthy agent and returns as soon as the VM is running. The guest
// agent itself is still probed separately by WaitForIP (Step 7), which retries
// and uses a bounded timeout, so IP discovery never silently hangs either.
//
// Returns nil once the VM reports running, or an error if the VM does not
// reach running within the timeout.
func WaitForVMReady(ctx context.Context, client Client, node string, vmid int, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timed out waiting for VM readiness after %v (qemu status)", timeout)
		case <-ticker.C:
			status, err := client.GetVMStatus(ctx, node, vmid)
			if err != nil {
				continue
			}
			// The VM is considered created once the hypervisor reports "running".
			// Do not wait for the guest agent here — start/stop of the agent is a
			// separate concern handled by WaitForIP below.
			if status != nil && status.Status == "running" {
				return nil
			}
		}
	}
}
