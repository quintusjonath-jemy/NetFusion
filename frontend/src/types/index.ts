export type InterfaceType = 'ethernet' | 'wifi' | 'cellular' | 'loopback' | 'unknown'
export type InterfaceStatus = 'connected' | 'degraded' | 'disconnected' | 'recovering'
export type PolicyMode = 'performance' | 'reliability' | 'balanced' | 'cost_saving'
export type EventSeverity = 'info' | 'warning' | 'critical' | 'success'

export interface NetworkInterface {
  id: string
  name: string
  type: InterfaceType
  ipAddress: string
  macAddress: string
  gateway: string
  subnet: string
  status: InterfaceStatus
  carrier: boolean
  mtu: number
  signalStrength: number
  rxBytes: number
  txBytes: number
  totalDataUsedBytes: number
  currentDownloadMbps: number
  currentUploadMbps: number
  latencyMs: number
  packetLoss: number
  jitterMs: number
  stabilityScore: number
  dynamicScore: number
  allocatedWeightPct: number
  isDefault: boolean
  isSimulated: boolean
  uptimeSeconds: number
  updatedAt: string
}

export interface TrafficDistributionEntry {
  interfaceName: string
  type: InterfaceType
  weightPct: number
  downloadMbps: number
  uploadMbps: number
  status: InterfaceStatus
}

export interface TrafficSummary {
  distribution: TrafficDistributionEntry[]
  totalDownloadMbps: number
  totalUploadMbps: number
  activePathsCount: number
  mptcpEnabled: boolean
  simulationMode: boolean
}

export interface PolicyWeights {
  bandwidthWeight: number
  latencyWeight: number
  stabilityWeight: number
  reliabilityWeight: number
  costPenalty: number
}

export interface Policy {
  id: string
  name: string
  mode: PolicyMode
  weights: PolicyWeights
  interfacePriorities: Record<string, number>
  dataLimitsMB: Record<string, number>
  failoverEnabled: boolean
  autoRecoveryEnabled: boolean
  degradedThresholdLatencyMs: number
  degradedThresholdLossPct: number
  isActive: boolean
  updatedAt?: string
}

export interface InterfaceContribution {
  bytes: number
  mbps: number
}

export interface DownloadSession {
  id: string
  url: string
  fileName: string
  fileSize: number
  downloadedBytes: number
  status: 'queued' | 'downloading' | 'completed' | 'failed' | 'cancelled'
  mode: 'single' | 'multi'
  selectedInterface?: string
  speedMbps: number
  progressPct: number
  etaSeconds: number
  activePaths: number
  interfaceContributions: Record<string, InterfaceContribution>
  errorMessage?: string
  createdAt: string
  completedAt?: string
}

export interface BenchmarkResult {
  id: string
  testName: string
  url: string
  durationSeconds: number
  singleInterfaceName: string
  singleMbps: number
  multiMbps: number
  improvementPct: number
  singleLatencyAvg: number
  multiLatencyAvg: number
  singlePacketLoss: number
  multiPacketLoss: number
  failoverRecoveryTimeMs: number
  createdAt: string
}

export interface NetworkEvent {
  id: string
  type: string
  interfaceName: string
  severity: EventSeverity
  message: string
  metadata?: Record<string, unknown>
  createdAt: string
}

export interface MPTCPEndpoint {
  id: number
  ip: string
  interface: string
  flags: string[]
}

export interface MPTCPStatus {
  supported: boolean
  enabled: boolean
  pathManager: string
  scheduler: string
  endpoints: MPTCPEndpoint[]
  errorMessage?: string
}

export interface SystemSettings {
  id: number
  simulationMode: boolean
  telemetryIntervalMs: number
  retentionDays: number
  logLevel: string
  mptcpEnabled: boolean
  activePolicyId: string
  updatedAt?: string
}

export interface WSMessage<T = unknown> {
  type: string
  timestamp: string
  data: T
}
