import React, { useState, useEffect } from 'react'
import {
  Play,
  TrendingUp,
  RefreshCw,
} from 'lucide-react'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
} from 'recharts'
import type { BenchmarkResult } from '../types'
import { api } from '../services/api'

export const BenchmarkPage: React.FC = () => {
  const [benchmarks, setBenchmarks] = useState<BenchmarkResult[]>([])
  const [latestResult, setLatestResult] = useState<BenchmarkResult | null>(null)
  const [testName, setTestName] = useState('Dual-Path Stress Benchmark')
  const [targetUrl, setTargetUrl] = useState('https://speed.cloudflare.com/__down?bytes=50000000')
  const [duration, setDuration] = useState(10)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchBenchmarks = async () => {
    try {
      const data = await api.getBenchmarks()
      if (data && data.length > 0) {
        setBenchmarks(data)
        if (!latestResult) {
          setLatestResult(data[0])
        }
      }
    } catch (err) {
      console.error('Failed to load benchmarks:', err)
    }
  }

  useEffect(() => {
    fetchBenchmarks()
  }, [])

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const res = await api.runBenchmark(testName, targetUrl, duration)
      setLatestResult(res)
      await fetchBenchmarks()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  // Data for comparative bar chart
  const barData = latestResult
    ? [
        {
          metric: 'Throughput (Mbps)',
          'Single Interface': latestResult.singleMbps,
          'Multi-Network Aggregated': latestResult.multiMbps,
        },
        {
          metric: 'Latency RTT (ms)',
          'Single Interface': latestResult.singleLatencyAvg,
          'Multi-Network Aggregated': latestResult.multiLatencyAvg,
        },
      ]
    : []

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Comparative Head-to-Head Performance Benchmark
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Evaluate throughput, latency distribution, packet loss, and recovery time of Single Network vs Multi-Network Aggregation.
        </p>
      </div>

      {/* Trigger Card */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <h3 style={{ fontSize: '1.125rem', color: '#ffffff', marginBottom: '16px' }}>
          Execute Real-Time Comparative Test
        </h3>

        <form onSubmit={handleRun} style={{ display: 'grid', gridTemplateColumns: '1.5fr 2fr 1fr auto', gap: '14px', alignItems: 'flex-end' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
              Test Identifier:
            </label>
            <input
              type="text"
              className="input-field"
              value={testName}
              onChange={(e) => setTestName(e.target.value)}
              required
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
              Benchmark Target URL:
            </label>
            <input
              type="url"
              className="input-field"
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              required
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '6px' }}>
              Duration (Seconds):
            </label>
            <select
              className="select-field"
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value))}
            >
              <option value={5}>5 Seconds (Quick)</option>
              <option value={10}>10 Seconds (Standard)</option>
              <option value={20}>20 Seconds (Thorough)</option>
            </select>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn btn-cyan"
            style={{ padding: '10px 24px', height: '42px' }}
          >
            <Play size={16} fill="currentColor" />
            <span>{loading ? 'Testing...' : 'Run Test'}</span>
          </button>
        </form>

        {error && (
          <p style={{ color: 'var(--neon-crimson)', fontSize: '0.8125rem', marginTop: '12px' }}>
            {error}
          </p>
        )}
      </div>

      {/* Latest Result Comparison Showcase */}
      {latestResult && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          {/* Dual Big Comparative HUD Cards */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '20px' }}>
            {/* Single Interface */}
            <div
              className="glass-panel"
              style={{
                padding: '24px',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                background: 'rgba(12, 17, 32, 0.8)',
              }}
            >
              <span
                style={{
                  fontSize: '0.75rem',
                  color: 'var(--text-muted)',
                  textTransform: 'uppercase',
                  fontWeight: 700,
                  display: 'block',
                  marginBottom: '8px',
                }}
              >
                SINGLE INTERFACE ({latestResult.singleInterfaceName})
              </span>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px', marginBottom: '16px' }}>
                <span className="mono" style={{ fontSize: '2.5rem', fontWeight: 800, color: '#ffffff' }}>
                  {latestResult.singleMbps.toFixed(1)}
                </span>
                <span style={{ fontSize: '1rem', color: 'var(--text-muted)' }}>Mbps</span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '0.8125rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-muted)' }}>Average Latency:</span>
                  <span className="mono" style={{ color: '#ffffff' }}>
                    {latestResult.singleLatencyAvg.toFixed(1)} ms
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-muted)' }}>Packet Loss:</span>
                  <span className="mono" style={{ color: '#ffffff' }}>
                    {latestResult.singlePacketLoss.toFixed(1)}%
                  </span>
                </div>
              </div>
            </div>

            {/* Multi-Network Aggregated */}
            <div
              className="glass-panel-glow"
              style={{
                padding: '24px',
                background: 'rgba(14, 22, 44, 0.95)',
                position: 'relative',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  position: 'absolute',
                  top: '-15px',
                  right: '-15px',
                  width: '80px',
                  height: '80px',
                  background: 'radial-gradient(circle, rgba(0, 240, 255, 0.4) 0%, transparent 70%)',
                }}
              />
              <span
                style={{
                  fontSize: '0.75rem',
                  color: 'var(--neon-cyan)',
                  textTransform: 'uppercase',
                  fontWeight: 700,
                  display: 'block',
                  marginBottom: '8px',
                }}
              >
                MULTI-NETWORK AGGREGATED (ALL PATHS)
              </span>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px', marginBottom: '16px' }}>
                <span className="mono" style={{ fontSize: '2.5rem', fontWeight: 800, color: 'var(--neon-cyan)' }}>
                  {latestResult.multiMbps.toFixed(1)}
                </span>
                <span style={{ fontSize: '1rem', color: 'var(--neon-cyan)' }}>Mbps</span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '0.8125rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-muted)' }}>Blended Latency:</span>
                  <span className="mono" style={{ color: '#ffffff' }}>
                    {latestResult.multiLatencyAvg.toFixed(1)} ms
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: 'var(--text-muted)' }}>Packet Loss:</span>
                  <span className="mono" style={{ color: 'var(--neon-emerald)' }}>
                    {latestResult.multiPacketLoss.toFixed(1)}%
                  </span>
                </div>
              </div>
            </div>

            {/* Speedup Improvement Hero Card */}
            <div
              className="glass-panel"
              style={{
                padding: '24px',
                background: 'linear-gradient(135deg, rgba(0, 240, 255, 0.1), rgba(112, 0, 255, 0.15))',
                border: '1px solid rgba(0, 240, 255, 0.3)',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                textAlign: 'center',
              }}
            >
              <TrendingUp size={36} color="var(--neon-emerald)" style={{ marginBottom: '10px' }} />
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase', fontWeight: 700 }}>
                THROUGHPUT IMPROVEMENT
              </span>
              <span className="mono" style={{ fontSize: '2.75rem', fontWeight: 900, color: 'var(--neon-emerald)' }}>
                +{latestResult.improvementPct.toFixed(1)}%
              </span>
              <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '4px' }}>
                Failover Recovery Time: <strong style={{ color: '#ffffff' }}>{latestResult.failoverRecoveryTimeMs} ms</strong>
              </p>
            </div>
          </div>

          {/* Bar Chart Comparison */}
          <div className="glass-panel" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.0625rem', color: '#ffffff', marginBottom: '18px' }}>
              Comparative Performance Metrics
            </h3>
            <div style={{ height: '240px', width: '100%' }}>
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={barData} margin={{ top: 10, right: 20, left: 0, bottom: 5 }}>
                  <XAxis dataKey="metric" stroke="#64748b" tickLine={false} />
                  <YAxis stroke="#64748b" tickLine={false} />
                  <Tooltip
                    contentStyle={{
                      background: 'rgba(14, 20, 36, 0.95)',
                      borderColor: 'rgba(56, 189, 248, 0.3)',
                      borderRadius: '8px',
                    }}
                  />
                  <Legend />
                  <Bar dataKey="Single Interface" fill="#64748b" radius={[6, 6, 0, 0]} />
                  <Bar dataKey="Multi-Network Aggregated" fill="#00f0ff" radius={[6, 6, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      )}

      {/* Historical Benchmarks Table */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Benchmark Execution History</h3>
          <button onClick={fetchBenchmarks} className="btn btn-ghost" style={{ padding: '6px', fontSize: '0.75rem' }}>
            <RefreshCw size={14} /> Refresh
          </button>
        </div>

        {benchmarks.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', textAlign: 'center', padding: '20px' }}>
            No historical benchmarks found. Run a test above to record comparative analytics.
          </p>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th>Test Name</th>
                <th>Single Mbps</th>
                <th>Multi Mbps</th>
                <th>Improvement</th>
                <th>Single Latency</th>
                <th>Multi Latency</th>
                <th>Failover MS</th>
                <th>Executed</th>
              </tr>
            </thead>
            <tbody>
              {benchmarks.map((bm) => (
                <tr key={bm.id}>
                  <td style={{ fontWeight: 600, color: '#ffffff' }}>{bm.testName}</td>
                  <td className="mono">{bm.singleMbps.toFixed(1)} Mbps</td>
                  <td className="mono" style={{ color: 'var(--neon-cyan)', fontWeight: 600 }}>
                    {bm.multiMbps.toFixed(1)} Mbps
                  </td>
                  <td>
                    <span
                      style={{
                        color: 'var(--neon-emerald)',
                        fontWeight: 700,
                        fontSize: '0.875rem',
                      }}
                    >
                      +{bm.improvementPct.toFixed(1)}%
                    </span>
                  </td>
                  <td className="mono">{bm.singleLatencyAvg.toFixed(1)} ms</td>
                  <td className="mono">{bm.multiLatencyAvg.toFixed(1)} ms</td>
                  <td className="mono">{bm.failoverRecoveryTimeMs} ms</td>
                  <td className="mono" style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                    {new Date(bm.createdAt).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
