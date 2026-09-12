package dto

// ServerStatus is GET /monitor/server. Host paths stay generic (root only).
type ServerStatus struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemUsed     uint64  `json:"mem_used"`
	MemTotal    uint64  `json:"mem_total"`
	MemPercent  float64 `json:"mem_percent"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskPercent float64 `json:"disk_percent"`
	DiskPath    string  `json:"disk_path"`
	Load1       float64 `json:"load1"`
	Load5       float64 `json:"load5"`
	Load15      float64 `json:"load15"`
	GOOS        string  `json:"goos"`
	GOARCH      string  `json:"goarch"`
	CollectedAt string  `json:"collected_at"`
}

// AppHealth is GET /monitor/app (runtime / GC, no host credentials).
type AppHealth struct {
	StartedAt     string  `json:"started_at"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Version       string  `json:"version"`
	GoVersion     string  `json:"go_version"`
	Goroutines    int     `json:"goroutines"`
	MemAlloc      uint64  `json:"mem_alloc"`
	MemSys        uint64  `json:"mem_sys"`
	HeapAlloc     uint64  `json:"heap_alloc"`
	HeapSys       uint64  `json:"heap_sys"`
	HeapInuse     uint64  `json:"heap_inuse"`
	StackInuse    uint64  `json:"stack_inuse"`
	NumGC         uint32  `json:"num_gc"`
	NextGC        uint64  `json:"next_gc"`
	LastGCUnixMs  uint64  `json:"last_gc_unix_ms"`
	PauseTotalNs  uint64  `json:"pause_total_ns"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
	CollectedAt   string  `json:"collected_at"`
}

// DatabaseStatus is GET /monitor/database (sql.DB pool Stats only).
type DatabaseStatus struct {
	Available          bool   `json:"available"`
	OpenConnections    int    `json:"open_connections"`
	InUse              int    `json:"in_use"`
	Idle               int    `json:"idle"`
	WaitCount          int64  `json:"wait_count"`
	WaitDurationMs     int64  `json:"wait_duration_ms"`
	MaxOpenConnections int    `json:"max_open_connections"`
	MaxIdleClosed      int64  `json:"max_idle_closed"`
	MaxLifetimeClosed  int64  `json:"max_lifetime_closed"`
	CollectedAt        string `json:"collected_at"`
}

// RedisStatus is GET /monitor/redis (INFO clients/memory/stats/keyspace).
type RedisStatus struct {
	Available        bool    `json:"available"`
	ConnectedClients int64   `json:"connected_clients"`
	UsedMemory       int64   `json:"used_memory"`
	MaxMemory        int64   `json:"max_memory"`
	KeyspaceHits     int64   `json:"keyspace_hits"`
	KeyspaceMisses   int64   `json:"keyspace_misses"`
	HitRate          float64 `json:"hit_rate"`
	Keys             int64   `json:"keys"`
	Version          string  `json:"version"`
	CollectedAt      string  `json:"collected_at"`
}

// APIStats is GET /monitor/api-stats. Phase-1 reads Prometheus counters only.
// P50/P95/P99 persistence is deferred (scrape /metrics histograms later).
type APIStats struct {
	Available       bool     `json:"available"`
	Source          string   `json:"source"`
	RequestTotal    float64  `json:"request_total"`
	ErrorTotal      float64  `json:"error_total"`
	ErrorRate       float64  `json:"error_rate"`
	P50Seconds      *float64 `json:"p50_seconds"`
	P95Seconds      *float64 `json:"p95_seconds"`
	P99Seconds      *float64 `json:"p99_seconds"`
	PercentilesNote string   `json:"percentiles_note"`
	CollectedAt     string   `json:"collected_at"`
}
