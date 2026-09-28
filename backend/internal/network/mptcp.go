package network

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MPTCPEndpoint represents an active Linux MPTCP path endpoint
type MPTCPEndpoint struct {
	ID        int      `json:"id"`
	IP        string   `json:"ip"`
	Interface string   `json:"interface"`
	Flags     []string `json:"flags"` // e.g. "subflow", "backup", "signal"
	Port      int      `json:"port,omitempty"`
}

// MPTCPStatus represents kernel-level MPTCP availability and settings
type MPTCPStatus struct {
	Supported    bool            `json:"supported"`
	Enabled      bool            `json:"enabled"`
	PathManager  string          `json:"pathManager"`
	Scheduler    string          `json:"scheduler"`
	Endpoints    []MPTCPEndpoint `json:"endpoints"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// CheckMPTCPStatus inspects Linux kernel MPTCP support and active endpoints
func CheckMPTCPStatus() MPTCPStatus {
	status := MPTCPStatus{
		Supported: false,
		Enabled:   false,
		Endpoints: make([]MPTCPEndpoint, 0),
	}

	// 1. Check /proc/sys/net/mptcp/enabled
	enabledData, err := os.ReadFile("/proc/sys/net/mptcp/enabled")
	if err != nil {
		status.ErrorMessage = "MPTCP not supported or kernel module not loaded (/proc/sys/net/mptcp/enabled missing)"
		return status
	}

	status.Supported = true
	enabledVal := strings.TrimSpace(string(enabledData))
	if enabledVal == "1" || enabledVal == "true" {
		status.Enabled = true
	} else {
		status.ErrorMessage = "MPTCP is supported by kernel but disabled in sysctl (net.mptcp.enabled = 0)"
	}

	// 2. Read path manager if available
	if pmData, err := os.ReadFile("/proc/sys/net/mptcp/pm_type"); err == nil {
		pmType := strings.TrimSpace(string(pmData))
		if pmType == "0" {
			status.PathManager = "in-kernel"
		} else if pmType == "1" {
			status.PathManager = "userspace"
		} else {
			status.PathManager = pmType
		}
	} else {
		status.PathManager = "kernel-default"
	}

	// 3. Query endpoints via `ip mptcp endpoint show`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ip", "mptcp", "endpoint", "show")
	output, err := cmd.Output()
	if err == nil {
		status.Endpoints = parseEndpoints(output)
	}

	return status
}

// parseEndpoints parses output of `ip mptcp endpoint show`
// Example line: "192.168.1.38 id 2 subflow dev wlan0"
func parseEndpoints(data []byte) []MPTCPEndpoint {
	var endpoints []MPTCPEndpoint
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		ep := MPTCPEndpoint{
			IP:    parts[0],
			Flags: make([]string, 0),
		}

		for i := 1; i < len(parts); i++ {
			switch parts[i] {
			case "id":
				if i+1 < len(parts) {
					if id, err := strconv.Atoi(parts[i+1]); err == nil {
						ep.ID = id
					}
					i++
				}
			case "dev":
				if i+1 < len(parts) {
					ep.Interface = parts[i+1]
					i++
				}
			case "port":
				if i+1 < len(parts) {
					if p, err := strconv.Atoi(parts[i+1]); err == nil {
						ep.Port = p
					}
					i++
				}
			default:
				// Recognized flags: subflow, backup, signal, fullmesh, laminar
				flag := parts[i]
				if flag != "" && !strings.Contains(flag, "=") {
					ep.Flags = append(ep.Flags, flag)
				}
			}
		}

		endpoints = append(endpoints, ep)
	}

	return endpoints
}
