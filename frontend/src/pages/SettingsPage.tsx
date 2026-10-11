import React, { useState } from 'react'
import {
  AlertTriangle,
  Radio,
  Power,
  RefreshCw,
  TrendingDown,
  Zap,
  Info,
} from 'lucide-react'
import { useNetFusion } from '../context/NetFusionContext'

export const SettingsPage: React.FC = () => {
  const {
    simulationMode,
    setSimulationMode,
    triggerChaos,
    interfaces,
    mptcpStatus,
  } = useNetFusion()

  const [selectedIface, setSelectedIface] = useState<string>(interfaces[0]?.name || 'sim-wlan0')
  const [latencyMs, setLatencyMs] = useState(120)
  const [lossPct, setLossPct] = useState(8)
  const [busy, setBusy] = useState(false)

  const handleChaosAction = async (action: 'disconnect' | 'degrade' | 'spike' | 'restore' | 'reset') => {
    setBusy(true)
    try {
      await triggerChaos(action, selectedIface, latencyMs, lossPct)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          System Settings & Chaos Control Center
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Configure Linux engine parameters, toggle real/simulation mode, and execute interactive network chaos injection.
        </p>
      </div>

      {/* Operational Mode Toggle */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h3 style={{ fontSize: '1.125rem', color: '#ffffff', marginBottom: '4px' }}>
              Engine Operational Mode
            </h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Switch between simulated multi-interface environment (for interactive testing & demos) and real Linux hardware interfaces (`wlan0`, `eth0`, `wwan0`).
            </p>
          </div>

          <button
            onClick={() => setSimulationMode(!simulationMode)}
            className={simulationMode ? 'btn btn-purple' : 'btn btn-cyan'}
            style={{ padding: '10px 20px', borderRadius: '10px' }}
          >
            <Radio size={16} />
            <span>{simulationMode ? 'SIMULATION MODE (ACTIVE)' : 'REAL LINUX MODE (ACTIVE)'}</span>
          </button>
        </div>
      </div>

      {/* Chaos Control Center (Enabled in Simulation Mode) */}
      <div
        className="glass-panel"
        style={{
          padding: '24px',
          border: simulationMode ? '1px solid rgba(239, 68, 68, 0.4)' : '1px solid var(--border-subtle)',
          boxShadow: simulationMode ? '0 0 30px rgba(239, 68, 68, 0.15)' : 'none',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '18px' }}>
          <AlertTriangle size={20} color="#ef4444" />
          <div>
            <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Interactive Chaos Fault Injection</h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Test automatic failover and seamless dynamic rebalancing under synthetic degradation and connection drops
            </p>
          </div>
        </div>

        {!simulationMode ? (
          <div
            style={{
              padding: '16px',
              borderRadius: '8px',
              background: 'rgba(245, 158, 11, 0.1)',
              color: 'var(--neon-amber)',
              fontSize: '0.8125rem',
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
            }}
          >
            <Info size={16} />
            <span>
              Chaos triggers are only permitted when <strong>Simulation Mode</strong> is enabled to prevent disrupting real Linux host networking.
            </span>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr 1fr', gap: '16px', alignItems: 'flex-end' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Target Interface:
                </label>
                <select
                  className="select-field"
                  value={selectedIface}
                  onChange={(e) => setSelectedIface(e.target.value)}
                >
                  {interfaces.map((inf) => (
                    <option key={inf.name} value={inf.name}>
                      {inf.name} ({inf.type.toUpperCase()}) — {inf.status.toUpperCase()}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Latency Degradation (ms):
                </label>
                <input
                  type="number"
                  className="input-field"
                  value={latencyMs}
                  onChange={(e) => setLatencyMs(Number(e.target.value))}
                  min={10}
                  max={1000}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Packet Loss Rate (%):
                </label>
                <input
                  type="number"
                  className="input-field"
                  value={lossPct}
                  onChange={(e) => setLossPct(Number(e.target.value))}
                  min={0}
                  max={50}
                />
              </div>
            </div>

            {/* Action Buttons */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px' }}>
              <button
                onClick={() => handleChaosAction('disconnect')}
                disabled={busy}
                className="btn btn-danger"
              >
                <Power size={14} />
                <span>Disconnect Interface</span>
              </button>

              <button
                onClick={() => handleChaosAction('degrade')}
                disabled={busy}
                className="btn btn-ghost"
                style={{ borderColor: 'rgba(245, 158, 11, 0.4)', color: 'var(--neon-amber)' }}
              >
                <TrendingDown size={14} />
                <span>Degrade Quality</span>
              </button>

              <button
                onClick={() => handleChaosAction('spike')}
                disabled={busy}
                className="btn btn-ghost"
                style={{ borderColor: 'rgba(157, 78, 221, 0.4)', color: 'var(--neon-purple)' }}
              >
                <Zap size={14} />
                <span>Spike Latency (10s)</span>
              </button>

              <button
                onClick={() => handleChaosAction('restore')}
                disabled={busy}
                className="btn btn-ghost"
                style={{ borderColor: 'rgba(16, 185, 129, 0.4)', color: 'var(--neon-emerald)' }}
              >
                <RefreshCw size={14} />
                <span>Restore Clean Link</span>
              </button>
            </div>

            <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '14px', textAlign: 'right' }}>
              <button
                onClick={() => handleChaosAction('reset')}
                disabled={busy}
                className="btn btn-ghost"
                style={{ fontSize: '0.75rem' }}
              >
                Reset All Chaos Faults Across All Interfaces
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Linux Kernel Diagnostics Panel */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <h3 style={{ fontSize: '1.125rem', color: '#ffffff', marginBottom: '16px' }}>
          Linux MPTCP Kernel Diagnostics
        </h3>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px' }}>
          <div style={{ background: 'rgba(11, 16, 30, 0.6)', padding: '16px', borderRadius: '10px' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block', marginBottom: '4px' }}>
              Kernel MPTCP Support
            </span>
            <span className="mono" style={{ color: mptcpStatus?.enabled ? 'var(--neon-emerald)' : 'var(--neon-amber)', fontWeight: 700 }}>
              {mptcpStatus?.enabled ? 'SUPPORTED & ACTIVE' : 'DISABLED / NOT PRESENT'}
            </span>
          </div>

          <div style={{ background: 'rgba(11, 16, 30, 0.6)', padding: '16px', borderRadius: '10px' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block', marginBottom: '4px' }}>
              Path Manager Type
            </span>
            <span className="mono" style={{ color: '#ffffff', fontWeight: 600 }}>
              {mptcpStatus?.pathManager || 'in-kernel'}
            </span>
          </div>

          <div style={{ background: 'rgba(11, 16, 30, 0.6)', padding: '16px', borderRadius: '10px' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block', marginBottom: '4px' }}>
              Active Endpoints Count
            </span>
            <span className="mono" style={{ color: 'var(--neon-cyan)', fontWeight: 700 }}>
              {mptcpStatus?.endpoints?.length || 0} Endpoints Registered
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
