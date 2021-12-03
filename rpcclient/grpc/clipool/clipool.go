package clipool

import (
	"sync"

	"google.golang.org/grpc"
)

type ClientPool struct {
	clients map[string]*grpc.ClientConn
	mu      sync.RWMutex

	cliOpts []grpc.DialOption
}

func New(opts ...grpc.DialOption) *ClientPool {
	return &ClientPool{
		clients: make(map[string]*grpc.ClientConn),
		cliOpts: append(opts, grpc.WithInsecure()),
	}
}

func (pool *ClientPool) Get(target string) (*grpc.ClientConn, error) {
	pool.mu.RLock()
	if client, ok := pool.clients[target]; ok {
		pool.mu.RUnlock()
		return client, nil
	}
	pool.mu.RUnlock()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	if client, err := grpc.Dial(target, pool.cliOpts...); err != nil {
		return nil, err
	} else {
		pool.clients[target] = client
		return client, nil
	}
}

func (pool *ClientPool) Put(target string) (*grpc.ClientConn, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if client, err := grpc.Dial(target, grpc.WithInsecure()); err != nil {
		return nil, err
	} else {
		pool.clients[target] = client
		return client, nil
	}
}

func (pool *ClientPool) Del(target string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if conn, ok := pool.clients[target]; ok {
		delete(pool.clients, target)
		conn.Close()
	}
}
