/*
   Copyright (c) 2018 Contributors as noted in the AUTHORS file.

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
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/kobolog/gorb/ipvs-shim"
)

type reconciler struct {
	period time.Duration
	syncCh chan struct{}
	store  *Store
	ipvs   ipvs_shim.IPVS
}

// New returns a reconciler that populates the ipvs state periodically and on demand.
func NewReconciler(period time.Duration, store *Store) *reconciler {
	return &reconciler{
		period: period,
		syncCh: make(chan struct{}),
		store:  store,
	}
}

func (r *reconciler) Start() {
	go func() {
		for {
			t := time.NewTimer(r.period)
			select {
			case <-t.C:
				r.reconcile()
			case <-r.syncCh:
				r.reconcile()
			}
		}
	}()
}

func (r *reconciler) Sync() {
	r.syncCh <- struct{}{}
}

func (r *reconciler) ListServices() ([]*ServiceOptions, error) {
	ipvsSvcs, err := r.ipvs.ListServices()
	if err != nil {
		return nil, err
	}
	var svcs []*ServiceOptions
	for _, is := range ipvsSvcs {
		svc := &ServiceOptions{
			Host:     is.VIP,
			Port:     is.Port,
			Protocol: is.Protocol,
			Method:   is.Scheduler,
			Flags:    strings.Join(is.Flags, "|"),
		}
		svcs = append(svcs, svc)
	}
	return svcs, nil
}

func (r *reconciler) ListBackends(key *ServiceOptions) ([]*BackendOptions, error) {
	ipvsBackends, err := r.ipvs.ListBackends(nil)
	if err != nil {
		return nil, err
	}
	var backends []*BackendOptions
	for _, ib := range ipvsBackends {
		backend := &BackendOptions{
			Host:   ib.IP,
			Port:   ib.Port,
			Method: ib.Forward,
			Weight: ib.Weight,
		}
		backends = append(backends, backend)
	}
	return backends, nil
}

func (r *reconciler) reconcile() {
	desiredServices, err := r.store.ListServices()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}
	desiredBackends, err := r.store.ListBackends()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}

	for _, v := range desiredServices {
		log.Debugf("SERVICE: %s", v)
	}
	for _, v := range desiredBackends {
		log.Debugf("  BACKEND: %s", v)
	}

	actualServices, err := r.ListServices()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}

	for _, actualService := range actualServices {
		actualBackends, err := r.ListBackends()
		if err != nil {
			log.Errorf("unable to populate: %v", err)
			return
		}
	}



	// synchronize services with store
	//for id, _ := range ctx.services {
	//	if _, ok := storeServices[id]; !ok {
	//		ctx.removeService(id)
	//	}
	//}
	//for id, storeServiceOptions := range storeServices {
	//	if service, ok := ctx.services[id]; ok {
	//		if service.options.CompareStoreOptions(storeServiceOptions) {
	//			continue
	//		}
	//		ctx.removeService(id)
	//	}
	//	ctx.createService(id, storeServiceOptions)
	//}
	//
	//// synchronize backends with store
	//for id, backend := range ctx.backends {
	//	if _, ok := storeBackends[id]; !ok {
	//		vsID := "(unknown)"
	//		if len(backend.options.VsID) > 0 {
	//			vsID = backend.options.VsID
	//		}
	//		ctx.removeBackend(vsID, id)
	//	}
	//}
	//for id, storeBackendOptions := range storeBackends {
	//	if backend, ok := ctx.backends[id]; ok {
	//		if backend.options.CompareStoreOptions(storeBackendOptions) {
	//			continue
	//		}
	//		ctx.removeBackend(storeBackendOptions.VsID, id)
	//	}
	//	if err := ctx.createBackend(storeBackendOptions.VsID, id, storeBackendOptions); err != nil {
	//		log.Warnf("create backend error: %s", err.Error())
	//	}
	//}
}
