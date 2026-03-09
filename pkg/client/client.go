package client

import (
	"context"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client interface {
	Namespaces(ctx context.Context) ([]Namespace, error)
	ReleasesTable(ctx context.Context) (*v1.Table, error)
	AddonsTable(ctx context.Context) (*v1.Table, error)
	Releases(ctx context.Context) (*v1alpha2.TanzuKubernetesReleaseList, error)
	Cluster(ctx context.Context, ns, name string) (*v1alpha2.TanzuKubernetesCluster, error)
	Clusters(ctx context.Context, ns string) (*v1.Table, error)
	Login(ctx context.Context, u, p string) error
	LoginCluster(ctx context.Context, cluster, namespace string) (*LoginClusterResponse, error)
}
