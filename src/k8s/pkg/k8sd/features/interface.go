package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// Interface abstracts the management of built-in Canonical Kubernetes features.
type Interface interface {
	// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
	ApplyDNS(context.Context, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (string, error)
	// ApplyNetwork is used to configure the network feature on Canonical Kubernetes.
	ApplyNetwork(context.Context, snap.Snap, types.Network, types.Annotations) error
	// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
	ApplyLoadBalancer(context.Context, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) error
	// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
	ApplyIngress(context.Context, snap.Snap, types.Ingress, types.Network, types.Annotations) error
	// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
	ApplyGateway(context.Context, snap.Snap, types.Gateway, types.Network, types.Annotations) error
	// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
	ApplyMetricsServer(context.Context, snap.Snap, types.MetricsServer, types.Annotations) error
	// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
	ApplyLocalStorage(context.Context, snap.Snap, types.LocalStorage, types.Annotations) error
}

// implementation implements Interface.
type implementation struct {
	applyDNS           func(context.Context, snap.Snap, types.DNS, types.Kubelet, types.Annotations) (string, error)
	applyNetwork       func(context.Context, snap.Snap, types.Network, types.Annotations) error
	applyLoadBalancer  func(context.Context, snap.Snap, types.LoadBalancer, types.Network, types.Annotations) error
	applyIngress       func(context.Context, snap.Snap, types.Ingress, types.Network, types.Annotations) error
	applyGateway       func(context.Context, snap.Snap, types.Gateway, types.Network, types.Annotations) error
	applyMetricsServer func(context.Context, snap.Snap, types.MetricsServer, types.Annotations) error
	applyLocalStorage  func(context.Context, snap.Snap, types.LocalStorage, types.Annotations) error
}

func (i *implementation) ApplyDNS(ctx context.Context, snap snap.Snap, dns types.DNS, kubelet types.Kubelet, annotations types.Annotations) (string, error) {
	return i.applyDNS(ctx, snap, dns, kubelet, annotations)
}

func (i *implementation) ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network, annotations types.Annotations) error {
	return i.applyNetwork(ctx, snap, cfg, annotations)
}

func (i *implementation) ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network, annotations types.Annotations) error {
	return i.applyLoadBalancer(ctx, snap, loadbalancer, network, annotations)
}

func (i *implementation) ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network, annotations types.Annotations) error {
	return i.applyIngress(ctx, snap, ingress, network, annotations)
}

func (i *implementation) ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, annotations types.Annotations) error {
	return i.applyGateway(ctx, snap, gateway, network, annotations)
}

func (i *implementation) ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) error {
	return i.applyMetricsServer(ctx, snap, cfg, annotations)
}

func (i *implementation) ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage, annotations types.Annotations) error {
	return i.applyLocalStorage(ctx, snap, cfg, annotations)
}
