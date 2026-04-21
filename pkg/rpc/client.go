package rpc

import (
	"fmt"

	"google.golang.org/grpc"
)

// NewClient creates a gRPC client connection.
func NewClient(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection %s: %w", target, err)
	}
	return conn, nil
}

// MustNewClient panics when NewClient returns an error.
func MustNewClient(target string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(target, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
