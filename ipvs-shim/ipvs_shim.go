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

	forwardingMethods = map[string]uint32{
		"dr":     libipvs.IP_VS_CONN_F_DROUTE,
		"nat":    libipvs.IP_VS_CONN_F_MASQ,
		"tunnel": libipvs.IP_VS_CONN_F_TUNNEL,
		"ipip":   libipvs.IP_VS_CONN_F_TUNNEL,
	}
)

type IPVS interface {
	Init() error
	Flush() error
	AddService(svc *Service) error
	UpdateService(svc *Service) error
	DelService(key *ServiceKey) error
	ListServices() ([]*Service, error)
	AddBackend(key *ServiceKey, backend *Backend) error
	UpdateBackend(key *ServiceKey, backend *Backend) error
	DelBackend(key *ServiceKey, backend *Backend) error
}

type ServiceKey struct {
	VIP      string
	Port     uint16
	Protocol string
}

type Service struct {
	ServiceKey
	Scheduler string
	Flags     []string
}

type Backend struct {
	IP      string
	Port    uint16
	Weight  uint32
	Forward string
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
	_, exists := forwardingMethods[fwd]
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

func initIPVSService(key *ServiceKey) (*libipvs.Service, error) {
	protNum, err := protocolNumber(key.Protocol)
	if err != nil {
		return nil, err
	}
	svc := &libipvs.Service{
		Address:       net.ParseIP(key.VIP),
		Protocol:      libipvs.Protocol(protNum),
		Port:          key.Port,
		AddressFamily: syscall.AF_INET,
	}
	return svc, nil
}

func createFlagbits(flags []string) uint32 {
	var flagbits uint32
	for _, flag := range flags {
		if b, exists := schedulerFlags[flag]; exists {
			flagbits |= b
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

func (s *shim) AddService(svc *Service) error {
	ipvsSvc, err := initIPVSService(&svc.ServiceKey)
	if err != nil {
		return err
	}
	ipvsSvc.SchedName = svc.Scheduler
	ipvsSvc.Flags.Flags = createFlagbits(svc.Flags)
	ipvsSvc.Flags.Mask = ^uint32(0)
	return s.handle.NewService(ipvsSvc)
}

func (s *shim) UpdateService(svc *Service) error {
	ipvsSvc, err := initIPVSService(&svc.ServiceKey)
	if err != nil {
		return err
	}
	ipvsSvc.SchedName = svc.Scheduler
	ipvsSvc.Flags.Flags = createFlagbits(svc.Flags)
	ipvsSvc.Flags.Mask = ^uint32(0)
	return s.handle.UpdateService(ipvsSvc)
}

func (s *shim) DelService(key *ServiceKey) error {
	svc, err := initIPVSService(key)
	if err != nil {
		return err
	}
	return s.handle.DelService(svc)
}

func convertFlagbits(flagbits uint32) []string {
	var flags []string
	var possibleFlags = map[uint32]string{
		libipvs.IP_VS_SVC_F_SCHED1: "flag-1",
		libipvs.IP_VS_SVC_F_SCHED2: "flag-2",
		libipvs.IP_VS_SVC_F_SCHED3: "flag-3",
	}
	for f, v := range possibleFlags {
		if flagbits&f != 0 {
			flags = append(flags, v)
		}
	}
	return flags
}

func convertShFlagbits(flagbits uint32) []string {
	var flags []string
	var possibleFlags = map[uint32]string{
		libipvs.IP_VS_SVC_F_SCHED_SH_FALLBACK: "sh-fallback",
		libipvs.IP_VS_SVC_F_SCHED_SH_PORT:     "sh-port",
	}
	for f, v := range possibleFlags {
		if flagbits&f != 0 {
			flags = append(flags, v)
		}
	}
	return flags
}

func (s *shim) ListServices() ([]*Service, error) {
	ipvsSvcs, err := s.handle.ListServices()
	if err != nil {
		return nil, err
	}

	var svcs []*Service
	for _, isvc := range ipvsSvcs {
		var flags []string
		if isvc.SchedName == "sh" {
			flags = convertShFlagbits(isvc.Flags.Flags)
		} else {
			flags = convertFlagbits(isvc.Flags.Flags)
		}

		svc := &Service{
			ServiceKey: ServiceKey{
				VIP:      isvc.Address.String(),
				Port:     isvc.Port,
				Protocol: isvc.Protocol.String(),
			},
			Scheduler: isvc.SchedName,
			Flags:     flags,
		}

		svcs = append(svcs, svc)
	}

	return svcs, nil
}

func createDest(backend *Backend, full bool) (*libipvs.Destination, error) {
	dest := &libipvs.Destination{
		Address:       net.ParseIP(backend.IP),
		Port:          backend.Port,
		AddressFamily: syscall.AF_INET,
	}
	if !full {
		return dest, nil
	}
	fwdbits, ok := forwardingMethods[backend.Forward]
	if !ok {
		return nil, fmt.Errorf("invalid forwarding method %q", backend.Forward)
	}
	dest.FwdMethod = libipvs.FwdMethod(fwdbits)
	dest.Weight = backend.Weight
	return dest, nil
}

func (s *shim) AddBackend(key *ServiceKey, backend *Backend) error {
	svc, err := initIPVSService(key)
	if err != nil {
		return err
	}
	dest, err := createDest(backend, true)
	if err != nil {
		return err
	}
	return s.handle.NewDestination(svc, dest)
}

func (s *shim) UpdateBackend(key *ServiceKey, backend *Backend) error {
	svc, err := initIPVSService(key)
	if err != nil {
		return err
	}
	dest, err := createDest(backend, true)
	if err != nil {
		return err
	}
	return s.handle.UpdateDestination(svc, dest)
}

func (s *shim) DelBackend(key *ServiceKey, backend *Backend) error {
	svc, err := initIPVSService(key)
	if err != nil {
		return err
	}
	dest, err := createDest(backend, false)
	if err != nil {
		return err
	}
	return s.handle.DelDestination(svc, dest)
}
