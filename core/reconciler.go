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

	log "github.com/Sirupsen/logrus"
	"github.com/kobolog/gorb/ipvs-shim"
	"github.com/kobolog/gorb/types"
)

type reconciler struct {
	period time.Duration
	syncCh chan struct{}
	store  reconcilerStore
	ipvs   ipvs_shim.IPVS
}

type reconcilerStore interface {
	ListServices() ([]*types.Service, error)
	ListBackends(vsID string) ([]*types.Backend, error)
}

// New returns a reconciler that populates the ipvs state periodically and on demand.
func NewReconciler(period time.Duration, store reconcilerStore) *reconciler {
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

func (r *reconciler) reconcile() {
	desiredServices, err := r.store.ListServices()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}
	for _, s := range desiredServices {
		log.Infof("DESIRED SERVICE: %v", *s)
	}

	actualServices, err := r.ipvs.ListServices()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}
	for _, s := range actualServices {
		log.Infof("ACTUAL SERVICE: %v", *s)
	}

	for _, desired := range desiredServices {
		var match *types.Service
		for _, actual := range actualServices {
			if desired.ServiceKey.Equal(&actual.ServiceKey) {
				match = actual
				break
			}
		}
		if match == nil {
			r.ipvs.AddService(desired)
		} else if !desired.Equal(match) {
			r.ipvs.UpdateService(desired)
		}
	}

	//desiredBackends, err := r.store.ListBackends("fixme")
	//if err != nil {
	//	log.Errorf("unable to populate: %v", err)
	//	return
	//}
	//for _, v := range desiredServices {
	//	log.Debugf("SERVICE: %s", v)
	//}
	//for _, v := range desiredBackends {
	//	log.Debugf("  BACKEND: %s", v)
	//}

	//for _, actualService := range actualServices {
	//	actualBackends, err := r.ListBackends()
	//	if err != nil {
	//		log.Errorf("unable to populate: %v", err)
	//		return
	//	}
	//}

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
