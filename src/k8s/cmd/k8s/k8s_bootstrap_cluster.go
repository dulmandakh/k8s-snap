package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	bootstrapClusterCmd = &cobra.Command{
		Use:   "bootstrap-cluster",
		Short: "Create new cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			client, err := cluster.NewClient(cmd.Context(), cluster.ClusterOpts{
				RemoteAddress: clusterCmdOpts.remoteAddress,
				Debug:         rootCmdOpts.logDebug,
				Port:          clusterCmdOpts.port,
				StorageDir:    clusterCmdOpts.storageDir,
				Verbose:       rootCmdOpts.logVerbose,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			cluster, err := client.Bootstrap(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to bootstrap cluster: %w", err)
			}

			logrus.Infof("Cluster with member %s on %s created.", cluster.Name, cluster.Address)
			return err
		},
	}
)

func init() {
	rootCmd.AddCommand(bootstrapClusterCmd)
}
