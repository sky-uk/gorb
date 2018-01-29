package ipvs_shim

import (
	"testing"

	"github.com/mqliang/libipvs"
)

func TestCreateFlagbits(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  uint32
	}{
		{
			"multiple flags",
			[]string{"flag-1", "flag-2"},
			libipvs.IP_VS_SVC_F_SCHED1 | libipvs.IP_VS_SVC_F_SCHED2,
		},
		{
			"invalid flags",
			[]string{"flag-1", "invalid-ignored", "flag-2"},
			libipvs.IP_VS_SVC_F_SCHED1 | libipvs.IP_VS_SVC_F_SCHED2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createFlagbits(tt.flags); got != tt.want {
				t.Errorf("createFlagbits() = %v, want %v", got, tt.want)
			}
		})
	}
}
