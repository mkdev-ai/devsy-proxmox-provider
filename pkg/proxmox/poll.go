package proxmox

import (
	"context"
	"fmt"
	"time"
)

// WaitForIP polls the QEMU guest agent until a non-loopback IPv4 address is found.
// Returns the first non-loopback IPv4 address, or an error after timeout.
func WaitForIP(ctx context.Context, client Client, node string, vmid int, timeout time.Duration) (string, error) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline:
			return "", fmt.Errorf("timed out waiting for VM IP after %v", timeout)
		case <-ticker.C:
			ifaces, err := client.GetNetworkInterfaces(ctx, node, vmid)
			if err != nil {
				// Agent may not be ready yet — keep trying
				continue
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

// WaitForVMReady polls the qemu status endpoint and the guest agent until both
// respond, confirming the newly started VM is reachable before the IP wait
// begins. Each tick requires BOTH reads to succeed; any single read error
// restarts the tick. Returns nil once both have cleared in one tick, or an
// error if the timeout (approx 90s) is exceeded.
func WaitForVMReady(ctx context.Context, client Client, node string, vmid int, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timed out waiting for VM readiness after %v (qemu status + guest agent)", timeout)
		case <-ticker.C:
			status, err := client.GetVMStatus(ctx, node, vmid)
			if err != nil {
				continue
			}
			// A ready VM should report a running qemu status; still require BOTH
			// reads to succeed in the same tick.
			if status == nil || status.Status != "running" {
				continue
			}
			if _, err := client.GetNetworkInterfaces(ctx, node, vmid); err != nil {
				continue
			}
			// Both responded: the VM is reachable (guest agent answered and the
			// status is running). Return so the create flow can proceed to WaitForIP.
			return nil
		}
	}
}
