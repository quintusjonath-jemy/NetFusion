import React from 'react'
import {
  Activity,
  Network,
  Share2,
  DownloadCloud,
  Gauge,
  History,
  ShieldAlert,
  Sliders,
  Terminal,
} from 'lucide-react'
import { useNetFusion } from '../../context/NetFusionContext'

export type NavTab =
  | 'dashboard'
  | 'networks'
  | 'traffic'
  | 'downloads'
  | 'benchmark'
  | 'history'
  | 'policies'
  | 'settings'

interface SidebarProps {
  currentTab: NavTab
  onSelectTab: (tab: NavTab) => void
}

export const Sidebar: React.FC<SidebarProps> = ({ currentTab, onSelectTab }) => {
  const { totalDownloadMbps, activePathsCount, activePolicy } = useNetFusion()

  const navItems = [
    { id: 'dashboard', label: 'Overview', icon: Activity, badge: null },
    { id: 'networks', label: 'Interfaces', icon: Network, badge: `${activePathsCount} Active` },
    { id: 'traffic', label: 'Traffic & Paths', icon: Share2, badge: null },
    { id: 'downloads', label: 'Downloader', icon: DownloadCloud, badge: null },
    { id: 'benchmark', label: 'Benchmark', icon: Gauge, badge: 'Head-to-Head' },
    { id: 'history', label: 'History & Logs', icon: History, badge: null },
    { id: 'policies', label: 'Policies', icon: ShieldAlert, badge: activePolicy?.mode || null },
    { id: 'settings', label: 'Chaos & System', icon: Sliders, badge: null },
  ] as const

  return (
    <aside
      style={{
        width: '260px',
        borderRight: '1px solid var(--border-subtle)',
        background: 'rgba(8, 12, 22, 0.75)',
        backdropFilter: 'blur(20px)',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: '20px 14px',
        flexShrink: 0,
        height: 'calc(100vh - 64px)',
        position: 'sticky',
        top: '64px',
      }}
    >
      {/* Navigation Links */}
      <nav style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
        <p
          style={{
            fontSize: '0.6875rem',
            textTransform: 'uppercase',
            letterSpacing: '0.08em',
            color: 'var(--text-muted)',
            padding: '0 12px 6px',
            fontWeight: 700,
          }}
        >
          Navigation
        </p>

        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = currentTab === item.id

          return (
            <button
              key={item.id}
              onClick={() => onSelectTab(item.id)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '10px 14px',
                borderRadius: '10px',
                border: isActive
                  ? '1px solid rgba(0, 240, 255, 0.4)'
                  : '1px solid transparent',
                background: isActive
                  ? 'linear-gradient(90deg, rgba(0, 240, 255, 0.12), rgba(112, 0, 255, 0.05))'
                  : 'transparent',
                color: isActive ? 'var(--neon-cyan)' : 'var(--text-secondary)',
                cursor: 'pointer',
                textAlign: 'left',
                transition: 'all 0.2s ease',
                boxShadow: isActive ? '0 0 16px rgba(0, 240, 255, 0.15)' : 'none',
              }}
              onMouseEnter={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = 'rgba(255, 255, 255, 0.04)'
                  e.currentTarget.style.color = 'var(--text-primary)'
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = 'transparent'
                  e.currentTarget.style.color = 'var(--text-secondary)'
                }
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <Icon size={18} strokeWidth={isActive ? 2.5 : 2} />
                <span
                  style={{
                    fontSize: '0.875rem',
                    fontWeight: isActive ? 600 : 500,
                  }}
                >
                  {item.label}
                </span>
              </div>

              {item.badge && (
                <span
                  style={{
                    fontSize: '0.6875rem',
                    padding: '2px 8px',
                    borderRadius: '9999px',
                    background: isActive
                      ? 'rgba(0, 240, 255, 0.2)'
                      : 'rgba(255, 255, 255, 0.05)',
                    color: isActive ? 'var(--neon-cyan)' : 'var(--text-muted)',
                    fontWeight: 600,
                    textTransform: 'capitalize',
                  }}
                >
                  {item.badge}
                </span>
              )}
            </button>
          )
        })}
      </nav>

      {/* Bottom Telemetry Mini HUD */}
      <div
        className="glass-panel"
        style={{
          padding: '14px',
          background: 'rgba(11, 16, 30, 0.9)',
          border: '1px solid rgba(56, 189, 248, 0.15)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '8px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <Terminal size={13} color="var(--neon-cyan)" />
            <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
              CORE ENGINE
            </span>
          </div>
          <span
            style={{
              fontSize: '0.6875rem',
              color: 'var(--neon-emerald)',
              fontWeight: 700,
            }}
          >
            ACTIVE
          </span>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem' }}>
            <span style={{ color: 'var(--text-muted)' }}>Aggregate Speed:</span>
            <span className="mono" style={{ color: 'var(--neon-cyan)', fontWeight: 600 }}>
              {totalDownloadMbps.toFixed(1)} Mbps
            </span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem' }}>
            <span style={{ color: 'var(--text-muted)' }}>Active Subflows:</span>
            <span className="mono" style={{ color: '#ffffff', fontWeight: 600 }}>
              {activePathsCount} Paths
            </span>
          </div>
        </div>
      </div>
    </aside>
  )
}
