package ssh_connections

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
)

const hostStatsCommand = `sh -c 'printf "cpu_cores=%s\n" "$(getconf _NPROCESSORS_ONLN 2>/dev/null || nproc 2>/dev/null || echo 0)"; awk "/^MemTotal:/ {printf \"memory_total_bytes=%.0f\\n\", \$2 * 1024} /^MemAvailable:/ {printf \"memory_available_bytes=%.0f\\n\", \$2 * 1024}" /proc/meminfo 2>/dev/null; df -B1 --total -x tmpfs -x devtmpfs -x squashfs 2>/dev/null | awk "/^total/ {print \"storage_total_bytes=\" \$2; print \"storage_available_bytes=\" \$4}"'`

type HostResourceStats struct {
	HostServerID          uuid.UUID `json:"hostServerId"`
	Hostname              string    `json:"hostname"`
	IPAddress             string    `json:"ipAddress"`
	CapacityRole          string    `json:"capacityRole"`
	HostServerTypes       []string  `json:"hostServerTypes,omitempty"`
	PlatformTypes         []string  `json:"platformTypes,omitempty"`
	CollectedAt           time.Time `json:"collectedAt"`
	Status                string    `json:"status"`
	Error                 string    `json:"error,omitempty"`
	CPUCores              *uint64   `json:"cpuCores,omitempty"`
	MemoryTotalBytes      *uint64   `json:"memoryTotalBytes,omitempty"`
	MemoryAvailableBytes  *uint64   `json:"memoryAvailableBytes,omitempty"`
	StorageTotalBytes     *uint64   `json:"storageTotalBytes,omitempty"`
	StorageAvailableBytes *uint64   `json:"storageAvailableBytes,omitempty"`
}

type HostResourceStatsSummary struct {
	CollectedAt                       time.Time           `json:"collectedAt"`
	HostCount                         int                 `json:"hostCount"`
	ReachableHostCount                int                 `json:"reachableHostCount"`
	CapacityHostCount                 int                 `json:"capacityHostCount"`
	GuestHostCount                    int                 `json:"guestHostCount"`
	UnclassifiedHostCount             int                 `json:"unclassifiedHostCount"`
	TotalCPUCores                     uint64              `json:"totalCpuCores"`
	MemoryTotalBytes                  uint64              `json:"memoryTotalBytes"`
	MemoryAvailableBytes              uint64              `json:"memoryAvailableBytes"`
	StorageTotalBytes                 uint64              `json:"storageTotalBytes"`
	StorageAvailableBytes             uint64              `json:"storageAvailableBytes"`
	GuestTotalCPUCores                uint64              `json:"guestTotalCpuCores"`
	GuestMemoryTotalBytes             uint64              `json:"guestMemoryTotalBytes"`
	GuestMemoryAvailableBytes         uint64              `json:"guestMemoryAvailableBytes"`
	GuestStorageTotalBytes            uint64              `json:"guestStorageTotalBytes"`
	GuestStorageAvailableBytes        uint64              `json:"guestStorageAvailableBytes"`
	UnclassifiedTotalCPUCores         uint64              `json:"unclassifiedTotalCpuCores"`
	UnclassifiedMemoryTotalBytes      uint64              `json:"unclassifiedMemoryTotalBytes"`
	UnclassifiedStorageTotalBytes     uint64              `json:"unclassifiedStorageTotalBytes"`
	UnclassifiedStorageAvailableBytes uint64              `json:"unclassifiedStorageAvailableBytes"`
	HasCPUCores                       bool                `json:"hasCpuCores"`
	HasMemory                         bool                `json:"hasMemory"`
	HasStorage                        bool                `json:"hasStorage"`
	HasGuestCPUCores                  bool                `json:"hasGuestCpuCores"`
	HasGuestMemory                    bool                `json:"hasGuestMemory"`
	HasGuestStorage                   bool                `json:"hasGuestStorage"`
	HasUnclassifiedStats              bool                `json:"hasUnclassifiedStats"`
	Hosts                             []HostResourceStats `json:"hosts"`
}

func (m *SSHConnectionManager) CollectHostStats(ctx context.Context, userID uuid.UUID, hostServerID uuid.UUID) HostResourceStats {
	hostInfo, err := m.getHostServerInfo(hostServerID)
	if err != nil {
		return failedHostStats(hostServerID, "", "", fmt.Errorf("host server not found: %w", err))
	}

	hostServerTypes, platformTypes := m.getHostClassification(ctx, hostServerID)
	stats := m.collectHostInfoStats(userID, hostInfo)
	stats.HostServerTypes = hostServerTypes
	stats.PlatformTypes = platformTypes
	stats.CapacityRole = classifyCapacityRole(hostServerTypes, platformTypes)
	return stats
}

