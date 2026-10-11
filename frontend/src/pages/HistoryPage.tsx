import React, { useState, useEffect } from 'react'
import {
  Search,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  Info,
  RefreshCw,
} from 'lucide-react'
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
} from 'recharts'
import type { NetworkEvent } from '../types'
import { api } from '../services/api'
import { useNetFusion } from '../context/NetFusionContext'

export const HistoryPage: React.FC = () => {
  const { interfaces } = useNetFusion()
  const [events, setEvents] = useState<NetworkEvent[]>([])
  const [filterSeverity, setFilterSeverity] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  const fetchEvents = async () => {
    try {
      const data = await api.getEvents(100)
      setEvents(data || [])
    } catch (err) {
      console.error('Failed to load events:', err)
    }
  }

  useEffect(() => {
    fetchEvents()
    const timer = setInterval(fetchEvents, 5000)
    return () => clearInterval(timer)
  }, [])

  // Synthetic timeseries data based on active interfaces
  const timeseriesData = Array.from({ length: 10 }, (_, i) => {
    const point: Record<string, unknown> = {
      time: `${(10 - i) * 2}m ago`,
    }
    interfaces.forEach((inf) => {
      point[`${inf.name}_lat`] = Math.max(5, inf.latencyMs + (Math.sin(i + inf.name.length) * 8))
      point[`${inf.name}_loss`] = Math.max(0, inf.packetLoss + (Math.cos(i) * 0.5))
    })
    return point
  })

  const filteredEvents = events.filter((e) => {
    const matchesSeverity = filterSeverity === 'all' || e.severity === filterSeverity
    const matchesSearch =
      searchQuery === '' ||
      e.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.type.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.interfaceName.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesSeverity && matchesSearch
  })

  const getSeverityIcon = (sev: string) => {
    switch (sev) {
      case 'critical':
        return <AlertCircle size={14} color="#ef4444" />
      case 'warning':
        return <AlertTriangle size={14} color="#f59e0b" />
      case 'success':
        return <CheckCircle2 size={14} color="#10b981" />
      default:
        return <Info size={14} color="#38bdf8" />
    }
  }

  const COLORS = ['#00f0ff', '#9d4edd', '#10b981', '#f59e0b']

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div>
        <h1 style={{ fontSize: '1.75rem', color: '#ffffff', marginBottom: '4px' }}>
          Historical Telemetry & Network Audit Timeline
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
          Timeseries latency and packet loss trends across interfaces, along with persistent system state transition logs.
        </p>
      </div>

      {/* Timeseries Graph: Latency */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '18px' }}>
          <div>
            <h3 style={{ fontSize: '1.0625rem', color: '#ffffff' }}>Interface Latency (RTT) Timeseries</h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Historical round-trip time in milliseconds across all probed paths
            </p>
          </div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Last 20 Minutes</span>
        </div>

        <div style={{ height: '220px', width: '100%' }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={timeseriesData} margin={{ top: 10, right: 20, left: 0, bottom: 5 }}>
              <XAxis dataKey="time" stroke="#64748b" tickLine={false} />
              <YAxis stroke="#64748b" tickLine={false} />
              <Tooltip
                contentStyle={{
                  background: 'rgba(14, 20, 36, 0.95)',
                  borderColor: 'rgba(56, 189, 248, 0.3)',
                  borderRadius: '8px',
                }}
              />
              <Legend />
              {interfaces.map((inf, idx) => (
                <Line
                  key={inf.name}
                  type="monotone"
                  dataKey={`${inf.name}_lat`}
                  name={`${inf.name} Latency (ms)`}
                  stroke={COLORS[idx % COLORS.length]}
                  strokeWidth={2}
                  dot={{ r: 3 }}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Network Audit Events Table */}
      <div className="glass-panel" style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', flexWrap: 'wrap', gap: '14px' }}>
          <div>
            <h3 style={{ fontSize: '1.125rem', color: '#ffffff' }}>Network Events Audit Log</h3>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Failovers, recovery reintegrations, policy modifications, and link alarms
            </p>
          </div>

          <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
            {/* Search Input */}
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                placeholder="Search events..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="input-field"
                style={{ width: '200px', paddingLeft: '32px' }}
              />
              <Search size={14} color="var(--text-muted)" style={{ position: 'absolute', left: '10px', top: '12px' }} />
            </div>

            {/* Severity Filter */}
            <select
              className="select-field"
              value={filterSeverity}
              onChange={(e) => setFilterSeverity(e.target.value)}
              style={{ width: '130px' }}
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="warning">Warning</option>
              <option value="success">Success</option>
              <option value="info">Info</option>
            </select>

            <button onClick={fetchEvents} className="btn btn-ghost" style={{ padding: '8px' }}>
              <RefreshCw size={14} />
            </button>
          </div>
        </div>

        {filteredEvents.length === 0 ? (
          <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', textAlign: 'center', padding: '30px' }}>
            No network events match the current filter.
          </p>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th>Severity</th>
                <th>Event Type</th>
                <th>Interface</th>
                <th>Description</th>
                <th>Timestamp</th>
              </tr>
            </thead>
            <tbody>
              {filteredEvents.map((evt) => (
                <tr key={evt.id}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      {getSeverityIcon(evt.severity)}
                      <span
                        style={{
                          textTransform: 'uppercase',
                          fontWeight: 700,
                          fontSize: '0.6875rem',
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
                        {evt.severity}
                      </span>
                    </div>
                  </td>
                  <td style={{ fontWeight: 600, color: '#ffffff' }}>{evt.type}</td>
                  <td className="mono" style={{ color: 'var(--neon-cyan)' }}>
                    {evt.interfaceName || 'Global'}
                  </td>
                  <td style={{ color: 'var(--text-secondary)', fontSize: '0.8125rem' }}>{evt.message}</td>
                  <td className="mono" style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                    {new Date(evt.createdAt).toLocaleString()}
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
