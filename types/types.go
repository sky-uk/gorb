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

package types

import (
	"net"

	"fmt"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/kobolog/gorb/pulse"
)

type ServiceKey struct {
	VIP      net.IP `json:"vip"`
	Port     uint16 `json:"port"`
	Protocol string `json:"protocol"`
}

// Service describes a virtual service.
type Service struct {
	ServiceKey
	Scheduler string   `json:"scheduler"`
	Flags     []string `json:"flags"`
	// StoreID uniquely identifies the virtual service in the store. It's optional and unused by ipvs.
	StoreID string `json:"id"`
}

func (s *Service) String() string {
	return fmt.Sprintf("%s:%d (%s) [%s (%s)]", s.VIP.String(), s.Port, s.Protocol, s.Scheduler,
		strings.Join(s.Flags, ","))
}

func (s *Service) HostIP() net.IP {
	return s.VIP
}

func (s *Service) Flagset() mapset.Set {
	fs := mapset.NewThreadUnsafeSet()
	for _, f := range s.Flags {
		fs.Add(f)
	}
	return fs
}

func (s *ServiceKey) Equal(other ServiceKey) bool {
	return s.VIP.Equal(other.VIP) &&
		s.Port == other.Port &&
		s.Protocol == other.Protocol
}

func (s *Service) Equal(other *Service) bool {
	return s.ServiceKey.Equal(other.ServiceKey) &&
		s.Flagset().Equal(other.Flagset()) &&
		s.Scheduler == other.Scheduler
}

type BackendKey struct {
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
}

// Backend describe a virtual service backend.
type Backend struct {
	BackendKey
	Weight  uint32 `json:"weight"`
	Forward string `json:"forward"`
	// StoreID uniquely identifies the backend in the store. It's optional and unused by ipvs.
	StoreID string `json:"id"`
	// Pulse is optional and unused by ipvs.
	PulseOptions *pulse.Options `json:"pulse"`
}

func (b *BackendKey) Equal(o BackendKey) bool {
	return b.IP.Equal(o.IP) && b.Port == o.Port
}

func (b *Backend) Equal(o *Backend) bool {
	return b.BackendKey.Equal(o.BackendKey) && b.Weight == o.Weight && b.Forward == o.Forward
}
