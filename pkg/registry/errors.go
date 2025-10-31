package registry

import "errors"

var (
	// 配置相关错误
	ErrEmptyEndpoints   = errors.New("etcd endpoints cannot be empty")
	ErrEmptyServiceName = errors.New("service name cannot be empty")
	ErrInvalidTTL       = errors.New("TTL must be at least 10 seconds")

	// 运行时错误
	ErrNotRegistered     = errors.New("service not registered")
	ErrAlreadyRegistered = errors.New("service already registered")
	ErrLeaseExpired      = errors.New("lease has expired")
	ErrClientClosed      = errors.New("etcd client is closed")
)