func (m *SSHConnectionManager) CollectAllHostStats(ctx context.Context, userID uuid.UUID) (HostResourceStatsSummary, error) {
	servers, err := m.db.GetAllHostServers(ctx)
	if err != nil {
		return HostResourceStatsSummary{}, fmt.Errorf("failed to list host servers: %w", err)
	}

	summary := HostResourceStatsSummary{
		CollectedAt: time.Now().UTC(),
		HostCount:   len(servers),
		Hosts:       make([]HostResourceStats, 0, len(servers)),
	}

	for _, server := range servers {
		hostInfo := hostInfoFromDB(server)
		hostServerTypes, platformTypes := m.getHostClassification(ctx, server.ID)
		stats := m.collectHostInfoStats(userID, hostInfo)
		stats.HostServerTypes = hostServerTypes
		stats.PlatformTypes = platformTypes
		stats.CapacityRole = classifyCapacityRole(hostServerTypes, platformTypes)
		summary.Hosts = append(summary.Hosts, stats)

		if stats.Status == "ok" {
			summary.ReachableHostCount++
		}
		summary.addStats(stats)
	}

	return summary, nil
}

func (m *SSHConnectionManager) collectHostInfoStats(userID uuid.UUID, hostInfo *HostServerInfo) HostResourceStats {
	stats := HostResourceStats{
		HostServerID: hostInfo.ID,
		Hostname:     hostInfo.Hostname,
		IPAddress:    hostInfo.IPAddress,
		CapacityRole: "unclassified",
		CollectedAt:  time.Now().UTC(),
		Status:       "ok",
	}

	sshKey, err := m.GetSSHKeyForHost(userID, hostInfo.ID)
	if err != nil {
		return stats.withError(fmt.Errorf("ssh key not found for host: %w", err))
	}

	client, err := newGophClient(hostInfo, sshKey, m.config)
	if err != nil {
		return stats.withError(fmt.Errorf("failed to connect over ssh: %w", err))
	}
	defer client.Close()

	output, err := client.Run(hostStatsCommand)
	if err != nil {
		return stats.withError(fmt.Errorf("failed to collect host stats: %w", err))
	}

	if err := applyHostStatsOutput(&stats, string(output)); err != nil {
		return stats.withError(err)
	}

	return stats
}

func applyHostStatsOutput(stats *HostResourceStats, output string) error {
	values := parseKeyedUintOutput(output)

	if value, ok := values["cpu_cores"]; ok {
		stats.CPUCores = &value
	}
	if value, ok := values["memory_total_bytes"]; ok {
		stats.MemoryTotalBytes = &value
	}
	if value, ok := values["memory_available_bytes"]; ok {
		stats.MemoryAvailableBytes = &value
	}
	if value, ok := values["storage_total_bytes"]; ok {
		stats.StorageTotalBytes = &value
	}
	if value, ok := values["storage_available_bytes"]; ok {
		stats.StorageAvailableBytes = &value
	}

	if stats.CPUCores == nil && stats.MemoryTotalBytes == nil && stats.StorageAvailableBytes == nil {
		return fmt.Errorf("host did not return any supported stats")
	}

	return nil
}

func parseKeyedUintOutput(output string) map[string]uint64 {
	values := make(map[string]uint64)

	for _, line := range strings.Split(output, "\n") {
		key, rawValue, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}

		value, err := strconv.ParseUint(strings.TrimSpace(rawValue), 10, 64)
		if err != nil {
			continue
		}

		values[strings.TrimSpace(key)] = value
	}

	return values
}

func hostInfoFromDB(server infra_db_pg.HostServer) *HostServerInfo {
	return &HostServerInfo{
		ID:        server.ID,
		Hostname:  server.Hostname,
		IPAddress: server.IpAddress.String(),
		Port:      22,
	}
}

func (m *SSHConnectionManager) getHostClassification(ctx context.Context, hostServerID uuid.UUID) ([]string, []string) {
	hostServerTypes := []string{}
	if mappings, err := m.db.GetHostServerTypeMappingsByHostId(ctx, hostServerID); err == nil {
		for _, mapping := range mappings {
			hostServerTypes = append(hostServerTypes, mapping.HostServerTypeName)
		}
	}

	platformTypes := []string{}
	if mappings, err := m.db.GetPlatformTypeMappingsByHostId(ctx, hostServerID); err == nil {
		for _, mapping := range mappings {
			platformTypes = append(platformTypes, mapping.PlatformTypeName)
		}
	}

	return hostServerTypes, platformTypes
}

