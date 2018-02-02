/*
   Copyright (c) 2015 Andrey Sibiryov <me@kobology.ru>
   Copyright (c) 2015 Other contributors as noted in the AUTHORS file.

   This file is part of GORB - Go Routing and Balancing.

   GORB is free software; you can redistribute it and/or modify
   it under the terms of the GNU Lesser General Public License as published by
   the Free Software Foundation; either version 3 of the License, or
   (at your option) any later version.

   GORB is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public License
   along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package core

import (
	"errors"
	"net"
	"sync"

	"github.com/kobolog/gorb/pulse"

	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/kobolog/gorb/store"
	"github.com/kobolog/gorb/types"
)

// Possible runtime errors.
var (
	ErrObjectExists    = errors.New("specified object already exists")
	ErrObjectNotFound  = errors.New("unable to locate specified object")
	ErrIncompatibleAFs = errors.New("incompatible address families")
)

type service struct {
	options *types.Service
}

type backend struct {
	options *types.Backend
	service *service
	monitor *pulse.Pulse
	metrics pulse.Metrics
}

// Context abstacts away the underlying IPVS bindings implementation.
type Context struct {
	store      store.Store
	stopCh     chan struct{}
	reconciler *reconciler

	sync.Mutex
}

// ContextOptions configure Context behavior.
type ContextOptions struct {
	Endpoints    []net.IP
	Flush        bool
	ListenPort   uint16
	VipInterface string
	SyncTime     time.Duration
	Store        store.Store
}

// NewContext creates a new Context and initializes IPVS.
func NewContext(options ContextOptions) (*Context, error) {
	log.Info("initializing IPVS context")

	server := &Context{
		stopCh: make(chan struct{}),
	}

	server.store = options.Store
	server.reconciler = NewReconciler(options.SyncTime, options.Store, options.Flush, server.stopCh)

	return server, nil
}

// Start the context.
func (s *Context) Start() error {
	// Fire off a pulse notifications sink goroutine.
	//go s.pulseHandler()
	if err := s.reconciler.Start(); err != nil {
		return err
	}
	return nil
}

// Close shuts down IPVS and closes the Context.
func (s *Context) Close() {
	log.Info("shutting down IPVS context")
	close(s.stopCh)
}

// CreateService registers a new virtual service with IPVS.
func (s *Context) CreateService(service *types.Service) error {
	s.Lock()
	defer s.Unlock()
	log.Infof("creating virtual service [%s] on %s:%d", service.StoreID, service.VIP, service.Port)

	// create service to external store
	//if ctx.store != nil {
	//	if err := ctx.store.CreateService(vsID, opts); err != nil {
	//		log.Errorf("error while create service : %s", err)
	//		return err
	//	}
	//}

	//var flags []string
	//if len(opts.Flags) > 0 {
	//	flags = strings.Split(opts.Flags, "|")
	//}
	//if err := ctx.ipvs.AddService(opts.host.String(), opts.Port, opts.Protocol, opts.Method, flags); err != nil {
	//	log.Errorf("error while creating virtual service: %s", err)
	//	return ErrIpvsSyscallFailed
	//}

	//ctx.services[vsID] = &service{options: opts}

	//if err := ctx.disco.Expose(vsID, opts.host.String(), opts.Port); err != nil {
	//	log.Errorf("error while exposing service to Disco: %s", err)
	//}

	return nil
}

// updateService updates a virtual service in IPVS.
func (s *Context) updateService(vsID string, opts *types.Service) error {
	if err := opts.Fill(s.endpoint); err != nil {
		return err
	}

	//old, exists := ctx.services[vsID]
	//if !exists {
	//	log.Infof("attempted to update a non-existent service [%s], will create instead", vsID)
	//	return ctx.createService(vsID, opts)
	//}

	// Check if not possible to update.
	//if old.options.Host != opts.Host ||
	//	old.options.Port != opts.Port ||
	//	old.options.Protocol != opts.Protocol {
	//	return fmt.Errorf("unable to update virtual service [%s] due to host/port/protocol changing", vsID)
	//}
	//
	//log.Infof("updating virtual service [%s] on %s:%d", vsID, opts.Host,
	//	opts.Port)

	// update service in external store
	if s.store != nil {
		if err := s.store.UpdateService(vsID, opts); err != nil {
			log.Errorf("error while updating service : %s", err)
			return err
		}
	}
	//
	//var flags []string
	//if len(opts.Flags) > 0 {
	//	flags = strings.Split(opts.Flags, "|")
	//}
	//if err := ctx.ipvs.UpdateService(opts.Host, opts.Port, opts.Protocol, opts.Method, flags); err != nil {
	//	log.Errorf("error while updating virtual service: %s", err)
	//	return ErrIpvsSyscallFailed
	//}
	//
	//ctx.services[vsID] = &service{options: opts}
	//
	//if err := ctx.disco.Expose(vsID, opts.Host, opts.Port); err != nil {
	//	log.Errorf("error while exposing service to Disco: %s", err)
	//}

	return nil
}

// CreateBackend registers a new backend with a virtual service.
func (s *Context) CreateBackend(vsID string, opts *types.Backend) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.createBackend(vsID, rsID, opts)
}

// CreateService registers a new virtual service with IPVS.
func (s *Context) UpdateService(opts *types.Service) error {
	s.Lock()
	defer s.Unlock()
	//if err := opts.Fill(); err != nil {
	//	return err
	//}
	//p, err := pulse.New(opts.Host, opts.Port, opts.Pulse)
	//if err != nil {
	//	return err
	//}
	//
	//if _, exists := ctx.backends[rsID]; exists {
	//	return ErrObjectExists
	//}
	//
	//vs, exists := ctx.services[vsID]
	//
	//if !exists {
	//	return ErrObjectNotFound
	//}

	//if util.AddrFamily(opts.HostIP()) != util.AddrFamily(vs.options.HostIP()) {
	//	return ErrIncompatibleAFs
	//}
	//
	//log.Infof("creating backend [%s] on %s:%d for virtual service [%s]",
	//	rsID,
	//	opts.Host,
	//	opts.Port,
	//	vsID)

	// create backend to external store
	if s.store != nil {
		if err := s.store.CreateBackend(vsID, rsID, opts); err != nil {
			log.Errorf("error while create backend : %s", err)
			return err
		}
	}

	//if err := ctx.ipvs.AddBackend(
	//	vs.options.Host,
	//	vs.options.Port,
	//	opts.Host,
	//	opts.Port,
	//	vs.options.Protocol,
	//	opts.Weight,
	//	opts.Method,
	//); err != nil {
	//	log.Errorf("error while creating backend: %s", err)
	//	return ErrIpvsSyscallFailed
	//}
	//
	//ctx.backends[rsID] = &backend{options: opts, service: vs, monitor: p}

	// Fire off the configured pulse goroutine, attach it to the Context.
	//go p.Loop(pulse.ID{VsID: vsID, RsID: rsID}, ctx.pulseCh, ctx.stopCh)

	return nil
}

// UpdateBackend updates the specified backend's weight.
func (s *Context) updateBackend(vsID, rsID string, weight uint32) (uint32, error) {
	rs, exists := s.backends[rsID]

	if !exists {
		return 0, ErrObjectNotFound
	}

	log.Infof("updating backend [%s/%s] with weight: %d", vsID, rsID,
		weight)

	//if err := ctx.ipvs.UpdateBackend(
	//	rs.service.options.Host,
	//	rs.service.options.Port,
	//	rs.options.Host,
	//	rs.options.Port,
	//	rs.service.options.Protocol,
	//	weight,
	//	rs.options.Method,
	//); err != nil {
	//	log.Errorf("error while updating backend [%s/%s]", vsID, rsID)
	//	return 0, ErrIpvsSyscallFailed
	//}

	var result uint32

	// Save the old backend weight and update the current backend weight.
	result, rs.options.Weight = rs.options.Weight, weight

	// Currently the backend options are changing only the weight.
	// The weight value is set to the value requested at the first setting,
	// and the weight value is updated when the pulse fails in the gorb.
	// In kvstore, it seems correct to record the request at the first setting and
	// not reflect the updated weight value.
	//if ctx.store != nil {
	//	ctx.store.UpdateBackend(vsID, rsID, rs.options)
	//}

	return result, nil
}

// UpdateBackend updates the specified backend's weight.
func (s *Context) UpdateBackend(vsID string, backend *types.Backend) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.updateBackend(vsID, rsID, weight)
}

// RemoveService deregisters a virtual service.
func (s *Context) removeService(vsID string) (*types.Service, error) {
	vs, exists := s.services[vsID]

	if !exists {
		return nil, ErrObjectNotFound
	}

	delete(s.services, vsID)

	//if ctx.vipInterface != nil && vs.options.delIfAddr == true {
	//	ifName := ctx.vipInterface.Attrs().Name
	//	vip := &netlink.Addr{IPNet: &net.IPNet{
	//		net.ParseIP(vs.options.host.String()), net.IPv4Mask(255, 255, 255, 255)}}
	//	if err := netlink.AddrDel(ctx.vipInterface, vip); err != nil {
	//		log.Infof(
	//			"failed to delete VIP %s to interface '%s' for service [%s]: %s",
	//			vs.options.host, ifName, vsID, err)
	//	}
	//	log.Infof("VIP %s has been deleted from interface '%s'", vs.options.host, ifName)
	//}
	//
	//log.Infof("removing virtual service [%s] from %s:%d", vsID,
	//	vs.options.host,
	//	vs.options.Port)
	//
	//if err := ctx.ipvs.DelService(
	//	vs.options.host.String(),
	//	vs.options.Port,
	//	vs.options.Protocol,
	//); err != nil {
	//	log.Errorf("error while removing virtual service [%s]", vsID)
	//	return nil, ErrIpvsSyscallFailed
	//}

	// delete service from external store
	if s.store != nil {
		if err := s.store.RemoveService(vsID); err != nil {
			log.Errorf("error while remove service : %s", err)
		}
	}

	for rsID, backend := range s.backends {
		if backend.service != vs {
			continue
		}

		log.Infof("cleaning up now orphaned backend [%s/%s]", vsID, rsID)

		// Stop the pulse goroutine.
		backend.monitor.Stop()

		delete(s.backends, rsID)

		// delete backend from external store
		if s.store != nil {
			s.store.RemoveBackend(rsID)
		}
	}

	// TODO(@kobolog): This will never happen in case of gorb-link.
	//if err := ctx.disco.Remove(vsID); err != nil {
	//	log.Errorf("error while removing service from Disco: %s", err)
	//}

	return vs.options, nil
}

// RemoveService deregisters a virtual service.
func (s *Context) RemoveService(vsID string) (*types.Service, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.removeService(vsID)
}

// RemoveBackend deregisters a backend.
func (s *Context) removeBackend(vsID, rsID string) (*types.Backend, error) {
	rs, exists := s.backends[rsID]

	if !exists {
		return nil, ErrObjectNotFound
	}

	log.Infof("removing backend [%s/%s]", vsID, rsID)

	// delete backend from external store
	if s.store != nil {
		if err := s.store.RemoveBackend(rsID); err != nil {
			log.Errorf("error while remove backend : %s", err)
		}
	}

	// Stop the pulse goroutine.
	rs.monitor.Stop()

	//if err := ctx.ipvs.DelBackend(
	//	rs.service.options.host.String(),
	//	rs.service.options.Port,
	//	rs.options.host.String(),
	//	rs.options.Port,
	//	rs.service.options.Protocol,
	//); err != nil {
	//	log.Errorf("error while removing backend [%s/%s]", vsID, rsID)
	//	return nil, ErrIpvsSyscallFailed
	//}

	delete(s.backends, rsID)

	return rs.options, nil
}

// RemoveBackend deregisters a backend.
func (s *Context) RemoveBackend(vsID, rsID string) (*types.Backend, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.removeBackend(vsID, rsID)
}

// ListServices returns a list of all registered services.
func (s *Context) ListServices() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	r := make([]string, 0, len(s.services))

	for vsID := range s.services {
		r = append(r, vsID)
	}

	return r, nil
}

// ServiceInfo contains information about virtual service options,
// its backends and overall virtual service health.
type ServiceInfo struct {
	Options  *types.Service `json:"options"`
	Health   float64        `json:"health"`
	Backends []string       `json:"backends"`
}

// GetService returns information about a virtual service.
func (s *Context) GetService(vsID string) (*ServiceInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	vs, exists := s.services[vsID]

	if !exists {
		return nil, ErrObjectNotFound
	}

	result := ServiceInfo{Options: vs.options}

	// This is O(n), can be optimized with reverse backend map.
	for rsID, backend := range s.backends {
		if backend.service != vs {
			continue
		}

		result.Backends = append(result.Backends, rsID)
		result.Health += backend.metrics.Health
	}

	if len(result.Backends) == 0 {
		// Service without backends is healthy, albeit useless.
		result.Health = 1.0
	} else {
		result.Health /= float64(len(result.Backends))
	}

	return &result, nil
}

// BackendInfo contains information about backend options and pulse.
type BackendInfo struct {
	Options *types.Backend `json:"options"`
	Metrics pulse.Metrics  `json:"metrics"`
}

// GetBackend returns information about a backend.
func (s *Context) GetBackend(vsID, rsID string) (*BackendInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rs, exists := s.backends[rsID]

	if !exists {
		return nil, ErrObjectNotFound
	}

	return &BackendInfo{rs.options, rs.metrics}, nil
}

//func (ctx *Context) Synchronize(storeServices map[string]*options.Service, storeBackends map[string]*options.Backend) {
//	ctx.mutex.Lock()
//	defer ctx.mutex.Unlock()
//
//	log.Debugf("============================== SYNC ========================================")
//	for k, v := range storeServices {
//		log.Debugf("SERVICE[%s]: %s", k, v)
//	}
//	for k, v := range storeBackends {
//		log.Debugf("  BACKEND[%s]: %s", k, v)
//	}
//	defer log.Debugf("============================================================================")
//
//	// synchronize services with store
//	for id, _ := range ctx.services {
//		if _, ok := storeServices[id]; !ok {
//			ctx.removeService(id)
//		}
//	}
//	for id, storeServiceOptions := range storeServices {
//		if service, ok := ctx.services[id]; ok {
//			if service.options.CompareStoreOptions(storeServiceOptions) {
//				continue
//			}
//			ctx.removeService(id)
//		}
//		ctx.createService(id, storeServiceOptions)
//	}
//
//	// synchronize backends with store
//	for id, backend := range ctx.backends {
//		if _, ok := storeBackends[id]; !ok {
//			vsID := "(unknown)"
//			if len(backend.options.VsID) > 0 {
//				vsID = backend.options.VsID
//			}
//			ctx.removeBackend(vsID, id)
//		}
//	}
//	for id, storeBackendOptions := range storeBackends {
//		if backend, ok := ctx.backends[id]; ok {
//			if backend.options.CompareStoreOptions(storeBackendOptions) {
//				continue
//			}
//			ctx.removeBackend(storeBackendOptions.VsID, id)
//		}
//		if err := ctx.createBackend(storeBackendOptions.VsID, id, storeBackendOptions); err != nil {
//			log.Warnf("create backend error: %s", err.Error())
//		}
//	}
//}
