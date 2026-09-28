package models

import (
	"time"
)

// InterfaceStatus defines the health of a network interface
type InterfaceStatus string

const (
	StatusConnected    InterfaceStatus = "connected"
	StatusDisconnected InterfaceStatus = "disconnected"
	StatusDegraded     InterfaceStatus = "degraded"
	StatusRecovering   InterfaceStatus = "recovering"
)

// InterfaceType defines the physical/logical type of connection
type InterfaceType string

const (
	TypeEthernet InterfaceType = "ethernet"
	TypeWiFi     InterfaceType = "wifi"
	TypeCellular InterfaceType = "cellular"
	TypeOther    InterfaceType = "other"
)

// NetworkInterface represents a tracked physical or simulated network interface
type NetworkInterface struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"` // eth0, wlan0, wwan0, etc.
	Type                InterfaceType   `json:"type"`
	IPAddress           string          `json:"ipAddress"`
	MACAddress          string          `json:"macAddress"`
	Gateway             string          `json:"gateway"`
	Subnet              string          `json:"subnet"`
	Status              InterfaceStatus `json:"status"`
	Carrier             bool            `json:"carrier"`
	MTU                 int             `json:"mtu"`
	SignalStrength      int             `json:"signalStrength"` // percentage 0-100 (for wifi/cellular)
	RxBytes             int64           `json:"rxBytes"`
	TxBytes             int64           `json:"txBytes"`
	TotalDataUsedBytes  int64           `json:"totalDataUsedBytes"`
	CurrentDownloadMbps float64         `json:"currentDownloadMbps"`
	CurrentUploadMbps   float64         `json:"currentUploadMbps"`
	LatencyMs           float64         `json:"latencyMs"`
	PacketLoss          float64         `json:"packetLoss"` // 0.0 - 100.0 %
	JitterMs            float64         `json:"jitterMs"`
	StabilityScore      float64         `json:"stabilityScore"` // 0 - 100
	DynamicScore        float64         `json:"dynamicScore"`   // computed by Decision Engine
	AllocatedWeightPct  float64         `json:"allocatedWeightPct"`
	IsDefault           bool            `json:"isDefault"`
	IsSimulated         bool            `json:"isSimulated"`
	UptimeSeconds       int64           `json:"uptimeSeconds"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

// NetworkMetric holds a time-series metric snapshot for an interface
type NetworkMetric struct {
	ID            int64     `json:"id"`
	InterfaceID   string    `json:"interfaceId"`
	InterfaceName string    `json:"interfaceName"`
	DownloadMbps  float64   `json:"downloadMbps"`
	UploadMbps    float64   `json:"uploadMbps"`
	LatencyMs     float64   `json:"latencyMs"`
	PacketLoss    float64   `json:"packetLoss"`
	JitterMs      float64   `json:"jitterMs"`
	DynamicScore  float64   `json:"dynamicScore"`
	Timestamp     time.Time `json:"timestamp"`
}

// TrafficDistributionEntry captures the current allocated weight and traffic per interface
type TrafficDistributionEntry struct {
	InterfaceName string  `json:"interfaceName"`
	Type          string  `json:"type"`
	WeightPct     float64 `json:"weightPct"`
	DownloadMbps  float64 `json:"downloadMbps"`
	UploadMbps    float64 `json:"uploadMbps"`
	Status        string  `json:"status"`
}

// DownloadStatus defines the state of a download test
type DownloadStatus string

const (
	DownloadStatusQueued      DownloadStatus = "queued"
	DownloadStatusDownloading DownloadStatus = "downloading"
	DownloadStatusPaused      DownloadStatus = "paused"
	DownloadStatusCompleted   DownloadStatus = "completed"
	DownloadStatusFailed      DownloadStatus = "failed"
	DownloadStatusCancelled   DownloadStatus = "cancelled"
)

// InterfaceContribution tracks download contributions per interface
type InterfaceContribution struct {
	InterfaceName string  `json:"interfaceName"`
	BytesReceived int64   `json:"bytesReceived"`
	SpeedMbps     float64 `json:"speedMbps"`
	Percentage    float64 `json:"percentage"`
}

// DownloadSession represents a single or multi-interface download job
type DownloadSession struct {
	ID                     string                           `json:"id"`
	URL                    string                           `json:"url"`
	FileName               string                           `json:"fileName"`
	FileSize               int64                            `json:"fileSize"`
	DownloadedBytes        int64                            `json:"downloadedBytes"`
	Status                 DownloadStatus                   `json:"status"`
	Mode                   string                           `json:"mode"` // "single" or "multi"
	SelectedInterface      string                           `json:"selectedInterface,omitempty"`
	SpeedMbps              float64                          `json:"speedMbps"`
	ProgressPct            float64                          `json:"progressPct"`
	ETASeconds             int64                            `json:"etaSeconds"`
	ActivePaths            int                              `json:"activePaths"`
	InterfaceContributions map[string]InterfaceContribution `json:"interfaceContributions"`
	ErrorMessage           string                           `json:"errorMessage,omitempty"`
	CreatedAt              time.Time                        `json:"createdAt"`
	CompletedAt            *time.Time                       `json:"completedAt,omitempty"`
}

// EventSeverity represents log/alert severity
type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeveritySuccess  EventSeverity = "success"
	SeverityWarning  EventSeverity = "warning"
	SeverityCritical EventSeverity = "critical"
)

// NetworkEvent represents a lifecycle, failover, or alert event
type NetworkEvent struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"` // e.g. "failover", "disconnect", "recovered", "rebalance", "alert"
	InterfaceName string        `json:"interfaceName,omitempty"`
	Severity      EventSeverity `json:"severity"`
	Message       string        `json:"message"`
	Metadata      string        `json:"metadata,omitempty"` // JSON string
	CreatedAt     time.Time     `json:"createdAt"`
}

