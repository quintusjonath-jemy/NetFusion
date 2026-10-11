import React, { useState } from 'react'
import {
  AlertTriangle,
  Zap,
  TrendingDown,
  RefreshCw,
  Power,
  X,
} from 'lucide-react'
import { useNetFusion } from '../../context/NetFusionContext'

interface ChaosModalProps {
  isOpen: boolean
  onClose: () => void
}

export const ChaosModal: React.FC<ChaosModalProps> = ({ isOpen, onClose }) => {
  const { interfaces, triggerChaos } = useNetFusion()
  const [selectedIface, setSelectedIface] = useState<string>(interfaces[0]?.name || 'sim-wlan0')
  const [latency, setLatency] = useState<number>(150)
  const [loss, setLoss] = useState<number>(10)
  const [loading, setLoading] = useState(false)

  if (!isOpen) return null

  const handleAction = async (action: 'disconnect' | 'degrade' | 'spike' | 'restore' | 'reset') => {
    setLoading(true)
    try {
      await triggerChaos(action, selectedIface, latency, loss)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(5, 7, 15, 0.8)',
        backdropFilter: 'blur(12px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
      }}
      onClick={onClose}
    >
      <div
        className="glass-panel"
        style={{
          width: '100%',
          maxWidth: '520px',
          background: 'rgba(12, 18, 34, 0.95)',
          border: '1px solid rgba(239, 68, 68, 0.4)',
          boxShadow: '0 0 40px rgba(239, 68, 68, 0.25)',
          padding: '24px',
          borderRadius: '16px',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '20px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(239, 68, 68, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                border: '1px solid rgba(239, 68, 68, 0.3)',
              }}
            >
              <AlertTriangle size={18} color="#ef4444" />
            </div>
            <div>
              <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Chaos Fault Injector</h3>
              <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Inject latency, drop packets, or disconnect interfaces to test automatic failover
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'var(--text-muted)',
              cursor: 'pointer',
            }}
          >
            <X size={18} />
          </button>
        </div>

        {/* Interface Selector */}
        <div style={{ marginBottom: '16px' }}>
          <label
            style={{
              display: 'block',
              fontSize: '0.75rem',
              fontWeight: 600,
              color: 'var(--text-secondary)',
              marginBottom: '6px',
            }}
          >
            Target Network Interface:
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

        {/* Sliders for Degrade / Spike */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '14px', marginBottom: '20px' }}>
          <div>
            <label
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: '0.75rem',
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              <span>Latency Injection:</span>
              <span className="mono" style={{ color: 'var(--neon-amber)' }}>
                {latency} ms
              </span>
            </label>
            <input
              type="range"
              min={20}
              max={500}
              step={10}
              value={latency}
              onChange={(e) => setLatency(Number(e.target.value))}
              style={{ width: '100%', accentColor: 'var(--neon-amber)' }}
            />
          </div>

          <div>
            <label
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: '0.75rem',
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              <span>Packet Loss Rate:</span>
              <span className="mono" style={{ color: 'var(--neon-crimson)' }}>
                {loss}%
              </span>
            </label>
            <input
              type="range"
              min={0}
              max={40}
              step={1}
              value={loss}
              onChange={(e) => setLoss(Number(e.target.value))}
              style={{ width: '100%', accentColor: 'var(--neon-crimson)' }}
            />
          </div>
        </div>

        {/* Actions Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
          <button
            onClick={() => handleAction('disconnect')}
            disabled={loading}
            className="btn btn-danger"
            style={{ width: '100%' }}
          >
            <Power size={14} />
            <span>Sever Link (Disconnect)</span>
          </button>

          <button
            onClick={() => handleAction('degrade')}
            disabled={loading}
            className="btn btn-ghost"
            style={{
              width: '100%',
              borderColor: 'rgba(245, 158, 11, 0.4)',
              color: 'var(--neon-amber)',
            }}
          >
            <TrendingDown size={14} />
            <span>Degrade Quality</span>
          </button>

          <button
            onClick={() => handleAction('spike')}
            disabled={loading}
            className="btn btn-ghost"
            style={{
              width: '100%',
              borderColor: 'rgba(157, 78, 221, 0.4)',
              color: 'var(--neon-purple)',
            }}
          >
            <Zap size={14} />
            <span>Spike Latency</span>
          </button>

          <button
            onClick={() => handleAction('restore')}
            disabled={loading}
            className="btn btn-ghost"
            style={{
              width: '100%',
              borderColor: 'rgba(16, 185, 129, 0.4)',
              color: 'var(--neon-emerald)',
            }}
          >
            <RefreshCw size={14} />
            <span>Restore Interface</span>
          </button>
        </div>

        {/* Reset All */}
        <div style={{ marginTop: '16px', borderTop: '1px solid var(--border-subtle)', paddingTop: '14px' }}>
          <button
            onClick={() => handleAction('reset')}
            disabled={loading}
            className="btn btn-ghost"
            style={{ width: '100%', fontSize: '0.8125rem' }}
          >
            Reset All Injected Faults to Clean State
          </button>
        </div>
      </div>
    </div>
  )
}
