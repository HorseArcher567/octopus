package registry

import (
	"errors"
	"fmt"
)

// Instance 服务实例信息
type Instance struct {
	Name     string            `json:"name"`
	Addr     string            `json:"addr"`
	Port     int               `json:"port"`
	Version  string            `json:"version"`
	Zone     string            `json:"zone,omitempty"`
	Weight   int               `json:"weight,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (s *Instance) Validate() error {
	if s == nil {
		return errors.New("service instance is nil")
	}
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.Addr == "" {
		return errors.New("addr is required")
	}
	if s.Port <= 0 {
		return errors.New("port is required")
	}
	return nil
}

func (s *Instance) Key() string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/%s:%d", s.Name, s.Addr, s.Port)
}
