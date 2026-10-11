import React, { useState, useEffect } from 'react'
import {
  Play,
  Square,
  AlertCircle,
  RefreshCw,
} from 'lucide-react'
import type { DownloadSession } from '../types'
import { api } from '../services/api'
import { useNetFusion } from '../context/NetFusionContext'

export const DownloadsPage: React.FC = () => {
  const { interfaces } = useNetFusion()
  const [downloads, setDownloads] = useState<DownloadSession[]>([])
  const [url, setUrl] = useState('https://speed.cloudflare.com/__down?bytes=50000000')
  const [mode, setMode] = useState<'multi' | 'single'>('multi')
  const [selectedInterface, setSelectedInterface] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchDownloads = async () => {
    try {
      const data = await api.getDownloads()
      setDownloads(data || [])
    } catch (err: unknown) {
      console.error('Failed to load downloads:', err)
    }
  }

  useEffect(() => {
    fetchDownloads()
    const timer = setInterval(fetchDownloads, 2000)
    return () => clearInterval(timer)
  }, [])

  const handleStart = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!url.trim()) return

    setLoading(true)
    setError(null)
    try {
      await api.createDownload(url.trim(), mode, mode === 'single' ? selectedInterface : undefined)
      await fetchDownloads()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const handleStop = async (id: string) => {
    try {
      await api.stopDownload(id)
      await fetchDownloads()
    } catch (err) {
      console.error('Failed to stop download:', err)
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Multi-Network Range Chunked Downloader
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Accelerate downloads by slicing large files into concurrent byte ranges across multiple interfaces with dynamic failover.
        </p>
      </div>

      {/* Start Download Card */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <h3 style={{ fontSize: '1.125rem', color: '#ffffff', marginBottom: '16px' }}>
          Start New Accelerated Download
        </h3>

        <form onSubmit={handleStart} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label
              style={{
                display: 'block',
                fontSize: '0.75rem',
                fontWeight: 600,
                color: 'var(--text-secondary)',
                marginBottom: '6px',
              }}
            >
              Target HTTP/HTTPS Download URL:
            </label>
            <input
              type="url"
              className="input-field"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/largefile.iso"
              required
            />
          </div>

          {/* Quick preset links */}
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', alignSelf: 'center' }}>
              Quick Presets:
            </span>
            <button
              type="button"
              onClick={() => setUrl('https://speed.cloudflare.com/__down?bytes=10000000')}
              className="btn btn-ghost"
              style={{ fontSize: '0.6875rem', padding: '3px 8px' }}
            >
              10 MB Sample
            </button>
            <button
              type="button"
              onClick={() => setUrl('https://speed.cloudflare.com/__down?bytes=50000000')}
              className="btn btn-ghost"
              style={{ fontSize: '0.6875rem', padding: '3px 8px' }}
            >
              50 MB Sample
            </button>
            <button
              type="button"
              onClick={() => setUrl('https://speed.cloudflare.com/__down?bytes=100000000')}
              className="btn btn-ghost"
              style={{ fontSize: '0.6875rem', padding: '3px 8px' }}
            >
              100 MB Sample
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: '0.75rem',
                  fontWeight: 600,
                  color: 'var(--text-secondary)',
                  marginBottom: '6px',
                }}
              >
                Aggregation Mode:
              </label>
              <select
                className="select-field"
                value={mode}
                onChange={(e) => setMode(e.target.value as 'multi' | 'single')}
              >
                <option value="multi">Multi-Network Aggregated (Split Across All Active Interfaces)</option>
                <option value="single">Single Interface (Baseline Benchmark Mode)</option>
              </select>
            </div>

            {mode === 'single' && (
              <div>
                <label
                  style={{
                    display: 'block',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    color: 'var(--text-secondary)',
                    marginBottom: '6px',
                  }}
                >
                  Bind to Interface:
                </label>
                <select
                  className="select-field"
                  value={selectedInterface}
                  onChange={(e) => setSelectedInterface(e.target.value)}
                >
                  <option value="">Auto (Default Route)</option>
                  {interfaces.map((inf) => (
                    <option key={inf.name} value={inf.name}>
                      {inf.name} ({inf.type.toUpperCase()})
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {error && (
            <div
              style={{
                padding: '10px 14px',
                borderRadius: '8px',
                background: 'rgba(239, 68, 68, 0.15)',
                color: '#ef4444',
                fontSize: '0.8125rem',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
            >
              <AlertCircle size={16} />
              <span>{error}</span>
            </div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '6px' }}>
            <button type="submit" disabled={loading} className="btn btn-cyan" style={{ padding: '10px 24px' }}>
              <Play size={16} fill="currentColor" />
              <span>{loading ? 'Initiating Sockets...' : 'Launch Range Download'}</span>
            </button>
          </div>
        </form>
      </div>

      {/* Active & Recent Downloads */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '18px' }}>
          <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Active & Historical Download Sessions</h3>
          <button onClick={fetchDownloads} className="btn btn-ghost" style={{ padding: '6px', fontSize: '0.75rem' }}>
            <RefreshCw size={14} /> Refresh
          </button>
        </div>

        {downloads.length === 0 ? (
          <div
            style={{
              padding: '36px',
              textAlign: 'center',
              color: 'var(--text-muted)',
              fontSize: '0.875rem',
            }}
          >
            No active downloads. Launch a multi-network download above to see range chunking and per-interface progress in real-time.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {downloads.map((dl) => (
              <div
                key={dl.id}
                style={{
                  background: 'rgba(11, 16, 30, 0.8)',
                  borderRadius: '12px',
                  padding: '18px 20px',
                  border: '1px solid var(--border-subtle)',
                }}
              >
                {/* Top Row: File Name & Controls */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '12px' }}>
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                      <h4 style={{ color: '#ffffff', fontSize: '1rem' }}>{dl.fileName}</h4>
                      <span
                        className="badge"
                        style={{
                          background:
                            dl.status === 'completed'
                              ? 'rgba(16, 185, 129, 0.15)'
                              : dl.status === 'downloading'
                              ? 'rgba(0, 240, 255, 0.15)'
                              : 'rgba(239, 68, 68, 0.15)',
                          color:
                            dl.status === 'completed'
                              ? '#10b981'
                              : dl.status === 'downloading'
                              ? 'var(--neon-cyan)'
                              : '#ef4444',
                        }}
                      >
                        {dl.status.toUpperCase()}
                      </span>
                      <span
                        className="badge"
                        style={{
                          background: 'rgba(157, 78, 221, 0.15)',
                          color: '#c084fc',
                        }}
                      >
                        {dl.mode === 'multi' ? `Multi-Network (${dl.activePaths} Paths)` : 'Single Interface'}
                      </span>
                    </div>
                    <span className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-muted)', wordBreak: 'break-all' }}>
                      {dl.url}
                    </span>
                  </div>

                  {dl.status === 'downloading' && (
                    <button
                      onClick={() => handleStop(dl.id)}
                      className="btn btn-danger"
                      style={{ fontSize: '0.75rem', padding: '5px 12px' }}
                    >
                      <Square size={12} fill="currentColor" />
                      <span>Stop</span>
                    </button>
                  )}
                </div>

                {/* Progress Bar */}
                <div style={{ marginBottom: '10px' }}>
                  <div
                    style={{
                      height: '8px',
                      background: 'rgba(255, 255, 255, 0.08)',
                      borderRadius: '9999px',
                      overflow: 'hidden',
                    }}
                  >
                    <div
                      style={{
                        height: '100%',
                        width: `${Math.min(100, dl.progressPct)}%`,
                        background: 'linear-gradient(90deg, #00f0ff 0%, #7000ff 100%)',
                        boxShadow: '0 0 12px rgba(0, 240, 255, 0.5)',
                        transition: 'width 0.3s ease',
                      }}
                    />
                  </div>
                </div>

                {/* Stats Bar */}
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                  <div>
                    <span className="mono" style={{ color: '#ffffff', fontWeight: 600 }}>
                      {formatBytes(dl.downloadedBytes)}
                    </span>
                    <span style={{ color: 'var(--text-muted)' }}>
                      {' '}
                      / {dl.fileSize > 0 ? formatBytes(dl.fileSize) : 'Estimating...'} ({dl.progressPct.toFixed(1)}%)
                    </span>
                  </div>

                  <div style={{ display: 'flex', gap: '20px' }}>
                    <div>
                      <span style={{ color: 'var(--text-muted)' }}>Speed: </span>
                      <span className="mono" style={{ color: 'var(--neon-cyan)', fontWeight: 600 }}>
                        {dl.speedMbps.toFixed(1)} Mbps
                      </span>
                    </div>
                    <div>
                      <span style={{ color: 'var(--text-muted)' }}>ETA: </span>
                      <span className="mono" style={{ color: '#ffffff' }}>
                        {dl.etaSeconds > 0 ? `${dl.etaSeconds}s` : '—'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
