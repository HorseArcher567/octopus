package clientpool

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"google.golang.org/grpc"
	"sync"
)

type ClientPool struct {
	dialOptions []grpc.DialOption

	mu      sync.RWMutex
	clients map[string]*grpc.ClientConn
}

// New returns a ClientPool pointer.
func New(opts ...grpc.DialOption) *ClientPool {
	cp := &ClientPool{
		dialOptions: make([]grpc.DialOption, 0, 16),
		clients:     make(map[string]*grpc.ClientConn),
	}
	cp.dialOptions = append(cp.dialOptions, opts...)

	return cp
}

func (cp *ClientPool) PrepareClient(target string) {
	if client, err := cp.newClient(target); err == nil {
		cp.clients[target] = client
	}
}

func (cp *ClientPool) GetClient(target string) (*grpc.ClientConn, error) {
	return cp.newClient(target)
}

func (cp *ClientPool) MustGetClient(target string) *grpc.ClientConn {
	client, err := cp.GetClient(target)
	if err != nil {
		log.Panicln(err)
		return nil
	}

	return client
}

func (cp *ClientPool) newClient(target string) (*grpc.ClientConn, error) {
	cp.mu.RLock()
	if client, ok := cp.clients[target]; ok {
		cp.mu.RUnlock()
		return client, nil
	}
	cp.mu.RUnlock()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if client, err := grpc.Dial(target, cp.dialOptions...); err == nil {
		cp.clients[target] = client
		return client, nil
	} else {
		return nil, err
	}
}

func (cp *ClientPool) DelClient(target string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if client, ok := cp.clients[target]; ok {
		delete(cp.clients, target)
		_ = client.Close()
	}
}
