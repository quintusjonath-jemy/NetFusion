import React, { useState } from 'react'
import {
  Shield,
  Zap,
  Scale,
  CheckCircle2,
  Sliders,
  DollarSign,
} from 'lucide-react'
import { useNetFusion } from '../context/NetFusionContext'

export const PoliciesPage: React.FC = () => {
  const { policies, activePolicy, activatePolicy } = useNetFusion()
  const [activatingId, setActivatingId] = useState<string | null>(null)

  const handleActivate = async (id: string) => {
    setActivatingId(id)
    try {
      await activatePolicy(id)
    } finally {
      setActivatingId(null)
    }
  }

  const getPolicyIcon = (mode: string) => {
    switch (mode) {
      case 'performance':
        return <Zap size={22} color="var(--neon-cyan)" />
      case 'reliability':
        return <Shield size={22} color="var(--neon-emerald)" />
      case 'balanced':
        return <Scale size={22} color="var(--neon-purple)" />
      case 'cost_saving':
        return <DollarSign size={22} color="var(--neon-amber)" />
      default:
        return <Sliders size={22} color="var(--neon-blue)" />
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Traffic Allocation Policies & Failover Rules
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Select or configure dynamic weighting algorithms governing multi-interface traffic normalization and cellular data throttling.
        </p>
      </div>

      {/* Grid of Policy Cards */}
      <div className="grid-2">
        {policies.map((p) => {
          const isActive = activePolicy?.id === p.id

          return (
            <div
              key={p.id}
              className={isActive ? 'glass-panel-glow' : 'glass-panel'}
              style={{
                padding: '24px',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
                position: 'relative',
              }}
            >
              <div>
                {/* Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '14px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
                    <div
                      style={{
                        width: '42px',
                        height: '42px',
                        borderRadius: '10px',
                        background: 'rgba(255, 255, 255, 0.05)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        border: '1px solid var(--border-subtle)',
                      }}
                    >
                      {getPolicyIcon(p.mode)}
                    </div>
                    <div>
                      <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>{p.name}</h3>
                      <span className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        Mode: {p.mode.toUpperCase()}
                      </span>
                    </div>
                  </div>

                  {isActive ? (
                    <span
                      className="badge badge-connected"
                      style={{ background: 'rgba(16, 185, 129, 0.2)', padding: '4px 12px' }}
                    >
                      <CheckCircle2 size={12} /> Active Policy
                    </span>
                  ) : (
                    <button
                      onClick={() => handleActivate(p.id)}
                      disabled={activatingId === p.id}
                      className="btn btn-ghost"
                      style={{ fontSize: '0.75rem', padding: '5px 12px' }}
                    >
                      {activatingId === p.id ? 'Applying...' : 'Activate Mode'}
                    </button>
                  )}
                </div>

                {/* Weights Breakdown Bars */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '18px' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', fontWeight: 600 }}>
                    Algorithm Component Weights:
                  </span>

                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.6875rem', marginBottom: '3px' }}>
                      <span style={{ color: 'var(--text-secondary)' }}>Bandwidth Preference</span>
                      <span className="mono" style={{ color: 'var(--neon-cyan)' }}>
                        {(p.weights.bandwidthWeight * 100).toFixed(0)}%
                      </span>
                    </div>
                    <div style={{ height: '4px', background: 'rgba(255, 255, 255, 0.06)', borderRadius: '2px' }}>
                      <div
                        style={{
                          height: '100%',
                          width: `${p.weights.bandwidthWeight * 100}%`,
                          background: 'var(--neon-cyan)',
                        }}
                      />
                    </div>
                  </div>

                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.6875rem', marginBottom: '3px' }}>
                      <span style={{ color: 'var(--text-secondary)' }}>Latency Sensitivity</span>
                      <span className="mono" style={{ color: 'var(--neon-purple)' }}>
                        {(p.weights.latencyWeight * 100).toFixed(0)}%
                      </span>
                    </div>
                    <div style={{ height: '4px', background: 'rgba(255, 255, 255, 0.06)', borderRadius: '2px' }}>
                      <div
                        style={{
                          height: '100%',
                          width: `${p.weights.latencyWeight * 100}%`,
                          background: 'var(--neon-purple)',
                        }}
                      />
                    </div>
                  </div>

                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.6875rem', marginBottom: '3px' }}>
                      <span style={{ color: 'var(--text-secondary)' }}>Cellular Cost Penalty</span>
                      <span className="mono" style={{ color: 'var(--neon-amber)' }}>
                        {(p.weights.costPenalty * 100).toFixed(0)}%
                      </span>
                    </div>
                    <div style={{ height: '4px', background: 'rgba(255, 255, 255, 0.06)', borderRadius: '2px' }}>
                      <div
                        style={{
                          height: '100%',
                          width: `${p.weights.costPenalty * 100}%`,
                          background: 'var(--neon-amber)',
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Bottom Policy Flags */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  paddingTop: '12px',
                  borderTop: '1px solid var(--border-subtle)',
                  fontSize: '0.75rem',
                  color: 'var(--text-muted)',
                }}
              >
                <span>Degraded Latency Threshold: {p.degradedThresholdLatencyMs} ms</span>
                <span>Max Packet Loss: {p.degradedThresholdLossPct}%</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
