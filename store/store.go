package store

import (
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/kobolog/gorb/types"
)

type Store interface {
	Close()
	ListServices() ([]*types.Service, error)
	ListBackends(vsID string) ([]*types.Backend, error)
	CreateService(vsID string, opts *types.Service) error
	UpdateService(vsID string, opts *types.Service) error
	CreateBackend(vsID, rsID string, opts *types.Backend) error
	UpdateBackend(vsID, rsID string, opts *types.Backend) error
	RemoveService(vsID string) error
	RemoveBackend(rsID string) error
}

type storeImpl struct {
	kvstore          store.Store
	storeServicePath string
	storeBackendPath string
	stopCh           chan struct{}
}

func New(storeURLs []string, storeServicePath, storeBackendPath string) (Store, error) {
	var scheme string
	var storePath string
	var hosts []string
	for _, storeURL := range storeURLs {
		uri, err := url.Parse(storeURL)
		if err != nil {
			return nil, err
		}
		uriScheme := strings.ToLower(uri.Scheme)
		if scheme != "" && scheme != uriScheme {
			return nil, errors.New("schemes must be the same for all store URLs")
		}
		if storePath != "" && storePath != uri.Path {
			return nil, errors.New("paths must be the same for all store URLs")
		}
		scheme = uriScheme
		storePath = uri.Path
		hosts = append(hosts, uri.Host)
	}

	var backend store.Backend
	switch scheme {
	case "consul":
		backend = store.CONSUL
	case "etcd":
		backend = store.ETCD
	case "zookeeper":
		backend = store.ZK
	case "boltdb":
		backend = store.BOLTDB
	case "mock":
		backend = "mock"
	default:
		return nil, errors.New("unsupported uri scheme : " + scheme)
	}
	kvstore, err := libkv.NewStore(
		backend,
		hosts,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		return nil, err
	}

	store := &storeImpl{
		kvstore:          kvstore,
		storeServicePath: path.Join(storePath, storeServicePath),
		storeBackendPath: path.Join(storePath, storeBackendPath),
		stopCh:           make(chan struct{}),
	}

	//store.Sync()
	//storeTimer := time.NewTicker(time.Duration(syncTime) * time.Second)
	//go func() {
	//	for {
	//		select {
	//		case <-storeTimer.C:
	//			store.Sync()
	//		case <-store.stopCh:
	//			storeTimer.Stop()
	//			return
	//		}
	//	}
	//}()
	//
	return store, nil
}

//func (s *storeImpl) Sync() {
//	// build external services map
//	services, err := s.getExternalServices()
//	if err != nil {
//		log.Errorf("error while get services: %s", err)
//		return
//	}
//	// build external backends map
//	backends, err := s.getExternalBackends()
//	if err != nil {
//		log.Errorf("error while get backends: %s", err)
//		return
//	}
//	// synchronize context
//	s.ctx.Synchronize(services, backends)
//}

func (s *storeImpl) ListServices() ([]*types.Service, error) {
	var services []*types.Service
	// build external service map (temporary all services)
	kvlist, err := s.kvstore.List(s.storeServicePath)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return services, nil
		}
		return nil, err
	}
	for _, kvpair := range kvlist {
		var options types.Service
		if err := json.Unmarshal(kvpair.Value, &options); err != nil {
			return nil, err
		}
		services = append(services, &options)
	}
	return services, nil
}

func (s *storeImpl) ListBackends(vsID string) ([]*types.Backend, error) {
	var backends []*types.Backend
	// build external backend map
	kvlist, err := s.kvstore.List(s.storeBackendPath)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return backends, nil
		}
		return nil, err
	}
	for _, kvpair := range kvlist {
		var options types.Backend
		if err := json.Unmarshal(kvpair.Value, &options); err != nil {
			return nil, err
		}
		backends = append(backends, &options)
	}
	return backends, nil
}

func (s *storeImpl) Close() {
	close(s.stopCh)
}

func (s *storeImpl) CreateService(vsID string, opts *types.Service) error {
	// put to store
	if err := s.put(s.storeServicePath+"/"+vsID, opts, false); err != nil {
		log.Errorf("error while put service to store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) UpdateService(vsID string, opts *types.Service) error {
	// put to store
	if err := s.put(s.storeServicePath+"/"+vsID, opts, true); err != nil {
		log.Errorf("error while put service to store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) CreateBackend(vsID, rsID string, opts *types.Backend) error {
	// put to store
	if err := s.put(s.storeBackendPath+"/"+rsID, opts, false); err != nil {
		log.Errorf("error while put backend to store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) UpdateBackend(vsID, rsID string, opts *types.Backend) error {
	// put to store
	if err := s.put(s.storeBackendPath+"/"+rsID, opts, true); err != nil {
		log.Errorf("error while put(update) backend to store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) RemoveService(vsID string) error {
	if err := s.kvstore.DeleteTree(s.storeServicePath + "/" + vsID); err != nil {
		log.Errorf("error while delete service from store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) RemoveBackend(rsID string) error {
	if err := s.kvstore.DeleteTree(s.storeBackendPath + "/" + rsID); err != nil {
		log.Errorf("error while delete backend from store: %s", err)
		return err
	}
	return nil
}

func (s *storeImpl) put(key string, value interface{}, overwrite bool) error {
	// marshal value
	var byteValue []byte
	var isDir bool
	if value == nil {
		byteValue = nil
		isDir = true
	} else {
		_bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		byteValue = _bytes
		isDir = false
	}
	// check key exist (create if not exists)
	exist, err := s.kvstore.Exists(key)
	if err != nil {
		return err
	}
	if !exist || overwrite {
		writeOptions := &store.WriteOptions{IsDir: isDir, TTL: 0}
		if err := s.kvstore.Put(key, byteValue, writeOptions); err != nil {
			return err
		}
	}
	return nil
}

func (s *storeImpl) getID(key string) string {
	index := strings.LastIndex(key, "/")
	if index <= 0 {
		return key
	}
	return key[index+1:]
}

func init() {
	consul.Register()
	etcd.Register()
	zookeeper.Register()
	boltdb.Register()
}
