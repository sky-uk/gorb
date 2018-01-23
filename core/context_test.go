package core

import (
	"testing"

	"syscall"

	"strings"

	"github.com/kobolog/gorb/disco"
	"github.com/kobolog/gorb/pulse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeDisco struct {
	mock.Mock
}

func (d *fakeDisco) Expose(name, host string, port uint16) error {
	args := d.Called(name, host, port)
	return args.Error(0)
}

func (d *fakeDisco) Remove(name string) error {
	args := d.Called(name)
	return args.Error(0)
}

type fakeIpvs struct {
	mock.Mock
}

func (f *fakeIpvs) Init() error {
	args := f.Called()
	return args.Error(0)
}

func (f *fakeIpvs) Exit() {
	f.Called()
}

func (f *fakeIpvs) Flush() error {
	args := f.Called()
	return args.Error(0)
}

func (f *fakeIpvs) AddService(vip string, port uint16, protocol uint16, sched string, flags []string) error {
	args := f.Called(vip, port, protocol, sched, flags)
	return args.Error(0)
}

func (f *fakeIpvs) DelService(vip string, port uint16, protocol uint16) error {
	args := f.Called(vip, port, protocol)
	return args.Error(0)
}

func (f *fakeIpvs) AddDestPort(vip string, vport uint16, rip string, rport uint16, protocol uint16, weight uint32, fwd string) error {
	args := f.Called(vip, vport, rip, rport, protocol, weight, fwd)
	return args.Error(0)
}

func (f *fakeIpvs) UpdateDestPort(vip string, vport uint16, rip string, rport uint16, protocol uint16, weight uint32, fwd string) error {
	args := f.Called(vip, vport, rip, rport, protocol, weight, fwd)
	return args.Error(0)

}
func (f *fakeIpvs) DelDestPort(vip string, vport uint16, rip string, rport uint16, protocol uint16) error {
	args := f.Called(vip, vport, rip, rport, protocol)
	return args.Error(0)
}

func newRoutineContext(backends map[string]*backend, ipvs IPVS) *Context {
	c := newContext(ipvs, &fakeDisco{})
	c.backends = backends
	return c
}

func newContext(ipvs IPVS, disco disco.Driver) *Context {
	return &Context{
		ipvs:     ipvs,
		services: map[string]*service{},
		backends: make(map[string]*backend),
		pulseCh:  make(chan pulse.Update),
		stopCh:   make(chan struct{}),
		disco:    disco,
	}
}

var (
	vsID           = "virtualServiceId"
	rsID           = "realServerID"
	virtualService = service{options: &ServiceOptions{Port: 80, Host: "localhost", Protocol: "tcp"}}
)

func TestServiceIsCreated(t *testing.T) {
	options := &ServiceOptions{Port: 80, Host: "localhost", Protocol: "tcp", Method: "sh"}
	mockIpvs := &fakeIpvs{}
	mockDisco := &fakeDisco{}
	c := newContext(mockIpvs, mockDisco)

	mockIpvs.On("AddService", "127.0.0.1", uint16(80), uint16(syscall.IPPROTO_TCP), "sh", []string(nil)).Return(nil)
	mockDisco.On("Expose", vsID, "127.0.0.1", uint16(80)).Return(nil)

	err := c.createService(vsID, options)
	assert.NoError(t, err)
	mockIpvs.AssertExpectations(t)
	mockDisco.AssertExpectations(t)
}

func TestServiceIsCreatedWithShFlags(t *testing.T) {
	options := &ServiceOptions{Port: 80, Host: "localhost", Protocol: "tcp", Method: "sh", Flags: "sh-port|sh-fallback"}
	mockIpvs := &fakeIpvs{}
	mockDisco := &fakeDisco{}
	c := newContext(mockIpvs, mockDisco)

	mockIpvs.On("AddService", "127.0.0.1", uint16(80), uint16(syscall.IPPROTO_TCP), "sh",
		strings.Split(options.Flags, "|")).Return(nil)
	mockDisco.On("Expose", vsID, "127.0.0.1", uint16(80)).Return(nil)

	err := c.createService(vsID, options)
	assert.NoError(t, err)
	mockIpvs.AssertExpectations(t)
	mockDisco.AssertExpectations(t)
}

