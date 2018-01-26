package core

import (
	"testing"

	"encoding/json"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	libkvmock "github.com/docker/libkv/store/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type storeMock struct {
	libkvmock.Mock
}

var storeURLs = []string{"mock://127.0.0.1:2000", "mock://127.0.0.2:2001", "mock://127.0.0.3:2002"}

func (s *storeMock) mockNew() func(endpoints []string, options *store.Config) (store.Store, error) {
	return func(endpoints []string, options *store.Config) (store.Store, error) {
		s.Endpoints = endpoints
		s.Options = options
		return &s.Mock, nil
	}
}

func TestMultipleURLs(t *testing.T) {
	assert := assert.New(t)
	m := storeMock{}
	libkv.AddStore("mock", m.mockNew())
	m.On("List", "/").Return([]*store.KVPair{}, nil)

	store, err := NewStore(storeURLs, "/", "/", 60, &Context{})

	assert.NoError(err)
	assert.Equal([]string{"127.0.0.1:2000", "127.0.0.2:2001", "127.0.0.3:2002"}, m.Endpoints)

	store.Close()
}

func TestErrorIfSchemeMismatch(t *testing.T) {
	assert := assert.New(t)
	m := storeMock{}
	libkv.AddStore("mock", m.mockNew())
	m.On("List", "/").Return([]*store.KVPair{}, nil)

	storeURLs := []string{"mock://127.0.0.1:2000", "mismatch://127.0.0.2:2001", "mock://127.0.0.3:2002"}
	_, err := NewStore(storeURLs, "/", "/", 60, &Context{})

	assert.Error(err)
}

func TestErrorIfPathMismatch(t *testing.T) {
	assert := assert.New(t)
	m := storeMock{}
	libkv.AddStore("mock", m.mockNew())
	m.On("List", "/").Return([]*store.KVPair{}, nil)

	storeURLs := []string{"mock://127.0.0.1:2000", "mock://127.0.0.2:2001/mismatched/path/", "mock://127.0.0.3:2002"}
	_, err := NewStore(storeURLs, "/", "/", 60, &Context{})

	assert.Error(err)
}

func TestUpdateService(t *testing.T) {
	m := storeMock{}
	libkv.AddStore("mock", m.mockNew())

	vsID := "my-virtual-server"
	opts := &ServiceOptions{
		Host:       "10.10.0.0",
		Port:       8080,
		Protocol:   "tcp",
		Method:     "sh",
		Flags:      "flag-1|flag-2",
		Persistent: false,
	}
	optsBytes, _ := json.Marshal(opts)
	m.On("List", "").Return([]*store.KVPair{}, nil)
	m.On("Exists", "/"+vsID).Return(false, nil)
	m.On("Put", "/"+vsID, optsBytes, mock.Anything).Return(nil)

	store, _ := NewStore(storeURLs, "", "", 60, &Context{})
	store.UpdateService(vsID, opts)

	m.AssertExpectations(t)
}
