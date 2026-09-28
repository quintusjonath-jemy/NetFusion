package network

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"netfusion/backend/internal/models"
)

// InterfaceIO holds byte accounting for transfer rate calculation
type InterfaceIO struct {
	RxBytes   int64
	TxBytes   int64
	Timestamp time.Time
}

// Discovery manages Linux interface discovery and rate calculation
type Discovery struct {
	mu           sync.Mutex
	previousIO   map[string]InterfaceIO
	uptimeStart  map[string]time.Time
	defaultRoute string
}

// NewDiscovery creates a new Discovery instance
func NewDiscovery() *Discovery {
	return &Discovery{
		previousIO:  make(map[string]InterfaceIO),
		uptimeStart: make(map[string]time.Time),
	}
}

// ScanInterfaces queries Linux sysfs, netlink, and procfs to discover all network interfaces
func (d *Discovery) ScanInterfaces() ([]models.NetworkInterface, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	sysNetDir := "/sys/class/net"
	entries, err := os.ReadDir(sysNetDir)
	if err != nil {
		// Fallback to standard Go net.Interfaces() if sysfs is not accessible
		return d.scanGoNetInterfaces()
	}

	// 1. Read /proc/net/dev for rx/tx byte counters
	ioMap := readProcNetDev()

	// 2. Read /proc/net/route for gateways
	routes, defaultIface := readProcNetRoute()
	d.defaultRoute = defaultIface

	// 3. Read /proc/net/wireless for Wi-Fi signal strength
	wifiSignals := readProcNetWireless()

	var interfaces []models.NetworkInterface
	now := time.Now()

	for _, entry := range entries {
		name := entry.Name()

		// Ignore loopback
		if name == "lo" {
			continue
		}

		ifacePath := filepath.Join(sysNetDir, name)

		// Determine connection type
		ifaceType := classifyInterface(name, ifacePath)

		// Skip virtual docker/veth interfaces unless they are active physical or wireless
		if ifaceType == models.TypeOther && (strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "br-")) {
			continue
		}

		// Read MAC address
		macAddr := ""
		if macBytes, err := os.ReadFile(filepath.Join(ifacePath, "address")); err == nil {
			macAddr = strings.TrimSpace(string(macBytes))
		}

		// Read Carrier & Operstate
		carrier := false
		if cBytes, err := os.ReadFile(filepath.Join(ifacePath, "carrier")); err == nil {
			carrier = strings.TrimSpace(string(cBytes)) == "1"
		}

		operState := "unknown"
		if opBytes, err := os.ReadFile(filepath.Join(ifacePath, "operstate")); err == nil {
			operState = strings.TrimSpace(string(opBytes))
		}

		// Read MTU
		mtu := 1500
		if mtuBytes, err := os.ReadFile(filepath.Join(ifacePath, "mtu")); err == nil {
			if parsedMTU, err := strconv.Atoi(strings.TrimSpace(string(mtuBytes))); err == nil {
				mtu = parsedMTU
			}
		}

		// Retrieve IP address and subnet
		ipAddr, subnet := getInterfaceIPs(name)

		// Gateway
		gateway := routes[name]

		// Signal strength for wireless/cellular
		signalStrength := 100
		if ifaceType == models.TypeWiFi {
			if sig, ok := wifiSignals[name]; ok {
				signalStrength = sig
			} else {
				signalStrength = 75
			}
		} else if ifaceType == models.TypeCellular {
			signalStrength = 80
		}

		// Byte accounting
		io := ioMap[name]
		rxBytes := io.RxBytes
		txBytes := io.TxBytes

		// Calculate transfer rates
		var dlMbps, ulMbps float64
		if prev, ok := d.previousIO[name]; ok {
			elapsed := now.Sub(prev.Timestamp).Seconds()
			if elapsed > 0 {
				deltaRx := rxBytes - prev.RxBytes
				deltaTx := txBytes - prev.TxBytes
				if deltaRx >= 0 {
					dlMbps = (float64(deltaRx) * 8.0) / (elapsed * 1_000_000.0)
				}
				if deltaTx >= 0 {
					ulMbps = (float64(deltaTx) * 8.0) / (elapsed * 1_000_000.0)
				}
			}
		}
		d.previousIO[name] = InterfaceIO{
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
			Timestamp: now,
		}

		// Uptime calculation
		if carrier && ipAddr != "" {
			if _, ok := d.uptimeStart[name]; !ok {
				d.uptimeStart[name] = now
			}
		} else {
			delete(d.uptimeStart, name)
		}

		var uptimeSeconds int64
		if start, ok := d.uptimeStart[name]; ok {
			uptimeSeconds = int64(now.Sub(start).Seconds())
		}

		// Initial interface status
		status := models.StatusConnected
		if !carrier || operState == "down" || ipAddr == "" {
			status = models.StatusDisconnected
		}

		isDefault := (name == d.defaultRoute)

		iface := models.NetworkInterface{
			ID:                  fmt.Sprintf("iface-%s", name),
			Name:                name,
			Type:                ifaceType,
			IPAddress:           ipAddr,
			MACAddress:          macAddr,
			Gateway:             gateway,
			Subnet:              subnet,
			Status:              status,
			Carrier:             carrier,
			MTU:                 mtu,
			SignalStrength:      signalStrength,
			RxBytes:             rxBytes,
			TxBytes:             txBytes,
			TotalDataUsedBytes:  rxBytes + txBytes,
			CurrentDownloadMbps: round(dlMbps, 2),
			CurrentUploadMbps:   round(ulMbps, 2),
			IsDefault:           isDefault,
			IsSimulated:         false,
			UptimeSeconds:       uptimeSeconds,
			UpdatedAt:           now,
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// scanGoNetInterfaces fallback using standard library
func (d *Discovery) scanGoNetInterfaces() ([]models.NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var results []models.NetworkInterface
	now := time.Now()

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		ipAddr := ""
		subnet := ""
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddr = ipNet.IP.String()
					subnet = net.IP(ipNet.Mask).String()
					break
				}
			}
		}

		status := models.StatusConnected
		if iface.Flags&net.FlagUp == 0 || ipAddr == "" {
			status = models.StatusDisconnected
		}

		results = append(results, models.NetworkInterface{
			ID:             fmt.Sprintf("iface-%s", iface.Name),
			Name:           iface.Name,
			Type:           classifyInterface(iface.Name, ""),
			IPAddress:      ipAddr,
			MACAddress:     iface.HardwareAddr.String(),
			Subnet:         subnet,
			Status:         status,
			Carrier:        (iface.Flags & net.FlagUp) != 0,
			MTU:            iface.MTU,
			SignalStrength: 100,
			IsDefault:      false,
			IsSimulated:    false,
			UpdatedAt:      now,
		})
	}
	return results, nil
}

