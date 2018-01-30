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
)

type populator struct {
	period time.Duration
	syncCh chan struct{}
}

// New returns a populator that populates the ipvs state periodically and on demand.
func NewPopulator(period time.Duration) *populator {
	return &populator{
		period: period,
		syncCh: make(chan struct{}),
	}
}

func (p *populator) Start() {
	go func() {
		for {
			t := time.NewTimer(p.period)
			select {
			case <-t.C:
				p.populate()
			case <-p.syncCh:
				p.populate()
			}
		}
	}()
}

func (p *populator) Sync() {
	p.syncCh <- struct{}{}
}

func (p *populator) populate() {
	log.Info("populate stuff!")
}
