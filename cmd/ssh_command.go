package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iamveen/devpod-proxmox-provider/pkg/options"
	"github.com/iamveen/devpod-proxmox-provider/pkg/proxmox"
)

func runSSHCommand(ctx context.Context, opts options.Options, client proxmox.Client, commandArgs []string, sshPath string, stdin io.Reader, stdout, stderr io.Writer, exit func(int)) error {
	if opts.MachineID == "" {
		return fmt.Errorf("MACHINE_ID is required to run a Devsy workspace command")
	}
	if opts.MachineFolder == "" {
		return fmt.Errorf("MACHINE_FOLDER is required to locate the Devsy workspace SSH key")
	}
	if len(commandArgs) == 0 {
		return fmt.Errorf("a remote command is required")
	}

	vmid, err := findDevsyVMByMachineID(ctx, client, opts.Node, opts.MachineID)
	if err != nil {
		return err
	}
	interfaces, err := client.GetNetworkInterfaces(ctx, opts.Node, vmid)
	if err != nil {
		return fmt.Errorf("getting Devsy VM %d network interfaces: %w", vmid, err)
	}
	ipAddress := findVMIPv4(interfaces)
	if ipAddress == "" {
		return fmt.Errorf("Devsy VM %d has no reachable IPv4 address", vmid)
	}

	keyPath := filepath.Join(opts.MachineFolder, "id_ed25519")
	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("locating Devsy workspace SSH key %q: %w", keyPath, err)
	}
	remoteCommand := strings.Join(commandArgs, " ")
	sshArgs := []string{
		"-T",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"devsy@" + ipAddress,
		remoteCommand,
	}

	process := exec.CommandContext(ctx, sshPath, sshArgs...)
	process.Stdin = stdin
	process.Stdout = stdout
	process.Stderr = stderr
	if err := process.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exit(exitErr.ExitCode())
			return nil
		}
		return fmt.Errorf("running SSH command for Devsy VM %d: %w", vmid, err)
	}
	return nil
}

func findDevsyVMByMachineID(ctx context.Context, client proxmox.Client, node, machineID string) (int, error) {
	expectedName := "devsy-" + machineID
	resources, err := client.GetClusterResources(ctx)
	if err != nil {
		return 0, fmt.Errorf("finding Devsy VM %q: %w", expectedName, err)
	}

	best := 0
	for _, resource := range resources {
		if resource.Type != "qemu" || resource.Name != expectedName || resource.Node != node {
			continue
		}
		if best == 0 || resource.Status == "running" {
			best = resource.VMID
		}
	}
	if best == 0 {
		return 0, fmt.Errorf("Devsy VM not found: %s", expectedName)
	}
	return best, nil
}

func findVMIPv4(interfaces []proxmox.NetworkInterface) string {
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			continue
		}
		for _, address := range iface.IPAddresses {
			ip := net.ParseIP(address.IPAddress)
			if address.Type == "ipv4" && ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return address.IPAddress
			}
		}
	}
	return ""
}
