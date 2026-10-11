import React from 'react'
import {
  HardDrive,
  Wifi,
  Signal,
  Network,
  Activity,
  Layers,
  ShieldCheck,
  Zap,
} from 'lucide-react'
import { useNetFusion } from '../context/NetFusionContext'

export const NetworksPage: React.FC = () => {
  const { interfaces } = useNetFusion()

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
  }

  const getInterfaceIcon = (type: string) => {
    switch (type) {
      case 'ethernet':
        return <HardDrive size={22} color="var(--neon-cyan)" />
      case 'wifi':
        return <Wifi size={22} color="var(--neon-purple)" />
      case 'cellular':
        return <Signal size={22} color="var(--neon-amber)" />
      default:
        return <Network size={22} color="var(--neon-blue)" />
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'connected':
        return <span className="badge badge-connected">Connected</span>
      case 'degraded':
        return <span className="badge badge-degraded">Degraded</span>
      case 'disconnected':
        return <span className="badge badge-disconnected">Disconnected</span>
      case 'recovering':
        return <span className="badge badge-recovering">Recovering</span>
      default:
        return <span className="badge">{status}</span>
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Network Interface Telemetry & Link Health
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Real-time interface link metrics, Linux socket binding capabilities, signal quality, and stability scoring.
        </p>
      </div>

      {/* Grid of Interface Detail Cards */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
        {interfaces.map((inf) => (
          <div
            key={inf.id}
            className="glass-panel"
            style={{
              padding: '24px',
              borderLeft: `4px solid ${
                inf.status === 'connected'
                  ? 'var(--neon-emerald)'
                  : inf.status === 'degraded'
                  ? 'var(--neon-amber)'
                  : inf.status === 'disconnected'
                  ? 'var(--neon-crimson)'
                  : 'var(--neon-cyan)'
              }`,
            }}
          >
            {/* Top Card Header */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderBottom: '1px solid var(--border-subtle)',
                paddingBottom: '16px',
                marginBottom: '20px',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                <div
                  style={{
                    width: '46px',
                    height: '46px',
                    borderRadius: '12px',
                    background: 'rgba(255, 255, 255, 0.05)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    border: '1px solid var(--border-subtle)',
                  }}
                >
                  {getInterfaceIcon(inf.type)}
                </div>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <h2 style={{ fontSize: '1.25rem', color: '#ffffff' }}>{inf.name}</h2>
                    {getStatusBadge(inf.status)}
                    {inf.isDefault && (
                      <span
                        className="badge"
                        style={{
                          background: 'rgba(0, 240, 255, 0.1)',
                          color: 'var(--neon-cyan)',
                          border: '1px solid rgba(0, 240, 255, 0.3)',
                        }}
                      >
                        Default Gateway
                      </span>
                    )}
                    {inf.isSimulated && (
                      <span
                        className="badge"
                        style={{
                          background: 'rgba(157, 78, 221, 0.1)',
                          color: '#c084fc',
                          border: '1px solid rgba(157, 78, 221, 0.3)',
                        }}
                      >
                        Simulated
                      </span>
                    )}
                  </div>
                  <span className="mono" style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
                    Type: {inf.type.toUpperCase()} • MAC: {inf.macAddress || 'N/A'} • Carrier:{' '}
                    {inf.carrier ? 'DETECTED' : 'NO CARRIER'}
                  </span>
                </div>
              </div>

              {/* Traffic Weight Pill */}
              <div style={{ textAlign: 'right' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block' }}>
                  Aggregated Allocation
                </span>
                <span
                  className="mono"
                  style={{
                    fontSize: '1.5rem',
                    fontWeight: 800,
                    color: inf.allocatedWeightPct > 0 ? 'var(--neon-cyan)' : 'var(--text-muted)',
                  }}
                >
                  {inf.allocatedWeightPct.toFixed(1)}%
                </span>
              </div>
            </div>

            {/* Metrics Breakdown Grid */}
            <div className="grid-4" style={{ marginBottom: '20px' }}>
              {/* Latency */}
              <div
                style={{
                  background: 'rgba(11, 16, 30, 0.6)',
                  padding: '14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(255, 255, 255, 0.04)',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>TCP RTT Latency</span>
                  <Activity size={14} color="var(--neon-cyan)" />
                </div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px' }}>
                  <span className="mono" style={{ fontSize: '1.25rem', fontWeight: 700, color: '#ffffff' }}>
                    {inf.latencyMs.toFixed(1)}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--neon-cyan)' }}>ms</span>
                </div>
                <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                  Jitter: {inf.jitterMs.toFixed(1)} ms
                </span>
              </div>

              {/* Packet Loss */}
              <div
                style={{
                  background: 'rgba(11, 16, 30, 0.6)',
                  padding: '14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(255, 255, 255, 0.04)',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Packet Loss</span>
                  <Layers size={14} color="var(--neon-amber)" />
                </div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px' }}>
                  <span
                    className="mono"
                    style={{
                      fontSize: '1.25rem',
                      fontWeight: 700,
                      color: inf.packetLoss > 2 ? 'var(--neon-crimson)' : 'var(--neon-emerald)',
                    }}
                  >
                    {inf.packetLoss.toFixed(1)}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>%</span>
                </div>
                <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                  Threshold: &lt; 5.0%
                </span>
              </div>

              {/* Stability Score */}
              <div
                style={{
                  background: 'rgba(11, 16, 30, 0.6)',
                  padding: '14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(255, 255, 255, 0.04)',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Stability Index</span>
                  <ShieldCheck size={14} color="var(--neon-emerald)" />
                </div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px' }}>
                  <span className="mono" style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--neon-emerald)' }}>
                    {inf.stabilityScore.toFixed(0)}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>/ 100</span>
                </div>
                <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                  Dynamic Rank: {inf.dynamicScore.toFixed(1)}
                </span>
              </div>

              {/* Current Bandwidth */}
              <div
                style={{
                  background: 'rgba(11, 16, 30, 0.6)',
                  padding: '14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(255, 255, 255, 0.04)',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Realtime Bandwidth</span>
                  <Zap size={14} color="var(--neon-purple)" />
                </div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px' }}>
                  <span className="mono" style={{ fontSize: '1.25rem', fontWeight: 700, color: '#ffffff' }}>
                    {inf.currentDownloadMbps.toFixed(1)}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--neon-purple)' }}>Mbps DL</span>
                </div>
                <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                  UL: {inf.currentUploadMbps.toFixed(1)} Mbps
                </span>
              </div>
            </div>

            {/* Hardware & Network Addressing Specs */}
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(4, 1fr)',
                gap: '12px',
                padding: '14px 18px',
                background: 'rgba(7, 10, 20, 0.5)',
                borderRadius: '8px',
                fontSize: '0.75rem',
              }}
            >
              <div>
                <span style={{ color: 'var(--text-muted)', display: 'block' }}>IPv4 Address</span>
                <span className="mono" style={{ color: '#ffffff', fontWeight: 600 }}>
                  {inf.ipAddress || 'Unassigned'}
                </span>
              </div>
              <div>
                <span style={{ color: 'var(--text-muted)', display: 'block' }}>Gateway & Subnet</span>
                <span className="mono" style={{ color: 'var(--text-secondary)' }}>
                  {inf.gateway || '0.0.0.0'} / {inf.subnet}
                </span>
              </div>
              <div>
                <span style={{ color: 'var(--text-muted)', display: 'block' }}>MTU Frame Size</span>
                <span className="mono" style={{ color: 'var(--text-secondary)' }}>
                  {inf.mtu} Bytes
                </span>
              </div>
              <div>
                <span style={{ color: 'var(--text-muted)', display: 'block' }}>Total Data Exchanged</span>
                <span className="mono" style={{ color: 'var(--neon-cyan)' }}>
                  {formatBytes(inf.totalDataUsedBytes)}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
