package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/options"
	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/proxmox"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a workspace VM",
	RunE:  runDelete,
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	opts := options.FromEnv()
	if err := opts.Validate(); err != nil {
		return err
	}

	client := newProxmoxClient(opts)

	vmid, err := findVMByMachineIDWithClient(ctx, client, opts)
	if err != nil {
		// VM not found — already deleted
		return nil
	}

	// Check current status
	s, err := client.GetVMStatus(ctx, opts.Node, vmid)
	if err != nil {
		// VM already gone
		return nil
	}

	// If running, shut it down first
	if s.Status == "running" {
		fmt.Fprintf(cmd.OutOrStdout(), "Shutting down VM %d...\n", vmid)
		supid, err := client.ShutdownVM(ctx, opts.Node, vmid)
		if err != nil {
			return fmt.Errorf("shutting down VM: %w", err)
		}
		// Wait for the shutdown task to complete; a running VM will make the
		// Proxmox destroy API return 500 "VM is still running".
		if err := client.WaitForTask(ctx, opts.Node, supid, 30*time.Second); err != nil {
			return fmt.Errorf("shutdown of VM %d timed out: %w", vmid, err)
		}
	}

	// Destroy the VM (VM must be stopped; a transient failure triggers an
	// emergency cleanup that stops, then destroys without purge).
	upid, err := client.DeleteVM(ctx, opts.Node, vmid, true)
	if err != nil {
		return fmt.Errorf("deleting VM: %w", err)
	}
	if err := client.WaitForTask(ctx, opts.Node, upid, 5*time.Minute); err != nil {
		return fmt.Errorf("waiting for delete: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "VM %d deleted\n", vmid)
	return nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a stopped workspace VM",
	RunE:  runStart,
}

func init() {
	RootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	opts := options.FromEnv()
	if err := opts.Validate(); err != nil {
		return err
	}

	client := newProxmoxClient(opts)

	vmid, err := findVMByMachineIDWithClient(ctx, client, opts)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	upid, err := client.StartVM(ctx, opts.Node, vmid)
	if err != nil {
		return fmt.Errorf("starting VM: %w", err)
	}
	if err := client.WaitForTask(ctx, opts.Node, upid, 5*time.Minute); err != nil {
		return fmt.Errorf("waiting for VM start: %w", err)
	}
	return nil
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running workspace VM",
	RunE:  runStop,
}

func init() {
	RootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	opts := options.FromEnv()
	if err := opts.Validate(); err != nil {
		return err
	}

	client := newProxmoxClient(opts)

	vmid, err := findVMByMachineIDWithClient(ctx, client, opts)
	if err != nil {
		return fmt.Errorf("VM not found: %w", err)
	}

	upid, err := client.ShutdownVM(ctx, opts.Node, vmid)
	if err != nil {
		return fmt.Errorf("shutting down VM: %w", err)
	}
	if err := client.WaitForTask(ctx, opts.Node, upid, 5*time.Minute); err != nil {
		return fmt.Errorf("waiting for VM shutdown: %w", err)
	}
	return nil
}

var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Run a Devsy agent command on the workspace VM over SSH",
	RunE:  runCommand,
}

func init() {
	RootCmd.AddCommand(commandCmd)
}

func runCommand(cmd *cobra.Command, args []string) error {
	opts := options.FromEnv()
	if err := opts.Validate(); err != nil {
		return err
	}
	client := proxmox.NewHTTPClient(opts.Host, opts.Port, opts.Token)
	return runSSHCommand(cmd.Context(), opts, client, args, "ssh", cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Exit)
}
