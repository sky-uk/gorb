package core

import (
	"testing"

	"github.com/kobolog/gorb/ipvs-shim"
	"github.com/kobolog/gorb/types"
	"github.com/stretchr/testify/mock"
)

type storeMock struct {
	mock.Mock
}

func (s *storeMock) Close() {
	panic("not implemented")
}
func (s *storeMock) ListServices() ([]*types.Service, error) {
	args := s.Mock.Called()
	return args.Get(0).([]*types.Service), args.Error(1)
}
func (s *storeMock) ListBackends(vsID string) ([]*types.BackendOptions, error) {
	args := s.Mock.Called(vsID)
	return args.Get(0).([]*types.BackendOptions), args.Error(1)
}

type ipvsMock struct {
	mock.Mock
}

func (i *ipvsMock) Init() error {
	panic("not implemented")
}
func (i *ipvsMock) Flush() error {
	panic("not implemented")
}
func (i *ipvsMock) AddService(svc *ipvs_shim.Service) error {
	args := i.Mock.Called(svc)
	return args.Error(0)
}
func (i *ipvsMock) UpdateService(svc *ipvs_shim.Service) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) DelService(key *ipvs_shim.ServiceKey) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) ListServices() ([]*ipvs_shim.Service, error) {
	args := i.Mock.Called()
	return args.Get(0).([]*ipvs_shim.Service), args.Error(1)
}
func (i *ipvsMock) AddBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) UpdateBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) DelBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) ListBackends(key *ipvs_shim.ServiceKey) ([]*ipvs_shim.Backend, error) {
	panic("not implemented")
	return nil, nil
}

func TestReconcile(t *testing.T) {
	storeSvc1 := &types.Service{
		Host:     "10.10.10.1",
		Port:     101,
		Protocol: "tcp",
		Method:   "sh",
		Flags:    []string{"flag-1", "flag-2"},
	}
	storeSvc1u := &types.Service{
		Host:     "10.10.10.1",
		Port:     101,
		Protocol: "tcp",
		Method:   "wrr",
		Flags:    []string{"flag-3"},
	}
	ipvsSvcKey1 := ipvs_shim.ServiceKey{
		VIP:      "10.10.10.1",
		Port:     101,
		Protocol: "tcp",
	}
	ipvsSvc1 := &ipvs_shim.Service{
		ServiceKey: ipvsSvcKey1,
		Scheduler:  "sh",
		Flags:      []string{"flag-1", "flag-2"},
	}
	ipvsSvc1u := &ipvs_shim.Service{
		ServiceKey: ipvsSvcKey1,
		Scheduler:  "wrr",
		Flags:      []string{"flag-3"},
	}

	tests := []struct {
		name            string
		actualServices  []*ipvs_shim.Service
		desiredServices []*types.Service
		createdServices []*ipvs_shim.Service
		updatedServices []*ipvs_shim.Service
	}{
		{
			name:            "add new service",
			desiredServices: []*types.Service{storeSvc1},
			createdServices: []*ipvs_shim.Service{ipvsSvc1},
		},
		{
			name:            "update service",
			actualServices:  []*ipvs_shim.Service{ipvsSvc1},
			desiredServices: []*types.Service{storeSvc1u},
			updatedServices: []*ipvs_shim.Service{ipvsSvc1u},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeMock := &storeMock{}
			ipvsMock := &ipvsMock{}

			r := &reconciler{
				store: storeMock,
				ipvs:  ipvsMock,
			}

			// set defaults
			if tt.actualServices == nil {
				tt.actualServices = []*ipvs_shim.Service{}
			}
			if tt.desiredServices == nil {
				tt.desiredServices = []*types.Service{}
			}

			// add expectations for store and ipvs
			ipvsMock.On("ListServices").Return(tt.actualServices, nil)
			storeMock.On("ListServices").Return(tt.desiredServices, nil)
			for _, s := range tt.createdServices {
				ipvsMock.On("AddService", s).Return(nil)
			}
			for _, s := range tt.updatedServices {
				ipvsMock.On("UpdateService", s).Return(nil)
			}

			// reconcile
			r.reconcile()

			// ensure expected outcomes
			storeMock.AssertExpectations(t)
			ipvsMock.AssertExpectations(t)
		})
	}
}
