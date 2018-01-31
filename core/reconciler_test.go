package core

import (
	"testing"
	"time"

	"github.com/kobolog/gorb/ipvs-shim"
)

func TestReconcile(t *testing.T) {
	type fields struct {
		period time.Duration
		syncCh chan struct{}
		store  *Store
		ipvs   ipvs_shim.IPVS
	}
	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reconciler{
				period: tt.fields.period,
				syncCh: tt.fields.syncCh,
				store:  tt.fields.store,
				ipvs:   tt.fields.ipvs,
			}
			r.reconcile()
		})
	}
}
