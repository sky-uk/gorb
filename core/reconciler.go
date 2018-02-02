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

	actualServices, err := r.ipvs.ListServices()
	if err != nil {
		log.Errorf("unable to populate: %v", err)
		return
	}

	// create or update services
	for _, desired := range desiredServices {
		var match *types.Service
		for _, actual := range actualServices {
			if desired.ServiceKey.Equal(&actual.ServiceKey) {
				match = actual
				break
			}
		}
		if match == nil {
			log.Infof("Adding new service: %v", desired)
			r.ipvs.AddService(desired)
		} else if !desired.Equal(match) {
			log.Infof("Updating service: %v", desired)
			r.ipvs.UpdateService(desired)
		}

		desiredBackends, err := r.store.ListBackends(desired.StoreID)
		if err != nil {
			log.Errorf("unable to list backends in store: %v", desired.StoreID, err)
			continue
		}
		actualBackends, err := r.ipvs.ListBackends(&desired.ServiceKey)
		if err != nil {
			log.Errorf("unable to list backends in ipvs: %v")
			continue
		}

		for _, desiredBackend := range desiredBackends {
			var match *types.Backend
			for _, actual := range actualBackends {
				if desiredBackend.EqualKey(actual) {
					match = actual
					break
				}
			}
			if match == nil {
				r.ipvs.AddBackend(&desired.ServiceKey, desiredBackend)
			} else if !desiredBackend.Equal(match) {
				r.ipvs.UpdateBackend(&desired.ServiceKey, desiredBackend)
			}
		}
	}

	// delete services
	for _, actual := range actualServices {
		var found bool
		for _, desired := range desiredServices {
			if actual.ServiceKey.Equal(&desired.ServiceKey) {
				found = true
				break
			}
		}
		if !found {
			r.ipvs.DeleteService(&actual.ServiceKey)
		}
	}
}
