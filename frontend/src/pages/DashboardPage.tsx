import React from 'react'
import {
  ArrowDownRight,
  ArrowUpRight,
  Network,
  Shield,
  Radio,
  Wifi,
  HardDrive,
  Signal,
} from 'lucide-react'
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { useNetFusion } from '../context/NetFusionContext'

export const DashboardPage: React.FC = () => {
  const {
    interfaces,
    trafficDistribution,
    totalDownloadMbps,
    totalUploadMbps,
    activePathsCount,
    mptcpStatus,
    activePolicy,
    events,
    simulationMode,
  } = useNetFusion()

  // Donut chart colors
  const COLORS = ['#00f0ff', '#9d4edd', '#10b981', '#f59e0b', '#3b82f6']

  const pieData = trafficDistribution.map((d, index) => ({
    name: d.interfaceName,
    value: d.weightPct,
    color: COLORS[index % COLORS.length],
  }))

  // Synthetic trend points for live HUD graph
  const trendData = [
    { time: '10s ago', dl: Math.max(10, totalDownloadMbps * 0.85), ul: Math.max(5, totalUploadMbps * 0.8) },
    { time: '8s ago', dl: Math.max(12, totalDownloadMbps * 0.92), ul: Math.max(6, totalUploadMbps * 0.9) },
    { time: '6s ago', dl: Math.max(15, totalDownloadMbps * 0.98), ul: Math.max(7, totalUploadMbps * 0.95) },
    { time: '4s ago', dl: Math.max(11, totalDownloadMbps * 0.89), ul: Math.max(5, totalUploadMbps * 0.85) },
    { time: '2s ago', dl: Math.max(14, totalDownloadMbps * 0.96), ul: Math.max(6, totalUploadMbps * 0.92) },
    { time: 'Now', dl: totalDownloadMbps, ul: totalUploadMbps },
  ]

  const getInterfaceIcon = (type: string) => {
    switch (type) {
      case 'ethernet':
        return <HardDrive size={18} color="var(--neon-cyan)" />
      case 'wifi':
        return <Wifi size={18} color="var(--neon-purple)" />
      case 'cellular':
        return <Signal size={18} color="var(--neon-amber)" />
      default:
        return <Network size={18} color="var(--neon-blue)" />
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'connected':
        return <span className="badge badge-connected">Connected</span>
      case 'degraded':
        return <span className="badge badge-degraded">Degraded</span>
      case 'disconnected':
        return <span className="badge badge-disconnected">Offline</span>
      case 'recovering':
        return <span className="badge badge-recovering">Recovering</span>
      default:
        return <span className="badge">{status}</span>
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Top Banner Title */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
            System Telemetry & Multi-Network Dashboard
          </h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
            Real-time multi-interface bandwidth aggregation, in-kernel MPTCP subflows, and dynamic failover controller.
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          {simulationMode && (
            <span
              className="badge"
              style={{
                background: 'rgba(157, 78, 221, 0.15)',
                color: '#c084fc',
                border: '1px solid rgba(157, 78, 221, 0.3)',
              }}
            >
              <Radio size={12} /> Interactive Simulation Active
            </span>
          )}
        </div>
      </div>

      {/* 4 HUD Metric Cards */}
      <div className="grid-4">
        {/* Metric 1: Download */}
        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', fontWeight: 600 }}>
              AGGREGATE DOWNLOAD
            </span>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(0, 240, 255, 0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <ArrowDownRight size={18} color="var(--neon-cyan)" />
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '8px' }}>
            <span className="mono" style={{ fontSize: '1.875rem', fontWeight: 800, color: '#ffffff' }}>
              {totalDownloadMbps.toFixed(1)}
            </span>
            <span style={{ fontSize: '0.875rem', color: 'var(--neon-cyan)', fontWeight: 600 }}>Mbps</span>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '6px' }}>
            Combined across {activePathsCount} active subflows
          </p>
        </div>

        {/* Metric 2: Upload */}
        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', fontWeight: 600 }}>
              AGGREGATE UPLOAD
            </span>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(157, 78, 221, 0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <ArrowUpRight size={18} color="var(--neon-purple)" />
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '8px' }}>
            <span className="mono" style={{ fontSize: '1.875rem', fontWeight: 800, color: '#ffffff' }}>
              {totalUploadMbps.toFixed(1)}
            </span>
            <span style={{ fontSize: '0.875rem', color: 'var(--neon-purple)', fontWeight: 600 }}>Mbps</span>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '6px' }}>
            Multi-path upstream aggregation
          </p>
        </div>

        {/* Metric 3: Active Interfaces */}
        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', fontWeight: 600 }}>
              ACTIVE INTERFACES
            </span>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(16, 185, 129, 0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Network size={18} color="var(--neon-emerald)" />
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '8px' }}>
            <span className="mono" style={{ fontSize: '1.875rem', fontWeight: 800, color: '#ffffff' }}>
              {interfaces.filter((i) => i.status === 'connected' || i.status === 'degraded').length}
            </span>
            <span style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>/ {interfaces.length} Discovered</span>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--neon-emerald)', marginTop: '6px' }}>
            {mptcpStatus?.enabled ? 'Linux MPTCP Path Manager Enabled' : 'Socket LocalAddr Fallback'}
          </p>
        </div>

        {/* Metric 4: Active Policy */}
        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', fontWeight: 600 }}>
              ALLOCATION POLICY
            </span>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(245, 158, 11, 0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Shield size={18} color="var(--neon-amber)" />
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '8px' }}>
            <span style={{ fontSize: '1.25rem', fontWeight: 700, color: '#ffffff', textTransform: 'capitalize' }}>
              {activePolicy?.mode.replace('_', ' ') || 'Balanced'}
            </span>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '6px' }}>
            Auto-Failover: {activePolicy?.failoverEnabled ? 'Enabled (Instant)' : 'Disabled'}
          </p>
        </div>
      </div>

      {/* Middle Row: Live Bandwidth Chart & Traffic Distribution Donut */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '20px' }}>
        {/* Chart: Aggregate Throughput */}
        <div className="glass-panel" style={{ padding: '22px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '18px' }}>
            <div>
              <h3 style={{ fontSize: '1.0625rem', color: '#ffffff' }}>Aggregate Throughput Timeline</h3>
              <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Combined real-time multi-network bandwidth rate
              </p>
            </div>
            <div style={{ display: 'flex', gap: '16px', fontSize: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <span style={{ width: '10px', height: '10px', borderRadius: '2px', background: 'var(--neon-cyan)' }} />
                <span style={{ color: 'var(--text-secondary)' }}>Download</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <span style={{ width: '10px', height: '10px', borderRadius: '2px', background: 'var(--neon-purple)' }} />
                <span style={{ color: 'var(--text-secondary)' }}>Upload</span>
              </div>
            </div>
          </div>

          <div style={{ height: '220px', width: '100%' }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="dlGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#00f0ff" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#00f0ff" stopOpacity={0.0} />
                  </linearGradient>
                  <linearGradient id="ulGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#9d4edd" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#9d4edd" stopOpacity={0.0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" stroke="#475569" fontSize={11} tickLine={false} />
                <YAxis stroke="#475569" fontSize={11} tickLine={false} />
                <Tooltip
                  contentStyle={{
                    background: 'rgba(14, 20, 36, 0.95)',
                    borderColor: 'rgba(56, 189, 248, 0.3)',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                />
                <Area type="monotone" dataKey="dl" stroke="#00f0ff" strokeWidth={2.5} fillOpacity={1} fill="url(#dlGradient)" />
                <Area type="monotone" dataKey="ul" stroke="#9d4edd" strokeWidth={2.5} fillOpacity={1} fill="url(#ulGradient)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Donut: Traffic Allocation Distribution */}
        <div className="glass-panel" style={{ padding: '22px' }}>
          <h3 style={{ fontSize: '1.0625rem', color: '#ffffff', marginBottom: '4px' }}>Traffic Allocation</h3>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '14px' }}>
            Normalized multi-path weights
          </p>

          <div style={{ height: '140px', width: '100%', position: 'relative' }}>
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={42}
                  outerRadius={62}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} stroke="none" />
                  ))}
                </Pie>
              </PieChart>
            </ResponsiveContainer>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '10px' }}>
            {trafficDistribution.map((entry, idx) => (
              <div
                key={entry.interfaceName}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  fontSize: '0.8125rem',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span
                    style={{
                      width: '8px',
                      height: '8px',
                      borderRadius: '50%',
                      background: COLORS[idx % COLORS.length],
                    }}
                  />
                  <span style={{ color: 'var(--text-secondary)' }}>{entry.interfaceName}</span>
                </div>
                <span className="mono" style={{ color: '#ffffff', fontWeight: 600 }}>
                  {entry.weightPct.toFixed(1)}%
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Bottom Row: Interface Quick Cards & Recent Event Audit */}
      <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: '20px' }}>
        {/* Interface Health Cards */}
        <div className="glass-panel" style={{ padding: '22px' }}>
          <h3 style={{ fontSize: '1.0625rem', color: '#ffffff', marginBottom: '4px' }}>Active Interface Status</h3>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '16px' }}>
            Per-interface latency, packet loss, stability index, and live weight
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {interfaces.map((inf) => (
              <div
                key={inf.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '12px 16px',
                  borderRadius: '10px',
                  background: 'rgba(11, 16, 30, 0.7)',
                  border: '1px solid var(--border-subtle)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
                  <div
                    style={{
                      width: '36px',
                      height: '36px',
                      borderRadius: '8px',
                      background: 'rgba(255, 255, 255, 0.04)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    {getInterfaceIcon(inf.type)}
                  </div>
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <span style={{ fontWeight: 600, color: '#ffffff', fontSize: '0.9375rem' }}>{inf.name}</span>
                      {getStatusBadge(inf.status)}
                    </div>
                    <span className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      {inf.ipAddress} • {inf.type.toUpperCase()}
                    </span>
                  </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
                  <div style={{ textAlign: 'right' }}>
                    <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', display: 'block' }}>Latency</span>
                    <span className="mono" style={{ color: 'var(--neon-cyan)', fontSize: '0.875rem', fontWeight: 600 }}>
                      {inf.latencyMs.toFixed(1)} ms
                    </span>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', display: 'block' }}>Loss</span>
                    <span
                      className="mono"
                      style={{
                        color: inf.packetLoss > 2 ? 'var(--neon-crimson)' : 'var(--neon-emerald)',
                        fontSize: '0.875rem',
                        fontWeight: 600,
                      }}
                    >
                      {inf.packetLoss.toFixed(1)}%
                    </span>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <span style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', display: 'block' }}>Allocation</span>
                    <span className="mono" style={{ color: '#ffffff', fontSize: '0.875rem', fontWeight: 700 }}>
                      {inf.allocatedWeightPct.toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Live Event Audit Log */}
        <div className="glass-panel" style={{ padding: '22px' }}>
          <h3 style={{ fontSize: '1.0625rem', color: '#ffffff', marginBottom: '4px' }}>Live Network Events</h3>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '16px' }}>
            Real-time failovers, rebalancing & telemetry state transitions
          </p>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: '10px',
              maxHeight: '260px',
              overflowY: 'auto',
              paddingRight: '6px',
            }}
          >
            {events.slice(0, 8).map((evt) => (
              <div
                key={evt.id}
                style={{
                  padding: '10px 12px',
                  borderRadius: '8px',
                  background: 'rgba(11, 16, 30, 0.6)',
                  borderLeft: `3px solid ${
                    evt.severity === 'critical'
                      ? '#ef4444'
                      : evt.severity === 'warning'
                      ? '#f59e0b'
                      : evt.severity === 'success'
                      ? '#10b981'
                      : '#38bdf8'
                  }`,
                  fontSize: '0.75rem',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                  <span
                    style={{
                      fontWeight: 700,
                      textTransform: 'uppercase',
                      color:
                        evt.severity === 'critical'
                          ? '#ef4444'
                          : evt.severity === 'warning'
                          ? '#f59e0b'
                          : evt.severity === 'success'
                          ? '#10b981'
                          : '#38bdf8',
                    }}
                  >
                    {evt.type} {evt.interfaceName ? `• ${evt.interfaceName}` : ''}
                  </span>
                  <span className="mono" style={{ color: 'var(--text-muted)', fontSize: '0.6875rem' }}>
                    {new Date(evt.createdAt).toLocaleTimeString()}
                  </span>
                </div>
                <p style={{ color: 'var(--text-secondary)', lineHeight: 1.4 }}>{evt.message}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
