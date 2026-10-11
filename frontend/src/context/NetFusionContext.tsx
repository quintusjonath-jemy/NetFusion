import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react'
import type {
  MPTCPStatus,
  NetworkEvent,
  NetworkInterface,
  Policy,
  SystemSettings,
  TrafficDistributionEntry,
  WSMessage,
} from '../types'
import { api } from '../services/api'

interface NetFusionContextType {
  interfaces: NetworkInterface[]
  trafficDistribution: TrafficDistributionEntry[]
  totalDownloadMbps: number
  totalUploadMbps: number
  activePathsCount: number
  mptcpStatus: MPTCPStatus | null
  activePolicy: Policy | null
  policies: Policy[]
  settings: SystemSettings | null
  events: NetworkEvent[]
  wsConnected: boolean
  simulationMode: boolean
  setSimulationMode: (enabled: boolean) => Promise<void>
  activatePolicy: (id: string) => Promise<void>
  triggerChaos: (action: 'disconnect' | 'degrade' | 'spike' | 'restore' | 'reset', iface?: string, latency?: number, loss?: number) => Promise<void>
  refreshData: () => Promise<void>
  notification: string | null
  clearNotification: () => void
}

const NetFusionContext = createContext<NetFusionContextType | undefined>(undefined)

export const NetFusionProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [interfaces, setInterfaces] = useState<NetworkInterface[]>([])
  const [trafficDistribution, setTrafficDistribution] = useState<TrafficDistributionEntry[]>([])
  const [totalDownloadMbps, setTotalDownloadMbps] = useState(0)
  const [totalUploadMbps, setTotalUploadMbps] = useState(0)
  const [activePathsCount, setActivePathsCount] = useState(0)
  const [mptcpStatus, setMptcpStatus] = useState<MPTCPStatus | null>(null)
  const [activePolicy, setActivePolicy] = useState<Policy | null>(null)
  const [policies, setPolicies] = useState<Policy[]>([])
  const [settings, setSettings] = useState<SystemSettings | null>(null)
  const [events, setEvents] = useState<NetworkEvent[]>([])
  const [wsConnected, setWsConnected] = useState(false)
  const [simulationMode, setSimulationModeState] = useState(true)
  const [notification, setNotification] = useState<string | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)

  const clearNotification = () => setNotification(null)

  const showNotification = (msg: string) => {
    setNotification(msg)
    setTimeout(() => {
      setNotification((curr) => (curr === msg ? null : curr))
    }, 5000)
  }

  // Initial Data Fetch
  const refreshData = useCallback(async () => {
    try {
      const [ifaces, traffic, pols, evts, settingsData] = await Promise.all([
        api.getNetworks().catch(() => []),
        api.getTraffic().catch(() => null),
        api.getPolicies().catch(() => []),
        api.getEvents(30).catch(() => []),
        api.getSettings().catch(() => null),
      ])

      if (ifaces) setInterfaces(ifaces)
      if (traffic) {
        setTrafficDistribution(traffic.distribution || [])
        setTotalDownloadMbps(traffic.totalDownloadMbps || 0)
        setTotalUploadMbps(traffic.totalUploadMbps || 0)
        setActivePathsCount(traffic.activePathsCount || 0)
      }
      if (pols) {
        setPolicies(pols)
        const active = pols.find((p) => p.isActive) || pols[0] || null
        setActivePolicy(active)
      }
      if (evts) setEvents(evts)
      if (settingsData) {
        setSettings(settingsData.settings)
        setMptcpStatus(settingsData.mptcp)
        setSimulationModeState(settingsData.settings.simulationMode)
        if (settingsData.activePolicy) setActivePolicy(settingsData.activePolicy)
      }
    } catch (err) {
      console.warn('Initial data load warning:', err)
    }
  }, [])

  // WebSocket Connection & Reconnection
  const connectWebSocket = useCallback(() => {
    if (wsRef.current && (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING)) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setWsConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        switch (msg.type) {
          case 'network_update':
            if (Array.isArray(msg.data)) {
              setInterfaces(msg.data as NetworkInterface[])
            }
            break
          case 'traffic_update':
            if (Array.isArray(msg.data)) {
              const dist = msg.data as TrafficDistributionEntry[]
              setTrafficDistribution(dist)
              let dl = 0
              let ul = 0
              dist.forEach((d) => {
                dl += d.downloadMbps
                ul += d.uploadMbps
              })
              setTotalDownloadMbps(parseFloat(dl.toFixed(2)))
              setTotalUploadMbps(parseFloat(ul.toFixed(2)))
              setActivePathsCount(dist.length)
            }
            break
          case 'network_event':
            {
              const newEvt = msg.data as NetworkEvent
              setEvents((prev) => [newEvt, ...prev.slice(0, 49)])
              showNotification(`[${newEvt.type.toUpperCase()}] ${newEvt.message}`)
            }
            break
          default:
            break
        }
      } catch (err) {
        console.error('Failed to parse WS frame:', err)
      }
    }

    ws.onclose = () => {
      setWsConnected(false)
      wsRef.current = null
      // Exponential backoff reconnect
      reconnectTimeoutRef.current = window.setTimeout(() => {
        connectWebSocket()
      }, 3000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    refreshData()
    connectWebSocket()

    // Fallback polling every 5s if WebSocket is disconnected
    const fallbackPoll = setInterval(() => {
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        refreshData()
      }
    }, 5000)

    return () => {
      clearInterval(fallbackPoll)
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current)
      if (wsRef.current) wsRef.current.close()
    }
  }, [connectWebSocket, refreshData])

  // Context Actions
  const setSimulationMode = async (enabled: boolean) => {
    try {
      await api.updateSettings({ simulationMode: enabled })
      setSimulationModeState(enabled)
      showNotification(`Simulation Mode ${enabled ? 'ENABLED' : 'DISABLED'}`)
      refreshData()
    } catch (err: unknown) {
      showNotification(`Failed to toggle simulation mode: ${err instanceof Error ? err.message : String(err)}`)
    }
  }

  const activatePolicy = async (id: string) => {
    try {
      await api.activatePolicy(id)
      const target = policies.find((p) => p.id === id) || null
      setActivePolicy(target)
      showNotification(`Policy '${target?.name || id}' activated!`)
      refreshData()
    } catch (err: unknown) {
      showNotification(`Error activating policy: ${err instanceof Error ? err.message : String(err)}`)
    }
  }

  const triggerChaos = async (
    action: 'disconnect' | 'degrade' | 'spike' | 'restore' | 'reset',
    iface?: string,
    latency?: number,
    loss?: number,
  ) => {
    try {
      await api.triggerChaos({
        action,
        interface: iface,
        latencyMs: latency,
        lossPct: loss,
      })
      showNotification(`Chaos action '${action}' triggered on ${iface || 'all interfaces'}`)
      refreshData()
    } catch (err: unknown) {
      showNotification(`Chaos injection failed: ${err instanceof Error ? err.message : String(err)}`)
    }
  }

  return (
    <NetFusionContext.Provider
      value={{
        interfaces,
        trafficDistribution,
        totalDownloadMbps,
        totalUploadMbps,
        activePathsCount,
        mptcpStatus,
        activePolicy,
        policies,
        settings,
        events,
        wsConnected,
        simulationMode,
        setSimulationMode,
        activatePolicy,
        triggerChaos,
        refreshData,
        notification,
        clearNotification,
      }}
    >
      {children}
    </NetFusionContext.Provider>
  )
}

export const useNetFusion = (): NetFusionContextType => {
  const context = useContext(NetFusionContext)
  if (!context) {
    throw new Error('useNetFusion must be used within a NetFusionProvider')
  }
  return context
}
