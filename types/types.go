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
	"errors"
	"net"

	"fmt"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/kobolog/gorb/pulse"
)

// Possible validation errors.
var (
	ErrMissingEndpoint = errors.New("endpoint information is missing")
	ErrUnknownMethod   = errors.New("specified forwarding method is unknown")
	ErrUnknownProtocol = errors.New("specified protocol is unknown")
	ErrUnknownFlag     = errors.New("specified flag is unknown")
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

// Fill missing fields and validates virtual service configuration.
func (s *Service) Fill(defaultHost net.IP) error {
	//if o.Port == 0 {
	//	return ErrMissingEndpoint
	//}

	//if len(o.Host) != 0 {
	//	if addr, err := net.ResolveIPAddr("ip", o.Host); err == nil {
	//		o.Host = addr.IP.String()
	//	} else {
	//		return err
	//	}
	//} else if defaultHost != nil {
	//	o.Host = defaultHost.String()
	//} else {
	//	return ErrMissingEndpoint
	//}

	//if len(o.Protocol) == 0 {
	//	o.Protocol = "tcp"
	//}
	//
	//o.Protocol = strings.ToLower(o.Protocol)
	//if !ipvs_shim.ValidProtocol(o.Protocol) {
	//	return ErrUnknownProtocol
	//}
	//
	//if len(o.Flags) != 0 {
	//	for _, flag := range o.Flags {
	//		if ok := ipvs_shim.ValidFlag(flag); !ok {
	//			return ErrUnknownFlag
	//		}
	//	}
	//}
	//
	//if len(o.Scheduler) == 0 {
	//	// WRR since Pulse will dynamically reweight backends.
	//	o.Scheduler = "wrr"
	//}

	return nil
}

func (s *ServiceKey) Equal(other *ServiceKey) bool {
	return s.VIP.Equal(other.VIP) &&
		s.Port == other.Port &&
		s.Protocol == other.Protocol
}

func (s *Service) Equal(other *Service) bool {
	return s.ServiceKey.Equal(&other.ServiceKey) &&
		s.Flagset().Equal(other.Flagset()) &&
		s.Scheduler == other.Scheduler
}

// Backend describe a virtual service backend.
type Backend struct {
	IP      net.IP `json:"ip"`
	Port    uint16 `json:"port"`
	Weight  uint32 `json:"weight"`
	Forward string `json:"forward"`
	// Pulse is optional and unused by ipvs.
	Pulse *pulse.Options `json:"pulse,omitempty"`
}

func (b *Backend) EqualKey(o *Backend) bool {
	return b.IP.Equal(o.IP) && b.Port == o.Port
}

func (b *Backend) Equal(o *Backend) bool {
	return b.EqualKey(o) && b.Weight == o.Weight && b.Forward == o.Forward
}

// Fill missing fields and validates backend configuration.
func (o *Backend) Fill() error {
	//if len(o.Host) == 0 || o.Port == 0 {
	//	return ErrMissingEndpoint
	//}
	//
	//if addr, err := net.ResolveIPAddr("ip", o.Host); err == nil {
	//	o.Host = addr.IP.String()
	//} else {
	//	return err
	//}
	//
	//if o.Weight <= 0 {
	//	o.Weight = 100
	//}
	//
	//if len(o.Method) == 0 {
	//	o.Method = "nat"
	//}
	//
	//o.Method = strings.ToLower(o.Method)
	//if !ipvs_shim.ValidForwarding(o.Method) {
	//	return ErrUnknownMethod
	//}
	//
	//if o.Pulse == nil {
	//	// It doesn't make much sense to have a backend with no Pulse.
	//	o.Pulse = &pulse.Options{}
	//}

	return nil
}

func (o *Backend) CompareStoreOptions(options *Backend) bool {
	//if o.Host != options.Host {
	//	return false
	//}
	//if o.Port != options.Port {
	//	return false
	//}
	//if o.Method != options.Method {
	//	return false
	//}
	return true
}
