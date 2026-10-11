#!/usr/bin/env bash
# ==============================================================================
# NetFusion — Linux Kernel MPTCP Configuration Script
# Enables in-kernel MPTCP path manager, sets subflow limits, and registers endpoints.
# Requires root privileges (CAP_NET_ADMIN).
# ==============================================================================

set -euo pipefail

echo "=========================================================="
echo " NetFusion: Configuring Linux Native MPTCP Subflows"
echo "=========================================================="

if [[ $EUID -ne 0 ]]; then
   echo "[-] ERROR: This script must be run as root (or with sudo)."
   exit 1
fi

# 1. Verify Kernel MPTCP support
if [[ ! -f /proc/sys/net/mptcp/enabled ]]; then
    echo "[-] ERROR: Linux kernel MPTCP is not supported on this kernel."
    echo "    Linux kernel >= 5.6 with CONFIG_MPTCP=y is required."
    exit 1
fi

echo "[+] 1. Enabling Linux Kernel MPTCP..."
sysctl -w net.mptcp.enabled=1
sysctl -w net.mptcp.pm_type=0 # 0 = in-kernel path manager

echo "[+] 2. Configuring MPTCP subflow limits..."
# Set subflow and address announcement limits
ip mptcp limits set subflow 4 add_addr_accepted 4

echo "[+] 3. Discovering available network interfaces..."
INTERFACES=$(ip -o link show | awk -F': ' '{print $2}' | grep -Ev '^(lo|docker|veth|br-)')

echo "    Active interfaces found:"
for IFACE in $INTERFACES; do
    IP=$(ip -4 addr show dev "$IFACE" | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n 1 || true)
    CARRIER=$(cat "/sys/class/net/$IFACE/carrier" 2>/dev/null || echo "0")
    echo "    - $IFACE (IP: ${IP:-None}, Carrier: $CARRIER)"
done

echo "[+] 4. Registering MPTCP subflow endpoints..."
# Clear existing endpoints
ip mptcp endpoint flush || true

EP_ID=1
for IFACE in $INTERFACES; do
    IP=$(ip -4 addr show dev "$IFACE" | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n 1 || true)
    CARRIER=$(cat "/sys/class/net/$IFACE/carrier" 2>/dev/null || echo "0")
    
    if [[ -n "$IP" && "$CARRIER" == "1" ]]; then
        echo "    Adding endpoint: $IFACE -> $IP (subflow)"
        ip mptcp endpoint add "$IP" dev "$IFACE" subflow id "$EP_ID" || true
        EP_ID=$((EP_ID + 1))
    fi
done

echo ""
echo "[+] 5. Current Active MPTCP Endpoints:"
ip mptcp endpoint show
echo ""
echo "[+] MPTCP configuration completed successfully!"
