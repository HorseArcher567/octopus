package etcd

// This file contains the etcd-backed discovery provider and its adapters.

import (
	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Provider is the etcd-backed discovery provider.
type Provider struct {
	log    *xlog.Logger
	client *clientv3.Client
}

// NewProvider creates an etcd-backed discovery provider.
func NewProvider(log *xlog.Logger, client *clientv3.Client) *Provider {
	return &Provider{log: log, client: client}
}

// Registrar returns an etcd-backed registrar.
func (p *Provider) Registrar() discovery.Registrar {
	return &Registrar{log: p.log, client: p.client}
}

// Resolver returns an etcd-backed resolver.
func (p *Provider) Resolver() discovery.Resolver {
	return &Resolver{log: p.log, client: p.client}
}

// GRPCResolverBuilder returns a gRPC resolver builder backed by the provider resolver.
func (p *Provider) GRPCResolverBuilder() any {
	return NewGRPCResolverBuilder(p.log, p.Resolver())
}

// Close releases provider-owned resources.
func (p *Provider) Close() error { return nil }
