package core

import (
	"testing"

	"github.com/kobolog/gorb/ipvs-shim"
	"github.com/kobolog/gorb/options"
	"github.com/mqliang/libipvs"
	"github.com/stretchr/testify/mock"
)

type storeMock struct {
	mock.Mock
}

func (s *storeMock) Close() {}
func (s *storeMock) ListServices() ([]*options.ServiceOptions, error) {
	return nil, nil
}
func (s *storeMock) ListBackends(vsID string) ([]*options.BackendOptions, error) {
	return nil, nil
}

type ipvsMock struct {
	mock.Mock
}

func (i *ipvsMock) Init() error {
	return nil
}
func (i *ipvsMock) Flush() error {
	return nil
}
func (i *ipvsMock) AddService(svc *ipvs_shim.Service) error {
	return nil
}
func (i *ipvsMock) UpdateService(svc *ipvs_shim.Service) error {
	return nil
}
func (i *ipvsMock) DelService(key *ipvs_shim.ServiceKey) error {
	return nil
}
func (i *ipvsMock) ListServices() ([]*ipvs_shim.Service, error) {
	return nil, nil
}
func (i *ipvsMock) AddBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	return nil
}
func (i *ipvsMock) UpdateBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	return nil
}
func (i *ipvsMock) DelBackend(key *ipvs_shim.ServiceKey, backend *ipvs_shim.Backend) error {
	return nil
}
func (i *ipvsMock) ListBackends(key *ipvs_shim.ServiceKey) ([]*ipvs_shim.Backend, error) {
	return nil, nil
}

func TestReconcile(t *testing.T) {
	svc1 := &options.ServiceOptions{}
	ipvsSvc1 := &libipvs.Service{}

	tests := []struct {
		name            string
		actualServices  []*libipvs.Service
		desiredServices []*options.ServiceOptions
		createdServices []*libipvs.Service
	}{
		{
			name:            "add new service",
			actualServices:  []*libipvs.Service{},
			desiredServices: []*options.ServiceOptions{svc1},
			createdServices: []*libipvs.Service{ipvsSvc1},
		},
	}
	for _, tt := range tests {
		storeMock := &storeMock{}
		ipvsMock := &ipvsMock{}

		t.Run(tt.name, func(t *testing.T) {
			r := &reconciler{
				store: storeMock,
				ipvs:  ipvsMock,
			}

			storeMock.On("ListServices").Return(tt.desiredServices, nil)
			ipvsMock.On("ListServices").Return(tt.actualServices, nil)
			for _, s := range tt.createdServices {
				ipvsMock.On("AddService", s).Return(nil)
			}

			r.reconcile()

			storeMock.AssertExpectations(t)
			ipvsMock.AssertExpectations(t)
		})
	}
}
