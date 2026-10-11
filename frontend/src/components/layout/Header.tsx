import React from 'react'
import {
  Activity,
  Cpu,
  Radio,
  RefreshCw,
  Zap,
  AlertTriangle,
} from 'lucide-react'
import { useNetFusion } from '../../context/NetFusionContext'

interface HeaderProps {
  onOpenChaos: () => void
}

export const Header: React.FC<HeaderProps> = ({ onOpenChaos }) => {
  const {
    wsConnected,
    mptcpStatus,
    simulationMode,
    setSimulationMode,
    refreshData,
    notification,
    clearNotification,
  } = useNetFusion()

  return (
    <header
      style={{
        display: 'flex',
        flexDirection: 'column',
        borderBottom: '1px solid var(--border-subtle)',
        background: 'rgba(7, 10, 19, 0.85)',
        backdropFilter: 'blur(20px)',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      {/* Top Banner / Notification */}
      {notification && (
        <div
          style={{
            background: 'linear-gradient(90deg, rgba(0, 240, 255, 0.15), rgba(157, 78, 221, 0.15))',
            borderBottom: '1px solid rgba(0, 240, 255, 0.3)',
            padding: '6px 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            fontSize: '0.8125rem',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Zap size={14} color="var(--neon-cyan)" />
            <span style={{ color: 'var(--text-primary)', fontWeight: 500 }}>{notification}</span>
          </div>
          <button
            onClick={clearNotification}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'var(--text-muted)',
              cursor: 'pointer',
              fontSize: '0.75rem',
            }}
          >
            ✕
          </button>
        </div>
      )}

      {/* Main Header Bar */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '12px 28px',
          height: '64px',
        }}
      >
        {/* Brand */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
          <div
            style={{
              width: '38px',
              height: '38px',
              borderRadius: '10px',
              background: 'linear-gradient(135deg, #00f0ff 0%, #7000ff 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 0 20px rgba(0, 240, 255, 0.4)',
            }}
          >
            <Activity size={22} color="#050814" strokeWidth={2.5} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <span
                style={{
                  fontFamily: 'var(--font-heading)',
                  fontSize: '1.25rem',
                  fontWeight: 800,
                  letterSpacing: '-0.02em',
                  background: 'linear-gradient(90deg, #ffffff 0%, #00f0ff 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                NETFUSION
              </span>
              <span
                style={{
                  fontSize: '0.625rem',
                  fontWeight: 700,
                  letterSpacing: '0.08em',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  background: 'rgba(56, 189, 248, 0.1)',
                  color: 'var(--neon-cyan)',
                  border: '1px solid rgba(56, 189, 248, 0.25)',
                }}
              >
                v1.0
              </span>
            </div>
            <p style={{ fontSize: '0.6875rem', color: 'var(--text-muted)', marginTop: '-2px' }}>
              Intelligent Multi-Network Traffic Controller
            </p>
          </div>
        </div>

        {/* Status Center Badges */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          {/* WebSocket Status */}
          <div
            className="badge"
            style={{
              background: wsConnected ? 'rgba(16, 185, 129, 0.12)' : 'rgba(239, 68, 68, 0.12)',
              color: wsConnected ? '#10b981' : '#ef4444',
              border: `1px solid ${wsConnected ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
            }}
          >
            <span
              className="pulse-dot"
              style={{
                background: wsConnected ? '#10b981' : '#ef4444',
                boxShadow: wsConnected ? '0 0 8px #10b981' : '0 0 8px #ef4444',
              }}
            />
            <span>{wsConnected ? 'LIVE TELEMETRY' : 'CONNECTING WS'}</span>
          </div>

          {/* MPTCP Status */}
          <div
            className="badge"
            style={{
              background: mptcpStatus?.enabled ? 'rgba(56, 189, 248, 0.12)' : 'rgba(245, 158, 11, 0.12)',
              color: mptcpStatus?.enabled ? '#38bdf8' : '#f59e0b',
              border: `1px solid ${mptcpStatus?.enabled ? 'rgba(56, 189, 248, 0.3)' : 'rgba(245, 158, 11, 0.3)'}`,
            }}
          >
            <Cpu size={12} />
            <span>
              {mptcpStatus?.enabled
                ? `MPTCP IN-KERNEL (${mptcpStatus.endpoints.length} PATHS)`
                : 'MPTCP FALLBACK'}
            </span>
          </div>

          {/* Simulation vs Real Mode Switch */}
          <button
            onClick={() => setSimulationMode(!simulationMode)}
            className="btn btn-ghost"
            style={{
              fontSize: '0.75rem',
              padding: '4px 12px',
              borderRadius: '9999px',
              borderColor: simulationMode ? 'rgba(157, 78, 221, 0.4)' : 'rgba(16, 185, 129, 0.4)',
              color: simulationMode ? '#c084fc' : '#34d399',
            }}
            title="Click to toggle between Simulation Mode and Real Linux Interface Mode"
          >
            <Radio size={13} />
            <span>{simulationMode ? 'SIMULATION MODE' : 'REAL LINUX MODE'}</span>
          </button>

          {/* Chaos Quick Trigger */}
          {simulationMode && (
            <button
              onClick={onOpenChaos}
              className="btn btn-danger"
              style={{
                fontSize: '0.75rem',
                padding: '5px 12px',
                borderRadius: '8px',
              }}
            >
              <AlertTriangle size={13} />
              <span>CHAOS INJECTOR</span>
            </button>
          )}

          {/* Manual Refresh */}
          <button
            onClick={refreshData}
            className="btn btn-ghost"
            style={{ padding: '6px', borderRadius: '8px' }}
            title="Refresh All Telemetry"
          >
            <RefreshCw size={15} />
          </button>
        </div>
      </div>
    </header>
  )
}
