package k8s

import (
	"fmt"
	"io"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

type JoinClusterResult struct {
	Name string `json:"name" yaml:"name"`
}

func (b JoinClusterResult) String() string {
	return fmt.Sprintf("Cluster services have started on %q.\nPlease allow some time for initial Kubernetes node registration.\n", b.Name)
}

func newJoinClusterCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		name         string
		address      string
		configFile   string
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "join-cluster <join-token>",
		Short:  "Join a cluster using the provided token",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Args:   cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]

			// Use hostname as default node name
			if opts.name == "" {
				// TODO(neoaggelos): use the encoded node name from the token, if available.
				hostname, err := os.Hostname()
				if err != nil {
					cmd.PrintErrf("Error: --name is not set and could not determine the current node name.\n\nThe error was: %v\n", err)
					env.Exit(1)
					return
				}
				opts.name = hostname
			}

			if opts.address == "" {
				opts.address = util.NetworkInterfaceAddress()
			}
			opts.address = util.CanonicalNetworkAddress(opts.address, config.DefaultPort)

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if client.IsBootstrapped(cmd.Context()) {
				cmd.PrintErrln("Error: The node is already part of a cluster")
				env.Exit(1)
				return
			}

			var joinClusterConfig string
			if opts.configFile != "" {
				var b []byte
				var err error

				if opts.configFile == "-" {
					b, err = io.ReadAll(os.Stdin)
					if err != nil {
						cmd.PrintErrf("Error: Failed to read join configuration from stdin. \n\nThe error was: %v\n", err)
						env.Exit(1)
						return
					}
				} else {
					b, err = os.ReadFile(opts.configFile)
					if err != nil {
						cmd.PrintErrf("Error: Failed to read join configuration from %q.\n\nThe error was: %v\n", opts.configFile, err)
						env.Exit(1)
						return
					}
				}
				joinClusterConfig = string(b)
			}

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(cmd.Context(), apiv1.JoinClusterRequest{Name: opts.name, Address: opts.address, Token: token, Config: joinClusterConfig}); err != nil {
				cmd.PrintErrf("Error: Failed to join the cluster using the provided token.\n\nThe error was: %v\n", err)

				cmd.PrintErrf("\n\nCleaning up...\n\n")
				if err := client.WaitForMicroclusterNodeToBeReady(cmd.Context(), opts.name); err != nil {
					cmd.PrintErrf("Warning: Failed to wait for the node %q to be ready.\n\nThe error was: %v\n", opts.name, err)
				}
				if err := client.CleanupKubernetesServices(cmd.Context()); err != nil {
					cmd.PrintErrf("Warning: Failed to cleanup Kubernetes services after failed bootstrap attempt.\n\nThe error was: %v\n", err)
				}

				// For worker nodes we need to reset the node from the cluster
				isWorker, err := snaputil.IsWorker(env.Snap)
				if err != nil {
					cmd.PrintErrf("Warning: Failed to check if this is worker-only node.\n\nThe error was: %v\n", err)
				}

				if isWorker {
					// triggers pre-remove hook>> remove node from kubernetes
					if err := client.ResetClusterMember(cmd.Context(), opts.name, true); err != nil {
						cmd.PrintErrf("Warning: Failed to reset the node %q.\n\nThe error was: %v\n", opts.name, err)
					}
					// Remove worker node marker
					if is, err := snaputil.IsWorker(env.Snap); err == nil && is {
						if err := snaputil.MarkAsWorkerNode(env.Snap, false); err != nil {
							cmd.PrintErrf("Warning: Failed to remove worker node mark %q.\n\nThe error was: %v\n", opts.name, err)
						}
					}
				} else {
					if err := client.DeleteClusterMember(cmd.Context(), apiv1.RemoveNodeRequest{Name: opts.name, Force: true}); err != nil {
						cmd.PrintErrf("Warning: Failed to remove the node %q from the cluster.\n\nThe error was: %v\n", opts.name, err)
					}
				}

				env.Exit(1)
				return
			}

			outputFormatter.Print(JoinClusterResult{Name: opts.name})
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "node name, defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "microcluster address, defaults to the node IP address")
	cmd.Flags().StringVar(&opts.configFile, "file", "", "path to the YAML file containing your custom cluster join configuration. Use '-' to read from stdin.")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	return cmd
}
