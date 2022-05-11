package connpool

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"google.golang.org/grpc"
	"sync"
)

type ConnPool struct {
	dialOptions []grpc.DialOption

	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn
}

// New returns a ConnPool pointer.
func New(opts ...grpc.DialOption) *ConnPool {
	cp := &ConnPool{
		dialOptions: make([]grpc.DialOption, 0, 16),
		conns:       make(map[string]*grpc.ClientConn),
	}
	cp.dialOptions = append(cp.dialOptions, opts...)

	return cp
}

func (cp *ConnPool) PrepareConn(target string) {
	if client, err := cp.newConn(target); err == nil {
		cp.conns[target] = client
	}
}

func (cp *ConnPool) GetConn(target string) (*grpc.ClientConn, error) {
	return cp.newConn(target)
}

func (cp *ConnPool) MustGetConn(target string) *grpc.ClientConn {
	client, err := cp.GetConn(target)
	if err != nil {
		log.Panicln(err)
		return nil
	}

	return client
}

func (cp *ConnPool) newConn(target string) (*grpc.ClientConn, error) {
	cp.mu.RLock()
	if client, ok := cp.conns[target]; ok {
		cp.mu.RUnlock()
		return client, nil
	}
	cp.mu.RUnlock()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if client, err := grpc.Dial(target, cp.dialOptions...); err == nil {
		cp.conns[target] = client
		return client, nil
	} else {
		return nil, err
	}
}

func (cp *ConnPool) DeleteConn(target string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if client, ok := cp.conns[target]; ok {
		delete(cp.conns, target)
		_ = client.Close()
	}
}
