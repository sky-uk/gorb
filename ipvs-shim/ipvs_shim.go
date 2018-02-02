// ipvs_shim encapsulates the details of the ipvs/netlink library.
package ipvs_shim

import (
	"syscall"

	"fmt"

	"github.com/kobolog/gorb/types"
	"github.com/mqliang/libipvs"
)

var (
	schedulerFlags = map[string]uint32{
		"flag-1": libipvs.IP_VS_SVC_F_SCHED1,
		"flag-2": libipvs.IP_VS_SVC_F_SCHED2,
		"flag-3": libipvs.IP_VS_SVC_F_SCHED3,
	}
	schedulerFlagsInverted map[uint32]string

	forwardingMethods = map[string]libipvs.FwdMethod{
		"dr":     libipvs.IP_VS_CONN_F_DROUTE,
		"nat":    libipvs.IP_VS_CONN_F_MASQ,
		"tunnel": libipvs.IP_VS_CONN_F_TUNNEL,
	}
	forwardingMethodsInverted map[libipvs.FwdMethod]string
)

func init() {
	schedulerFlagsInverted = make(map[uint32]string)
	for k, v := range schedulerFlags {
		schedulerFlagsInverted[v] = k
	}
	forwardingMethodsInverted = make(map[libipvs.FwdMethod]string)
	for k, v := range forwardingMethods {
		forwardingMethodsInverted[v] = k
	}
}

type IPVS interface {
	Init() error
	Flush() error
	AddService(svc *types.Service) error
	UpdateService(svc *types.Service) error
	DeleteService(key *types.ServiceKey) error
	ListServices() ([]*types.Service, error)
	AddBackend(key *types.ServiceKey, backend *types.Backend) error
	UpdateBackend(key *types.ServiceKey, backend *types.Backend) error
	DeleteBackend(key *types.ServiceKey, backend *types.Backend) error
	ListBackends(key *types.ServiceKey) ([]*types.Backend, error)
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

func initIPVSService(key *types.ServiceKey) (*libipvs.Service, error) {
	protNum, err := protocolNumber(key.Protocol)
	if err != nil {
		return nil, err
	}
	svc := &libipvs.Service{
		Address:       key.VIP,
		Protocol:      libipvs.Protocol(protNum),
		Port:          key.Port,
		AddressFamily: syscall.AF_INET,
	}
	return svc, nil
}

func createFlagbits(flags []string) (libipvs.Flags, error) {
	var flagbits uint32
	for _, flag := range flags {
		if b, exists := schedulerFlags[flag]; exists {
			flagbits |= b
		} else {
			return libipvs.Flags{}, fmt.Errorf("unknown scheduler flag %q, ignoring", flag)
		}
	}
	r := libipvs.Flags{
		Flags: flagbits,
		// set all bits to 1
		Mask: ^uint32(0),
	}
	return r, nil
}

func (s *shim) AddService(svc *types.Service) error {
	ipvsSvc, err := initIPVSService(&svc.ServiceKey)
	if err != nil {
		return err
	}
	ipvsSvc.SchedName = svc.Scheduler
	ipvsSvc.Flags, err = createFlagbits(svc.Flags)
	if err != nil {
		return err
	}
	return s.handle.NewService(ipvsSvc)
}

func (s *shim) UpdateService(svc *types.Service) error {
	ipvsSvc, err := initIPVSService(&svc.ServiceKey)
	if err != nil {
		return err
	}
	ipvsSvc.SchedName = svc.Scheduler
	ipvsSvc.Flags, err = createFlagbits(svc.Flags)
	if err != nil {
		return err
	}
	return s.handle.UpdateService(ipvsSvc)
}

func (s *shim) DeleteService(key *types.ServiceKey) error {
	svc, err := initIPVSService(key)
	if err != nil {
		return err
	}
	return s.handle.DelService(svc)
}

func convertFlagbits(flagbits uint32) []string {
	var flags []string
	for f, v := range schedulerFlagsInverted {
		if flagbits&f != 0 {
			flags = append(flags, v)
		}
	}
	return flags
}

func (s *shim) ListServices() ([]*types.Service, error) {
	ipvsSvcs, err := s.handle.ListServices()
	if err != nil {
		return nil, err
	}

	var svcs []*types.Service
	for _, isvc := range ipvsSvcs {
		svc := &types.Service{
			ServiceKey: types.ServiceKey{
				VIP:      isvc.Address,
				Port:     isvc.Port,
				Protocol: isvc.Protocol.String(),
			},
			Scheduler: isvc.SchedName,
			Flags:     convertFlagbits(isvc.Flags.Flags),
		}

		svcs = append(svcs, svc)
	}

	return svcs, nil
}

func createDest(backend *types.Backend, full bool) (*libipvs.Destination, error) {
	dest := &libipvs.Destination{
		Address:       backend.IP,
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

func (s *shim) AddBackend(key *types.ServiceKey, backend *types.Backend) error {
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

func (s *shim) UpdateBackend(key *types.ServiceKey, backend *types.Backend) error {
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

func (s *shim) DeleteBackend(key *types.ServiceKey, backend *types.Backend) error {
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

func (s *shim) ListBackends(key *types.ServiceKey) ([]*types.Backend, error) {
	svc, err := initIPVSService(key)
	if err != nil {
		return nil, err
	}

	dests, err := s.handle.ListDestinations(svc)
	if err != nil {
		return nil, err
	}

	var backends []*types.Backend
	for _, dest := range dests {
		fwd, ok := forwardingMethodsInverted[dest.FwdMethod]
		if !ok {
			return nil, fmt.Errorf("unable to list backends, unexpected forward method %#x", dest.FwdMethod)
		}
		backend := &types.Backend{
			BackendKey: types.BackendKey{
				IP:   dest.Address,
				Port: dest.Port,
			},
			Weight:  dest.Weight,
			Forward: fwd,
		}
		backends = append(backends, backend)
	}

	return backends, nil
}