func classifyCapacityRole(hostServerTypes []string, platformTypes []string) string {
	names := append([]string{}, hostServerTypes...)
	names = append(names, platformTypes...)

	for _, name := range names {
		normalized := normalizeClassificationName(name)
		if strings.Contains(normalized, "proxmox") || strings.Contains(normalized, "hypervisor") {
			return "hypervisor"
		}
		if strings.Contains(normalized, "bare metal") || strings.Contains(normalized, "physical") {
			return "physical"
		}
	}

	for _, name := range names {
		normalized := normalizeClassificationName(name)
		if strings.Contains(normalized, "virtual machine") ||
			strings.Contains(normalized, " qemu ") ||
			strings.Contains(normalized, " vm ") ||
			strings.Contains(normalized, " lxc ") ||
			strings.Contains(normalized, "container guest") ||
			(strings.Contains(normalized, "container") && !strings.Contains(normalized, "host")) {
			return "guest"
		}
	}

	for _, name := range names {
		normalized := normalizeClassificationName(name)
		if strings.Contains(normalized, "host") || strings.Contains(normalized, "node") {
			return "physical"
		}
	}

	return "unclassified"
}

func normalizeClassificationName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	return " " + strings.Join(strings.Fields(normalized), " ") + " "
}

func (s *HostResourceStatsSummary) addStats(stats HostResourceStats) {
	switch stats.CapacityRole {
	case "physical", "hypervisor":
		s.CapacityHostCount++
		if stats.CPUCores != nil {
			s.TotalCPUCores += *stats.CPUCores
			s.HasCPUCores = true
		}
		if stats.MemoryTotalBytes != nil {
			s.MemoryTotalBytes += *stats.MemoryTotalBytes
			s.HasMemory = true
		}
		if stats.MemoryAvailableBytes != nil {
			s.MemoryAvailableBytes += *stats.MemoryAvailableBytes
			s.HasMemory = true
		}
		if stats.StorageTotalBytes != nil {
			s.StorageTotalBytes += *stats.StorageTotalBytes
			s.HasStorage = true
		}
		if stats.StorageAvailableBytes != nil {
			s.StorageAvailableBytes += *stats.StorageAvailableBytes
			s.HasStorage = true
		}
	case "guest":
		s.GuestHostCount++
		if stats.CPUCores != nil {
			s.GuestTotalCPUCores += *stats.CPUCores
			s.HasGuestCPUCores = true
		}
		if stats.MemoryTotalBytes != nil {
			s.GuestMemoryTotalBytes += *stats.MemoryTotalBytes
			s.HasGuestMemory = true
		}
		if stats.MemoryAvailableBytes != nil {
			s.GuestMemoryAvailableBytes += *stats.MemoryAvailableBytes
			s.HasGuestMemory = true
		}
		if stats.StorageTotalBytes != nil {
			s.GuestStorageTotalBytes += *stats.StorageTotalBytes
			s.HasGuestStorage = true
		}
		if stats.StorageAvailableBytes != nil {
			s.GuestStorageAvailableBytes += *stats.StorageAvailableBytes
			s.HasGuestStorage = true
		}
	default:
		s.UnclassifiedHostCount++
		if stats.CPUCores != nil {
			s.UnclassifiedTotalCPUCores += *stats.CPUCores
			s.HasUnclassifiedStats = true
		}
		if stats.MemoryTotalBytes != nil {
			s.UnclassifiedMemoryTotalBytes += *stats.MemoryTotalBytes
			s.HasUnclassifiedStats = true
		}
		if stats.StorageTotalBytes != nil {
			s.UnclassifiedStorageTotalBytes += *stats.StorageTotalBytes
			s.HasUnclassifiedStats = true
		}
		if stats.StorageAvailableBytes != nil {
			s.UnclassifiedStorageAvailableBytes += *stats.StorageAvailableBytes
			s.HasUnclassifiedStats = true
		}
	}
}

func failedHostStats(hostServerID uuid.UUID, hostname string, ipAddress string, err error) HostResourceStats {
	return HostResourceStats{
		HostServerID: hostServerID,
		Hostname:     hostname,
		IPAddress:    ipAddress,
		CapacityRole: "unclassified",
		CollectedAt:  time.Now().UTC(),
		Status:       "error",
		Error:        err.Error(),
	}
}

func (s HostResourceStats) withError(err error) HostResourceStats {
	s.Status = "error"
	s.Error = err.Error()
	return s
}
