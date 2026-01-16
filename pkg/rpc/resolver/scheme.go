// Package resolver provides gRPC resolver implementations for service discovery.
// It includes resolvers for etcd-based service discovery and direct connection.
package resolver

const (
	// SchemeEtcd is the scheme used for etcd-based service discovery.
	// Targets using this scheme should be in the format: etcd:///service-name
	SchemeEtcd = "etcd"

	// SchemeDirect is the scheme used for direct connection without service discovery.
	// Targets using this scheme should be in the format: direct:///ip1:port1,ip2:port2
	SchemeDirect = "direct"
)
