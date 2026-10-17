package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/options"
	"github.com/mkdev-ai/devsy-proxmox-provider/pkg/proxmox"
	"github.com/spf13/cobra"
)

var (
	versionStr       = "dev"
	newProxmoxClient = func(opts options.Options) proxmox.Client {
		return proxmox.NewHTTPClient(opts.Host, opts.Port, opts.Token)
	}

	// Root command is exported for testing.
	RootCmd = &cobra.Command{
		Use:     "proxmox-provider",
		Short:   "Devsy provider for Proxmox VE",
		Long:    "Provisions KVM virtual machines on Proxmox VE to host Devsy development workspaces.",
		Version: versionStr,
	}
)

// Execute is the entry point called from main.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.SetOut(os.Stdout)
	RootCmd.SetErr(os.Stderr)
}

// versionCmd prints the binary version.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the provider version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), versionStr)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

// initCmd validates Proxmox connectivity.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize and validate the Proxmox provider",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	opts := options.FromEnv()
	if err := opts.Validate(); err != nil {
		return err
	}
	version, err := validateProxmoxReadiness(ctx, newProxmoxClient(opts), opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Connected to Proxmox %s (%s)\n", version.Version, version.Release)
	return nil
}

func validateProxmoxReadiness(ctx context.Context, client proxmox.Client, opts options.Options) (*proxmox.VersionResponse, error) {
	version, err := client.GetVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Proxmox API; check PROXMOX_HOST, PROXMOX_PORT, and PROXMOX_TOKEN: %w", err)
	}
	resources, err := client.GetClusterResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Proxmox cluster resources; verify the API token has cluster resource access: %w", err)
	}

	nodeOnline := false
	templateFound := false
	for _, resource := range resources {
		if resource.Type == "node" && resource.Node == opts.Node && resource.Status == "online" {
			nodeOnline = true
		}
		if resource.Type == "qemu" && resource.Node == opts.Node && resource.Template != 0 && resource.Name == opts.Template {
			templateFound = true
		}
	}
	if !nodeOnline {
		return nil, fmt.Errorf("Proxmox node %q is not available or online; check PROXMOX_NODE and node status", opts.Node)
	}
	if !templateFound {
		return nil, fmt.Errorf("Proxmox VM template %q was not found on node %q; check PROXMOX_TEMPLATE", opts.Template, opts.Node)
	}
	return version, nil
}

func init() {
	RootCmd.AddCommand(initCmd)
}

// statusCmd reports the VM status to Devsy.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Report the status of a workspace VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		opts := options.FromEnv()
		if err := opts.Validate(); err != nil {
			return err
		}

		client := newProxmoxClient(opts)
		vmid, err := findVMByMachineIDWithClient(ctx, client, opts)
		if err != nil {
			// VM not found
			fmt.Fprintln(cmd.OutOrStdout(), "notfound")
			return nil
		}

		s, err := client.GetVMStatus(ctx, opts.Node, vmid)
		if err != nil {
			// 404 / not found from API
			fmt.Fprintln(cmd.OutOrStdout(), "notfound")
			return nil
		}

		switch s.Status {
		case "running":
			fmt.Fprintln(cmd.OutOrStdout(), "running")
		case "stopped":
			fmt.Fprintln(cmd.OutOrStdout(), "stopped")
		default:
			fmt.Fprintln(cmd.OutOrStdout(), "busy")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}

// findVMByMachineID scans cluster resources for a VM matching the machine ID.
func findVMByMachineID(ctx context.Context, opts options.Options) (int, error) {
	return findVMByMachineIDWithClient(ctx, newProxmoxClient(opts), opts)
}

func findVMByMachineIDWithClient(ctx context.Context, client proxmox.Client, opts options.Options) (int, error) {
	resources, err := client.GetClusterResources(ctx)
	if err != nil {
		return 0, err
	}

	expectedName := "devsy-" + opts.MachineID
	best := 0
	for _, r := range resources {
		if r.Type != "qemu" || r.Name != expectedName || r.Node != opts.Node {
			continue
		}
		if best == 0 || r.Status == "running" {
			best = r.VMID
		}
	}
	if best == 0 {
		return 0, fmt.Errorf("VM not found: %s", expectedName)
	}
	return best, nil
}