func TestServiceIsCreatedWithGenericCustomFlags(t *testing.T) {
	options := &ServiceOptions{Port: 80, Host: "localhost", Protocol: "tcp", Method: "sh", Flags: "flag-1|flag-2|flag-3"}
	mockIpvs := &fakeIpvs{}
	mockDisco := &fakeDisco{}
	c := newContext(mockIpvs, mockDisco)

	mockIpvs.On("AddService", "127.0.0.1", uint16(80), uint16(syscall.IPPROTO_TCP), "sh",
		strings.Split(options.Flags, "|")).Return(nil)
	mockDisco.On("Expose", vsID, "127.0.0.1", uint16(80)).Return(nil)

	err := c.createService(vsID, options)
	assert.NoError(t, err)
	mockIpvs.AssertExpectations(t)
	mockDisco.AssertExpectations(t)
}

func TestPulseUpdateSetsBackendWeightToZeroOnStatusDown(t *testing.T) {
	stash := make(map[pulse.ID]uint32)
	backends := map[string]*backend{rsID: {service: &virtualService, options: &BackendOptions{Weight: 100}}}
	mockIpvs := &fakeIpvs{}

	c := newRoutineContext(backends, mockIpvs)

	mockIpvs.On("UpdateDestPort", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, uint32(0), mock.Anything).Return(nil)

	c.processPulseUpdate(stash, pulse.Update{pulse.ID{VsID: vsID, RsID: rsID}, pulse.Metrics{Status: pulse.StatusDown}})
	assert.Equal(t, len(stash), 1)
	assert.Equal(t, stash[pulse.ID{VsID: vsID, RsID: rsID}], uint32(100))
	mockIpvs.AssertExpectations(t)
}

func TestPulseUpdateIncreasesBackendWeightRelativeToTheHealthOnStatusUp(t *testing.T) {
	stash := map[pulse.ID]uint32{pulse.ID{VsID: vsID, RsID: rsID}: uint32(12)}
	backends := map[string]*backend{rsID: {service: &virtualService, options: &BackendOptions{}}}
	mockIpvs := &fakeIpvs{}

	c := newRoutineContext(backends, mockIpvs)

	mockIpvs.On("UpdateDestPort", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, uint32(6), mock.Anything).Return(nil)

	c.processPulseUpdate(stash, pulse.Update{pulse.ID{VsID: vsID, RsID: rsID}, pulse.Metrics{Status: pulse.StatusUp, Health: 0.5}})
	assert.Equal(t, len(stash), 1)
	assert.Equal(t, stash[pulse.ID{VsID: vsID, RsID: rsID}], uint32(12))
	mockIpvs.AssertExpectations(t)
}

func TestPulseUpdateRemovesStashWhenBackendHasFullyRecovered(t *testing.T) {
	stash := map[pulse.ID]uint32{pulse.ID{VsID: vsID, RsID: rsID}: uint32(12)}
	backends := map[string]*backend{rsID: {service: &virtualService, options: &BackendOptions{}}}
	mockIpvs := &fakeIpvs{}

	c := newRoutineContext(backends, mockIpvs)

	mockIpvs.On("UpdateDestPort", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, uint32(12), mock.Anything).Return(nil)

	c.processPulseUpdate(stash, pulse.Update{pulse.ID{VsID: vsID, RsID: rsID}, pulse.Metrics{Status: pulse.StatusUp, Health: 1}})
	assert.Empty(t, stash)
	mockIpvs.AssertExpectations(t)
}

func TestPulseUpdateRemovesStashWhenBackendIsDeleted(t *testing.T) {
	stash := map[pulse.ID]uint32{pulse.ID{VsID: vsID, RsID: rsID}: uint32(0)}
	backends := make(map[string]*backend)
	mockIpvs := &fakeIpvs{}

	c := newRoutineContext(backends, mockIpvs)
	c.processPulseUpdate(stash, pulse.Update{pulse.ID{VsID: vsID, RsID: rsID}, pulse.Metrics{}})

	assert.Empty(t, stash)
	mockIpvs.AssertExpectations(t)
}

func TestPulseUpdateRemovesStashWhenDeletedAfterNotification(t *testing.T) {
	stash := map[pulse.ID]uint32{pulse.ID{VsID: vsID, RsID: rsID}: uint32(0)}
	backends := map[string]*backend{rsID: {service: &virtualService, options: &BackendOptions{}}}
	mockIpvs := &fakeIpvs{}

	c := newRoutineContext(backends, mockIpvs)
	c.processPulseUpdate(stash, pulse.Update{pulse.ID{VsID: vsID, RsID: rsID}, pulse.Metrics{Status: pulse.StatusRemoved}})

	assert.Empty(t, stash)
	mockIpvs.AssertExpectations(t)
}
