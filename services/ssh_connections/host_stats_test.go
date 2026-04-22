package ssh_connections

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseKeyedUintOutput(t *testing.T) {
	values := parseKeyedUintOutput(`
cpu_cores=8
memory_total_bytes=33554432
memory_available_bytes=16777216
storage_total_bytes=1000000000
storage_available_bytes=250000000
ignored
bad=not-a-number
`)

	assertUintValue(t, values, "cpu_cores", 8)
	assertUintValue(t, values, "memory_total_bytes", 33554432)
	assertUintValue(t, values, "memory_available_bytes", 16777216)
	assertUintValue(t, values, "storage_total_bytes", 1000000000)
	assertUintValue(t, values, "storage_available_bytes", 250000000)

	if _, ok := values["bad"]; ok {
		t.Fatal("expected invalid numeric values to be skipped")
	}
}

func TestApplyHostStatsOutput(t *testing.T) {
	stats := HostResourceStats{HostServerID: uuid.New(), Status: "ok"}

	err := applyHostStatsOutput(&stats, `
cpu_cores=4
memory_total_bytes=16000
memory_available_bytes=8000
storage_total_bytes=64000
storage_available_bytes=32000
`)
	if err != nil {
		t.Fatalf("expected stats output to parse: %v", err)
	}

	if stats.CPUCores == nil || *stats.CPUCores != 4 {
		t.Fatalf("unexpected cpu cores: %v", stats.CPUCores)
	}
	if stats.MemoryTotalBytes == nil || *stats.MemoryTotalBytes != 16000 {
		t.Fatalf("unexpected memory total: %v", stats.MemoryTotalBytes)
	}
	if stats.StorageAvailableBytes == nil || *stats.StorageAvailableBytes != 32000 {
		t.Fatalf("unexpected storage available: %v", stats.StorageAvailableBytes)
	}
}

func TestApplyHostStatsOutputRequiresSupportedStats(t *testing.T) {
	stats := HostResourceStats{HostServerID: uuid.New(), Status: "ok"}

	if err := applyHostStatsOutput(&stats, "unsupported=1\n"); err == nil {
		t.Fatal("expected unsupported output to fail")
	}
}

func TestClassifyCapacityRole(t *testing.T) {
	tests := []struct {
		name            string
		hostServerTypes []string
		platformTypes   []string
		want            string
	}{
		{
			name:          "proxmox node is hypervisor capacity",
			platformTypes: []string{"Proxmox VE"},
			want:          "hypervisor",
		},
		{
			name:            "container host is physical capacity",
			hostServerTypes: []string{"Container Host"},
			want:            "physical",
		},
		{
			name:            "virtual machine is guest",
			hostServerTypes: []string{"Virtual Machine"},
			want:            "guest",
		},
		{
			name:          "lxc is guest",
			platformTypes: []string{"LXC Container"},
			want:          "guest",
		},
		{
			name: "missing types are unclassified",
			want: "unclassified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyCapacityRole(tt.hostServerTypes, tt.platformTypes)
			if got != tt.want {
				t.Fatalf("unexpected capacity role: got %q want %q", got, tt.want)
			}
		})
	}
}

func assertUintValue(t *testing.T, values map[string]uint64, key string, want uint64) {
	t.Helper()

	got, ok := values[key]
	if !ok {
		t.Fatalf("expected %q to be parsed", key)
	}
	if got != want {
		t.Fatalf("unexpected value for %q: got %d want %d", key, got, want)
	}
}
