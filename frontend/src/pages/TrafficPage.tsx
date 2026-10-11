import React from 'react'
import {
  Share2,
  Cpu,
  Layers,
  HardDrive,
  Wifi,
  Signal,
  CheckCircle2,
} from 'lucide-react'
import { useNetFusion } from '../context/NetFusionContext'

export const TrafficPage: React.FC = () => {
  const { trafficDistribution, totalDownloadMbps, mptcpStatus, activePolicy } = useNetFusion()

  const getInterfaceIcon = (type: string) => {
    switch (type) {
      case 'ethernet':
        return <HardDrive size={18} color="var(--neon-cyan)" />
      case 'wifi':
        return <Wifi size={18} color="var(--neon-purple)" />
      case 'cellular':
        return <Signal size={18} color="var(--neon-amber)" />
      default:
        return <Layers size={18} color="var(--neon-blue)" />
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Real-Time Traffic Allocation & MPTCP Subflow Topology
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Dynamic multi-path traffic splitting, normalized interface weights, and in-kernel Multipath TCP topology.
        </p>
      </div>

      {/* Visual Multi-Path Topology Diagram */}
      <div className="glass-panel" style={{ padding: '28px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
          <div>
            <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Multipath Traffic Distribution Flow</h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Traffic multiplexing based on policy '{activePolicy?.name || 'Balanced Adaptive Mode'}'
            </p>
          </div>
          <span
            className="badge"
            style={{
              background: 'rgba(0, 240, 255, 0.1)',
              color: 'var(--neon-cyan)',
              border: '1px solid rgba(0, 240, 255, 0.3)',
            }}
          >
            Aggregated: {totalDownloadMbps.toFixed(1)} Mbps
          </span>
        </div>

        {/* Visual Flow Architecture */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr auto 1fr',
            gap: '30px',
            alignItems: 'center',
            padding: '20px 0',
          }}
        >
          {/* Left: Client Applications */}
          <div
            style={{
              background: 'rgba(12, 18, 36, 0.8)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '12px',
              padding: '20px',
              textAlign: 'center',
            }}
          >
            <div
              style={{
                width: '44px',
                height: '44px',
                borderRadius: '10px',
                background: 'linear-gradient(135deg, #00f0ff, #7000ff)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 12px',
              }}
            >
              <Cpu size={24} color="#050814" />
            </div>
            <h4 style={{ color: '#ffffff', fontSize: '1rem', marginBottom: '4px' }}>NetFusion Engine</h4>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Dynamic Multi-Interface Multiplexer
            </p>
            <span
              className="badge"
              style={{
                marginTop: '10px',
                background: 'rgba(16, 185, 129, 0.15)',
                color: '#10b981',
                fontSize: '0.6875rem',
              }}
            >
              Single Virtual Socket
            </span>
          </div>

          {/* Middle: Flow Connector Icon */}
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px' }}>
            <Share2 size={28} color="var(--neon-cyan)" />
            <span style={{ fontSize: '0.6875rem', color: 'var(--neon-cyan)', fontWeight: 600 }}>
              PATH MANAGER
            </span>
          </div>

          {/* Right: Subflow Branches */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {trafficDistribution.map((entry) => (
              <div
                key={entry.interfaceName}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '12px 16px',
                  borderRadius: '10px',
                  background: 'rgba(14, 22, 42, 0.9)',
                  border: `1px solid ${
                    entry.weightPct > 0 ? 'rgba(56, 189, 248, 0.3)' : 'rgba(239, 68, 68, 0.3)'
                  }`,
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  {getInterfaceIcon(entry.type)}
                  <div>
                    <span style={{ fontWeight: 600, color: '#ffffff', fontSize: '0.875rem' }}>
                      {entry.interfaceName}
                    </span>
                    <span style={{ display: 'block', fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                      {entry.type.toUpperCase()} • {entry.status.toUpperCase()}
                    </span>
                  </div>
                </div>

                <div style={{ textAlign: 'right' }}>
                  <span
                    className="mono"
                    style={{
                      fontSize: '1.125rem',
                      fontWeight: 700,
                      color: entry.weightPct > 0 ? 'var(--neon-cyan)' : 'var(--neon-crimson)',
                    }}
                  >
                    {entry.weightPct.toFixed(1)}%
                  </span>
                  <span style={{ display: 'block', fontSize: '0.6875rem', color: 'var(--text-muted)' }}>
                    {entry.downloadMbps.toFixed(1)} Mbps DL
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Linux MPTCP Endpoints & Kernel Inspection */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '18px' }}>
          <div>
            <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Linux Kernel MPTCP Subflow Endpoints</h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Native Linux in-kernel path manager endpoints (`ip mptcp endpoint show`)
            </p>
          </div>
          <span
            className="badge"
            style={{
              background: mptcpStatus?.enabled ? 'rgba(16, 185, 129, 0.15)' : 'rgba(245, 158, 11, 0.15)',
              color: mptcpStatus?.enabled ? '#10b981' : '#f59e0b',
            }}
          >
            {mptcpStatus?.enabled ? 'In-Kernel PM Active' : 'Fallback Mode'}
          </span>
        </div>

        {mptcpStatus?.endpoints && mptcpStatus.endpoints.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Endpoint ID</th>
                <th>Interface</th>
                <th>Subflow IP Address</th>
                <th>MPTCP Flags</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {mptcpStatus.endpoints.map((ep) => (
                <tr key={ep.id}>
                  <td className="mono" style={{ color: 'var(--neon-cyan)', fontWeight: 600 }}>
                    #{ep.id}
                  </td>
                  <td style={{ fontWeight: 600, color: '#ffffff' }}>{ep.interface || 'Default'}</td>
                  <td className="mono" style={{ color: 'var(--text-secondary)' }}>
                    {ep.ip}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: '6px' }}>
                      {ep.flags.map((flag) => (
                        <span
                          key={flag}
                          style={{
                            fontSize: '0.6875rem',
                            padding: '2px 6px',
                            borderRadius: '4px',
                            background: 'rgba(56, 189, 248, 0.1)',
                            color: 'var(--neon-cyan)',
                            border: '1px solid rgba(56, 189, 248, 0.25)',
                          }}
                        >
                          {flag}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td>
                    <span className="badge badge-connected">
                      <CheckCircle2 size={10} /> Active Subflow
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div
            style={{
              padding: '24px',
              textAlign: 'center',
              background: 'rgba(10, 15, 29, 0.5)',
              borderRadius: '8px',
              color: 'var(--text-muted)',
              fontSize: '0.875rem',
            }}
          >
            No kernel MPTCP endpoints configured. Real-time traffic allocation is currently managed via Linux `SO_BINDTODEVICE` and `LocalAddr` multiplexing.
          </div>
        )}
      </div>
    </div>
  )
}
