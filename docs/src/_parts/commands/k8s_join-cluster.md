## k8s join-cluster

Join a cluster using the provided token

```
k8s join-cluster <join-token> [flags]
```

### Options

```
      --address string         microcluster address, defaults to the node IP address
      --file string            path to the YAML file containing your custom cluster join configuration. Use '-' to read from stdin.
  -h, --help                   help for join-cluster
      --name string            node name, defaults to hostname
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

