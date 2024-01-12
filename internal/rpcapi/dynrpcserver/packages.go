package dynrpcserver

import (
	"context"
	"sync"

	tf1 "github.com/hashicorp/terraform/internal/rpcapi/terraform1"
)

type Packages struct {
	impl tf1.PackagesServer
	mu   sync.RWMutex
}

var _ tf1.PackagesServer = (*Packages)(nil)

func NewPackagesStub() *Packages {
	return &Packages{}
}

func (s *Packages) FetchModulePackage(a0 context.Context, a1 *tf1.FetchModulePackage_Request) (*tf1.FetchModulePackage_Response, error) {
	impl, err := s.realRPCServer()
	if err != nil {
		return nil, err
	}
	return impl.FetchModulePackage(a0, a1)
}

func (s *Packages) FetchProviderPackage(a0 context.Context, a1 *tf1.FetchProviderPackage_Request) (*tf1.FetchProviderPackage_Response, error) {
	impl, err := s.realRPCServer()
	if err != nil {
		return nil, err
	}
	return impl.FetchProviderPackage(a0, a1)
}

func (s *Packages) ModulePackageSourceAddr(a0 context.Context, a1 *tf1.ModulePackageSourceAddr_Request) (*tf1.ModulePackageSourceAddr_Response, error) {
	impl, err := s.realRPCServer()
	if err != nil {
		return nil, err
	}
	return impl.ModulePackageSourceAddr(a0, a1)
}

func (s *Packages) ModulePackageVersions(a0 context.Context, a1 *tf1.ModulePackageVersions_Request) (*tf1.ModulePackageVersions_Response, error) {
	impl, err := s.realRPCServer()
	if err != nil {
		return nil, err
	}
	return impl.ModulePackageVersions(a0, a1)
}

func (s *Packages) ProviderPackageVersions(a0 context.Context, a1 *tf1.ProviderPackageVersions_Request) (*tf1.ProviderPackageVersions_Response, error) {
	impl, err := s.realRPCServer()
	if err != nil {
		return nil, err
	}
	return impl.ProviderPackageVersions(a0, a1)
}

func (s *Packages) ActivateRPCServer(impl tf1.PackagesServer) {
	s.mu.Lock()
	s.impl = impl
	s.mu.Unlock()
}

func (s *Packages) realRPCServer() (tf1.PackagesServer, error) {
	s.mu.RLock()
	impl := s.impl
	s.mu.RUnlock()
	if impl == nil {
		return nil, unavailableErr
	}
	return impl, nil
}
