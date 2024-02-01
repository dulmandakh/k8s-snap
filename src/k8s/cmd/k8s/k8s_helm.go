package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	helmCmd = &cobra.Command{
		Use:                "helm",
		Hidden:             true,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Allow users to provide their own kubeconfig but
			// fallback to the admin config if nothing is provided.
			if os.Getenv("KUBECONFIG") == "" {
				os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf")
			}
			path, err := exec.LookPath("helm")
			if err != nil {
				return fmt.Errorf("helm not found")
			}

			command := append(
				[]string{"helm"},
				args...,
			)
			// completly replace the executable with helm
			// as we want to be as close as possible to a "real"
			// helm invocation.
			return syscall.Exec(path, command, os.Environ())
		},
	}
)

func init() {
	rootCmd.AddCommand(helmCmd)
}