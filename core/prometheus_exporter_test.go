package core

import (
	"testing"

	"github.com/kobolog/gorb/pulse"
	"github.com/kobolog/gorb/types"
)

func TestCollector(t *testing.T) {
	ctx := &Context{
		services: make(map[string]*service),
		backends: make(map[string]*backend),
	}
	ctx.services["service1"] = &service{options: &types.Service{
		Host:     "localhost",
		Port:     1234,
		Protocol: "tcp",
		Method:   "wlc",
	}}
	ctx.backends["service1-backend1"] = &backend{options: &types.BackendOptions{
		Host:   "localhost",
		Port:   1234,
		Weight: 1,
		Method: "nat",
		VsID:   "service1",
	}, service: ctx.services["service1"], monitor: &pulse.Pulse{}}
	exporter := NewExporter(ctx)
	err := exporter.collect()
	if err != nil {
		t.Fatal(err)
	}
}
