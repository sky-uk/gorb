// ipvs_shim encapsulates the details of the ipvs/netlink library.
package ipvs_shim

import (
	"net"
	"syscall"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mqliang/libipvs"
)

var (
	schedulerFlags = map[string]uint32{
		"sh-fallback": libipvs.IP_VS_SVC_F_SCHED_SH_FALLBACK,
		"sh-port":     libipvs.IP_VS_SVC_F_SCHED_SH_PORT,
		"flag-1":      libipvs.IP_VS_SVC_F_SCHED1,
		"flag-2":      libipvs.IP_VS_SVC_F_SCHED2,
		"flag-3":      libipvs.IP_VS_SVC_F_SCHED3,
	}

	backendForwarding = map[string]uint32{
		"dr":     libipvs.IP_VS_CONN_F_DROUTE,
		"nat":    libipvs.IP_VS_CONN_F_MASQ,
		"tunnel": libipvs.IP_VS_CONN_F_TUNNEL,
		"ipip":   libipvs.IP_VS_CONN_F_TUNNEL,
	}
)

type IPVS interface {
	Init() error
	Flush() error
	AddService(vip string, port uint16, protocol string, sched string, flags []string) error
	UpdateService(vip string, port uint16, protocol string, sched string, flags []string) error
	DelService(vip string, port uint16, protocol string) error
	AddDestPort(vip string, vport uint16, rip string, rport uint16, protocol string, weight uint32, fwd string) error
	UpdateDestPort(vip string, vport uint16, rip string, rport uint16, protocol string, weight uint32, fwd string) error
	DelDestPort(vip string, vport uint16, rip string, rport uint16, protocol string) error
}

type shim struct {
	handle libipvs.IPVSHandle
}

func New() IPVS {
	return &shim{}
}

func ValidProtocol(protocol string) bool {
	if _, err := protocolNumber(protocol); err != nil {
		return false
	}
	return true
}

func ValidFlag(flag string) bool {
	_, exists := schedulerFlags[flag]
	return exists
}

func ValidForwarding(fwd string) bool {
	_, exists := backendForwarding[fwd]
	return exists
}

func (s *shim) Init() error {
	h, err := libipvs.New()
	if err != nil {
		return err
	}
	s.handle = h
	return nil
}

func (s *shim) Flush() error {
	return s.handle.Flush()
}

func createSvcKey(vip string, protocol string, port uint16) (*libipvs.Service, error) {
	protNum, err := protocolNumber(protocol)
	if err != nil {
		return nil, err
	}
	svc := &libipvs.Service{
		Address:       net.ParseIP(vip),
		Protocol:      libipvs.Protocol(protNum),
		Port:          port,
		AddressFamily: syscall.AF_INET,
	}
	return svc, nil
}

func createFlagbits(flags []string) uint32 {
	var flagbits uint32
	for _, flag := range flags {
		if b, exists := schedulerFlags[flag]; exists {
			// libipvs incorrectly repacks flags as big endian. This is a hack to ensure the flag bytes get passed
			// correctly. It assumes the scheduler flags all fit into a single byte (which they do).
			flagbits |= b << 24
		} else {
			log.Warnf("Unknown scheduler flag %q, ignoring", flag)
		}
	}
	return flagbits
}

func protocolNumber(protocol string) (uint16, error) {
	switch protocol {
	case "tcp":
		return syscall.IPPROTO_TCP, nil
	case "udp":
		return syscall.IPPROTO_UDP, nil
	default:
		return 0, fmt.Errorf("unknown protocol %q", protocol)
	}
}

func (s *shim) AddService(vip string, port uint16, protocol string, sched string, flags []string) error {
	log.Infof("flags: %v", flags)
	svc, err := createSvcKey(vip, protocol, port)
	if err != nil {
		return err
	}
	svc.SchedName = sched
	svc.Flags.Flags = createFlagbits(flags)
	svc.Flags.Mask = ^uint32(0)
	return s.handle.NewService(svc)
}

func (s *shim) UpdateService(vip string, port uint16, protocol string, sched string, flags []string) error {
	svc, err := createSvcKey(vip, protocol, port)
	if err != nil {
		return err
	}
	svc.SchedName = sched
	svc.Flags.Flags = createFlagbits(flags)
	svc.Flags.Mask = ^uint32(0)
	return s.handle.UpdateService(svc)
}

func (s *shim) DelService(vip string, port uint16, protocol string) error {
	svc, err := createSvcKey(vip, protocol, port)
	if err != nil {
		return err
	}
	return s.handle.DelService(svc)
}

func createDest(rip string, rport uint16, fwd uint32, weight uint32) *libipvs.Destination {
	dest := &libipvs.Destination{
		Address:       net.ParseIP(rip),
		Port:          rport,
		AddressFamily: syscall.AF_INET,
		FwdMethod:     libipvs.FwdMethod(fwd),
		Weight:        weight,
	}
	return dest
}

func (s *shim) AddDestPort(vip string, vport uint16, rip string, rport uint16, protocol string, weight uint32, fwd string) error {
	svc, err := createSvcKey(vip, protocol, vport)
	if err != nil {
		return err
	}
	fwdbits, ok := backendForwarding[fwd]
	if !ok {
		return fmt.Errorf("invalid forwarding method %q", fwd)
	}
	dest := createDest(rip, rport, fwdbits, weight)
	return s.handle.NewDestination(svc, dest)
}

func (s *shim) UpdateDestPort(vip string, vport uint16, rip string, rport uint16, protocol string, weight uint32, fwd string) error {
	svc, err := createSvcKey(vip, protocol, vport)
	if err != nil {
		return err
	}
	fwdbits, ok := backendForwarding[fwd]
	if !ok {
		return fmt.Errorf("invalid forwarding method %q", fwd)
	}
	dest := createDest(rip, rport, fwdbits, weight)
	return s.handle.UpdateDestination(svc, dest)
}

func (s *shim) DelDestPort(vip string, vport uint16, rip string, rport uint16, protocol string) error {
	svc, err := createSvcKey(vip, protocol, vport)
	if err != nil {
		return err
	}
	dest := createDest(rip, rport, 0, 0)
	return s.handle.DelDestination(svc, dest)
}
