package core

import (
	"testing"

	"net"

	"github.com/kobolog/gorb/pulse"
	"github.com/kobolog/gorb/types"
)

func TestCollector(t *testing.T) {
	ctx := &Context{
		services: make(map[string]*service),
		backends: make(map[string]*backend),
	}
	ctx.services["service1"] = &service{options: &types.Service{
		ServiceKey: types.ServiceKey{
			VIP:      net.ParseIP("127.0.0.1"),
			Port:     1234,
			Protocol: "tcp",
		},
		Scheduler: "wlc",
	}}
	ctx.backends["service1-backend1"] = &backend{options: &types.Backend{
		BackendKey: types.BackendKey{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 1234,
		},
		Weight:  1,
		Forward: "nat",
	}, service: ctx.services["service1"], monitor: &pulse.Pulse{}}
	exporter := NewExporter(ctx)
	err := exporter.collect()
	if err != nil {
		t.Fatal(err)
	}
}