// PolicyMode defines the active traffic strategy
type PolicyMode string

const (
	PolicyPerformance PolicyMode = "performance"
	PolicyReliability PolicyMode = "reliability"
	PolicyBalanced    PolicyMode = "balanced"
	PolicyCostSaving  PolicyMode = "cost_saving"
)

// PolicyWeights configures the scoring components
type PolicyWeights struct {
	BandwidthWeight   float64 `json:"bandwidthWeight"`   // e.g. 0.40
	LatencyWeight     float64 `json:"latencyWeight"`     // e.g. 0.30
	StabilityWeight   float64 `json:"stabilityWeight"`   // e.g. 0.15
	ReliabilityWeight float64 `json:"reliabilityWeight"` // e.g. 0.15
	CostPenalty       float64 `json:"costPenalty"`       // e.g. 0.20
}

// Policy defines a traffic management strategy
type Policy struct {
	ID                         string             `json:"id"`
	Name                       string             `json:"name"`
	Mode                       PolicyMode         `json:"mode"`
	Weights                    PolicyWeights      `json:"weights"`
	InterfacePriorities        map[string]int     `json:"interfacePriorities"` // interface -> priority (1-10)
	DataLimitsMB               map[string]int64   `json:"dataLimitsMB"`        // interface -> monthly limit in MB
	FailoverEnabled            bool               `json:"failoverEnabled"`
	AutoRecoveryEnabled        bool               `json:"autoRecoveryEnabled"`
	DegradedThresholdLatencyMs float64            `json:"degradedThresholdLatencyMs"` // e.g. 150 ms
	DegradedThresholdLossPct   float64            `json:"degradedThresholdLossPct"`   // e.g. 5.0 %
	IsActive                   bool               `json:"isActive"`
	UpdatedAt                  time.Time          `json:"updatedAt"`
}

// BenchmarkResult represents head-to-head comparison
type BenchmarkResult struct {
	ID                     string    `json:"id"`
	TestName               string    `json:"testName"`
	URL                    string    `json:"url"`
	DurationSeconds        int       `json:"durationSeconds"`
	SingleInterfaceName    string    `json:"singleInterfaceName"`
	SingleMbps             float64   `json:"singleMbps"`
	MultiMbps              float64   `json:"multiMbps"`
	ImprovementPct         float64   `json:"improvementPct"`
	SingleLatencyAvg       float64   `json:"singleLatencyAvg"`
	MultiLatencyAvg        float64   `json:"multiLatencyAvg"`
	SinglePacketLoss       float64   `json:"singlePacketLoss"`
	MultiPacketLoss        float64   `json:"multiPacketLoss"`
	FailoverRecoveryTimeMs int64     `json:"failoverRecoveryTimeMs"`
	CreatedAt              time.Time `json:"createdAt"`
}

// SystemSettings encapsulates operational parameters
type SystemSettings struct {
	SimulationMode    bool      `json:"simulationMode"`
	TelemetryInterval int       `json:"telemetryIntervalMs"` // in ms
	RetentionDays     int       `json:"retentionDays"`
	LogLevel          string    `json:"logLevel"`
	MPTCPEnabled      bool      `json:"mptcpEnabled"`
	ActivePolicyID    string    `json:"activePolicyId"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