// classifyInterface categorizes interface names
func classifyInterface(name, sysPath string) models.InterfaceType {
	lower := strings.ToLower(name)

	// Check sysfs wireless directory first
	if sysPath != "" {
		if _, err := os.Stat(filepath.Join(sysPath, "wireless")); err == nil {
			return models.TypeWiFi
		}
		if _, err := os.Stat(filepath.Join(sysPath, "phy80211")); err == nil {
			return models.TypeWiFi
		}
	}

	if strings.HasPrefix(lower, "wl") || strings.Contains(lower, "wifi") {
		return models.TypeWiFi
	}
	if strings.HasPrefix(lower, "eth") || strings.HasPrefix(lower, "en") {
		return models.TypeEthernet
	}
	if strings.HasPrefix(lower, "ww") || strings.HasPrefix(lower, "usb") || strings.HasPrefix(lower, "lte") {
		return models.TypeCellular
	}
	return models.TypeOther
}

// readProcNetDev reads /proc/net/dev counters
func readProcNetDev() map[string]InterfaceIO {
	result := make(map[string]InterfaceIO)
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		ifaceName := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) >= 9 {
			rx, _ := strconv.ParseInt(fields[0], 10, 64)
			tx, _ := strconv.ParseInt(fields[8], 10, 64)
			result[ifaceName] = InterfaceIO{
				RxBytes:   rx,
				TxBytes:   tx,
				Timestamp: time.Now(),
			}
		}
	}
	return result
}

// readProcNetRoute extracts gateways and the default route interface
func readProcNetRoute() (map[string]string, string) {
	routes := make(map[string]string)
	defaultIface := ""

	file, err := os.Open("/proc/net/route")
	if err != nil {
		return routes, defaultIface
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header
	if scanner.Scan() {}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		iface := fields[0]
		dest := fields[1]
		gwHex := fields[2]

		if dest == "00000000" && gwHex != "00000000" {
			gwIP := parseHexIP(gwHex)
			routes[iface] = gwIP
			if defaultIface == "" {
				defaultIface = iface
			}
		}
	}
	return routes, defaultIface
}

// parseHexIP converts hex network byte order IP (e.g. 0101A8C0 -> 192.168.1.1)
func parseHexIP(hexStr string) string {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil || len(bytes) != 4 {
		return ""
	}
	// Reverse order for little-endian network byte order
	return fmt.Sprintf("%d.%d.%d.%d", bytes[3], bytes[2], bytes[1], bytes[0])
}

// readProcNetWireless reads Wi-Fi link quality from /proc/net/wireless
func readProcNetWireless() map[string]int {
	signals := make(map[string]int)
	file, err := os.Open("/proc/net/wireless")
	if err != nil {
		return signals
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) >= 2 {
			// fields[1] is link quality, e.g. "61."
			qualityStr := strings.TrimSuffix(fields[1], ".")
			if q, err := strconv.Atoi(qualityStr); err == nil {
				// Normalize 0-70 scale commonly used by Linux wireless drivers to 0-100%
				pct := int((float64(q) / 70.0) * 100.0)
				if pct > 100 {
					pct = 100
				}
				signals[iface] = pct
			}
		}
	}
	return signals
}

// getInterfaceIPs extracts IPv4 and subnet mask
func getInterfaceIPs(ifaceName string) (string, string) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), net.IP(ipNet.Mask).String()
			}
		}
	}
	return "", ""
}

func round(val float64, precision int) float64 {
	format := fmt.Sprintf("%%.%df", precision)
	parsed, _ := strconv.ParseFloat(fmt.Sprintf(format, val), 64)
	return parsed
}
