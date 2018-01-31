package ipvs_shim

import (
	"reflect"
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

func TestConvertFlagbits(t *testing.T) {
	tests := []struct {
		name     string
		flagbits uint32
		want     []string
	}{
		{
			"flag-1",
			libipvs.IP_VS_SVC_F_SCHED1,
			[]string{"flag-1"},
		},
		{
			"flag-2",
			libipvs.IP_VS_SVC_F_SCHED2,
			[]string{"flag-2"},
		},
		{
			"flag-3",
			libipvs.IP_VS_SVC_F_SCHED3,
			[]string{"flag-3"},
		},
		{
			"multiple flags",
			libipvs.IP_VS_SVC_F_SCHED1 | libipvs.IP_VS_SVC_F_SCHED2 | libipvs.IP_VS_SVC_F_SCHED3,
			[]string{"flag-1", "flag-2", "flag-3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertFlagbits(tt.flagbits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertFlagbits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertShFlagbits(t *testing.T) {
	tests := []struct {
		name     string
		flagbits uint32
		want     []string
	}{
		{
			"sh-fallback",
			libipvs.IP_VS_SVC_F_SCHED_SH_FALLBACK,
			[]string{"sh-fallback"},
		},
		{
			"sh-port",
			libipvs.IP_VS_SVC_F_SCHED_SH_PORT,
			[]string{"sh-port"},
		},
		{
			"multiple flags",
			libipvs.IP_VS_SVC_F_SCHED_SH_FALLBACK | libipvs.IP_VS_SVC_F_SCHED_SH_PORT,
			[]string{"sh-fallback", "sh-port"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertShFlagbits(tt.flagbits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertFlagbits() = %v, want %v", got, tt.want)
			}
		})
	}
}
