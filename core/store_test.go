package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store/mock"
)

func TestMultipleURIs(t *testing.T) {
	assert := assert.New(t)
	libkv.AddStore("mock", mock.New)

	storeURLs := []string{"mock://127.0.0.1:2000", "mock://127.0.0.2:2001", "mock://127.0.0.3:2002"}
	store, err := NewStore(storeURLs, "/", "/", 60, &Context{})
	assert.NoError(err)
	store.Close()
}
