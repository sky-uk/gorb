package core

import (
	"testing"

	"net"

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
func (s *storeMock) ListBackends(vsID string) ([]*types.Backend, error) {
	args := s.Mock.Called(vsID)
	return args.Get(0).([]*types.Backend), args.Error(1)
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
func (i *ipvsMock) AddService(svc *types.Service) error {
	args := i.Mock.Called(svc)
	return args.Error(0)
}
func (i *ipvsMock) UpdateService(svc *types.Service) error {
	args := i.Mock.Called(svc)
	return args.Error(0)
}
func (i *ipvsMock) DeleteService(key *types.ServiceKey) error {
	args := i.Mock.Called(key)
	return args.Error(0)
}
func (i *ipvsMock) ListServices() ([]*types.Service, error) {
	args := i.Mock.Called()
	return args.Get(0).([]*types.Service), args.Error(1)
}
func (i *ipvsMock) AddBackend(key *types.ServiceKey, backend *types.Backend) error {
	args := i.Mock.Called(key, backend)
	return args.Error(0)
}
func (i *ipvsMock) UpdateBackend(key *types.ServiceKey, backend *types.Backend) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) DeleteBackend(key *types.ServiceKey, backend *types.Backend) error {
	panic("not implemented")
	return nil
}
func (i *ipvsMock) ListBackends(key *types.ServiceKey) ([]*types.Backend, error) {
	args := i.Mock.Called(key)
	return args.Get(0).([]*types.Backend), args.Error(1)
}

func TestReconcile(t *testing.T) {
	svcKey1 := types.ServiceKey{
		VIP:      net.ParseIP("10.10.10.1"),
		Port:     101,
		Protocol: "tcp",
	}
	svc1 := &types.Service{
		ServiceKey: svcKey1,
		Scheduler:  "sh",
		Flags:      []string{"flag-1", "flag-2"},
		StoreID:    "svc1",
	}
	svc1u := &types.Service{
		ServiceKey: svcKey1,
		Scheduler:  "wrr",
		Flags:      []string{"flag-3"},
	}
	svcKey2 := types.ServiceKey{
		VIP:      net.ParseIP("10.10.10.2"),
		Port:     102,
		Protocol: "udp",
	}
	svc2 := &types.Service{
		ServiceKey: svcKey2,
		Scheduler:  "rr",
		Flags:      []string{"flag-1"},
		StoreID:    "svc2",
	}

	backend1 := &types.Backend{
		IP:      net.ParseIP("172.16.1.1"),
		Port:    501,
		Weight:  1.0,
		Forward: "dr",
	}

	type keyToBackends map[*types.Service][]*types.Backend

	tests := []struct {
		name string

		// services
		actualServices  []*types.Service
		desiredServices []*types.Service
		createdServices []*types.Service
		updatedServices []*types.Service
		deletedServices []*types.Service

		// backends
		actualBackends  keyToBackends
		desiredBackends keyToBackends
		createdBackends keyToBackends
	}{
		{
			name:            "add new service",
			actualServices:  []*types.Service{svc2},
			desiredServices: []*types.Service{svc1, svc2},
			createdServices: []*types.Service{svc1},
		},
		{
			name:            "update service",
			actualServices:  []*types.Service{svc1, svc2},
			desiredServices: []*types.Service{svc1u, svc2},
			updatedServices: []*types.Service{svc1u},
		},
		{
			name:            "no change in service",
			actualServices:  []*types.Service{svc1u, svc2},
			desiredServices: []*types.Service{svc1u, svc2},
		},
		{
			name:            "delete service",
			actualServices:  []*types.Service{svc1, svc2},
			desiredServices: []*types.Service{svc1},
			deletedServices: []*types.Service{svc2},
		},
		{
			name:            "add backend",
			actualServices:  []*types.Service{svc1},
			desiredServices: []*types.Service{svc1},
			desiredBackends: keyToBackends{svc1: {backend1}},
			createdBackends: keyToBackends{svc1: {backend1}},
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
				tt.actualServices = []*types.Service{}
			}
			if tt.desiredServices == nil {
				tt.desiredServices = []*types.Service{}
			}
			if tt.actualBackends == nil {
				tt.actualBackends = make(keyToBackends)
			}
			if tt.desiredBackends == nil {
				tt.desiredBackends = make(keyToBackends)
			}
			for _, svc := range tt.desiredServices {
				if _, ok := tt.actualBackends[svc]; !ok {
					tt.actualBackends[svc] = []*types.Backend{}
				}
				if _, ok := tt.desiredBackends[svc]; !ok {
					tt.desiredBackends[svc] = []*types.Backend{}
				}
			}

			// add expectations for store and ipvs
			// services
			ipvsMock.On("ListServices").Return(tt.actualServices, nil)
			storeMock.On("ListServices").Return(tt.desiredServices, nil)
			for _, s := range tt.createdServices {
				ipvsMock.On("AddService", s).Return(nil)
			}
			for _, s := range tt.updatedServices {
				ipvsMock.On("UpdateService", s).Return(nil)
			}
			for _, s := range tt.deletedServices {
				ipvsMock.On("DeleteService", &s.ServiceKey).Return(nil)
			}
			// backends
			for k, v := range tt.desiredBackends {
				storeMock.On("ListBackends", k.StoreID).Return(v, nil)
			}
			for k, v := range tt.actualBackends {
				ipvsMock.On("ListBackends", &k.ServiceKey).Return(v, nil)
			}
			for k, v := range tt.createdBackends {
				for _, backend := range v {
					ipvsMock.On("AddBackend", &k.ServiceKey, backend).Return(nil)
				}
			}

			// reconcile
			r.reconcile()

			// ensure expected outcomes
			storeMock.AssertExpectations(t)
			ipvsMock.AssertExpectations(t)
		})
	}
}
