package discovery

import (
	"testing"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	grpcresolver "google.golang.org/grpc/resolver"
	serviceconfig "google.golang.org/grpc/serviceconfig"
)

type testClientConn struct {
	state grpcresolver.State
}

func (c *testClientConn) UpdateState(state grpcresolver.State) error {
	c.state = state
	return nil
}
func (c *testClientConn) ReportError(error)                                    {}
func (c *testClientConn) NewAddress([]grpcresolver.Address)                    {}
func (c *testClientConn) NewServiceConfig(string)                              {}
func (c *testClientConn) ParseServiceConfig(string) *serviceconfig.ParseResult { return nil }

func TestDirectResolverBuilderScheme(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	builder := NewDirectResolver(log).Builder()
	if builder.Scheme() != "direct" {
		t.Fatalf("unexpected scheme: %s", builder.Scheme())
	}
}

func TestDirectResolverBuilderBuild(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	builder := NewDirectResolver(log).Builder()
	cc := &testClientConn{}
	resolver, err := builder.Build(grpcresolver.Target{URL: *mustParseURL(t, "direct:///127.0.0.1:9001,127.0.0.1:9002")}, cc, grpcresolver.BuildOptions{})
	if err != nil {
		t.Fatalf("build resolver: %v", err)
	}
	defer resolver.Close()

	if len(cc.state.Addresses) != 2 {
		t.Fatalf("unexpected address count: %d", len(cc.state.Addresses))
	}
	if cc.state.Addresses[0].Addr != "127.0.0.1:9001" {
		t.Fatalf("unexpected first addr: %s", cc.state.Addresses[0].Addr)
	}
	if cc.state.Addresses[1].Addr != "127.0.0.1:9002" {
		t.Fatalf("unexpected second addr: %s", cc.state.Addresses[1].Addr)
	}
}
