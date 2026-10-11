import type {
  BenchmarkResult,
  DownloadSession,
  MPTCPStatus,
  NetworkEvent,
  NetworkInterface,
  Policy,
  SystemSettings,
  TrafficSummary,
} from '../types'

const API_BASE = '/api'

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })

  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(errorBody.error || `HTTP ${res.status}`)
  }

  return res.json()
}

export const api = {
  // Health
  async getHealth(): Promise<{
    status: string
    component: string
    version: string
    simulationMode: boolean
    interfacesCount: number
    mptcp: MPTCPStatus
    activePolicy: string
  }> {
    const res = await fetch('/healthz')
    return res.json()
  },

  // Networks & Telemetry
  async getNetworks(): Promise<NetworkInterface[]> {
    return request<NetworkInterface[]>('/networks')
  },

  async getNetworkByID(id: string): Promise<{ interface: NetworkInterface; metrics: unknown[] }> {
    return request(`/networks/${id}`)
  },

  async getMetrics(iface?: string, minutes = 15): Promise<Record<string, unknown[]>> {
    const query = new URLSearchParams()
    if (iface) query.set('interface', iface)
    query.set('minutes', minutes.toString())
    return request(`/metrics?${query.toString()}`)
  },

  // Traffic
  async getTraffic(): Promise<TrafficSummary> {
    return request<TrafficSummary>('/traffic')
  },

  // Downloader
  async getDownloads(): Promise<DownloadSession[]> {
    return request<DownloadSession[]>('/downloads')
  },

  async createDownload(url: string, mode: 'single' | 'multi', selectedInterface?: string): Promise<DownloadSession> {
    return request<DownloadSession>('/downloads', {
      method: 'POST',
      body: JSON.stringify({ url, mode, selectedInterface }),
    })
  },

  async stopDownload(id: string): Promise<{ status: string }> {
    return request(`/downloads/${id}/stop`, {
      method: 'POST',
    })
  },

  // Benchmarks
  async getBenchmarks(): Promise<BenchmarkResult[]> {
    return request<BenchmarkResult[]>('/benchmarks')
  },

  async runBenchmark(testName: string, url: string, durationSeconds = 10): Promise<BenchmarkResult> {
    return request<BenchmarkResult>('/benchmarks', {
      method: 'POST',
      body: JSON.stringify({ testName, url, durationSeconds }),
    })
  },

  // Policies
  async getPolicies(): Promise<Policy[]> {
    return request<Policy[]>('/policies')
  },

  async updatePolicy(policy: Policy): Promise<Policy> {
    return request<Policy>('/policies', {
      method: 'PUT',
      body: JSON.stringify(policy),
    })
  },

  async activatePolicy(id: string): Promise<{ status: string; policyId: string }> {
    return request(`/policies/${id}/activate`, {
      method: 'POST',
    })
  },

  // Events & Settings
  async getEvents(limit = 50): Promise<NetworkEvent[]> {
    return request<NetworkEvent[]>(`/events?limit=${limit}`)
  },

  async getSettings(): Promise<{ settings: SystemSettings; mptcp: MPTCPStatus; activePolicy: Policy }> {
    return request('/settings')
  },

  async updateSettings(settings: Partial<SystemSettings>): Promise<SystemSettings> {
    return request<SystemSettings>('/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    })
  },

  // Simulation Chaos
  async triggerChaos(payload: {
    action: 'disconnect' | 'degrade' | 'spike' | 'restore' | 'reset'
    interface?: string
    latencyMs?: number
    lossPct?: number
    durationSec?: number
  }): Promise<{ status: string; action: string }> {
    return request('/simulation/chaos', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
