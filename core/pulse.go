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
	"github.com/kobolog/gorb/pulse"

	log "github.com/Sirupsen/logrus"
)

func (s *Context) pulseHandler() {
	stash := make(map[pulse.ID]uint32)

	for {
		select {
		case u := <-s.pulseCh:
			s.processPulseUpdate(stash, u)
		case <-s.stopCh:
			log.Debug("notificationLoop has been stopped")
			return
		}
	}
}

func (s *Context) processPulseUpdate(stash map[pulse.ID]uint32, u pulse.Update) {
	vsID, rsID := u.Source.VsID, u.Source.RsID

	s.mutex.Lock()

	// check exist
	if _, ok := s.backends[rsID]; !ok || u.Metrics.Status == pulse.StatusRemoved {
		if _, exists := stash[u.Source]; exists {
			log.Debugf("backend %s has been deleted, so deleting it from stash too", u.Source)
			delete(stash, u.Source)
		}
		s.mutex.Unlock()
		return
	}

	if s.backends[rsID].metrics.Status != u.Metrics.Status {
		log.Warnf("backend %s status: %s", u.Source, u.Metrics.Status)
	}

	// This is a copy of metrics structure from Pulse.
	s.backends[rsID].metrics = u.Metrics

	s.mutex.Unlock()

	switch u.Metrics.Status {
	case pulse.StatusUp:
		// Weight is gonna be stashed until the backend is recovered.
		weight, exists := stash[u.Source]

		if !exists {
			return
		}

		// Calculate a relative weight considering backend's health.
		weight = uint32(float64(weight) * u.Metrics.Health)

		if _, err := s.UpdateBackend(vsID, rsID, weight); err != nil {
			log.Errorf("error while unstashing a backend: %s", err)
		} else if weight == stash[u.Source] {
			log.Debugf("backend %s has completely recovered, so deleting it from stash.", u.Source)
			// This means that the backend has completely recovered.
			delete(stash, u.Source)
		}

	case pulse.StatusDown:
		if _, exists := stash[u.Source]; exists {
			return
		}

		if weight, err := s.UpdateBackend(vsID, rsID, 0); err != nil {
			log.Errorf("error while stashing a backend: %s", err)
		} else {
			stash[u.Source] = weight
		}
	}
}
