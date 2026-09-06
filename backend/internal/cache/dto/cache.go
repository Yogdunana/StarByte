package dto

import "github.com/Yogdunana/StarByte/backend/pkg/cache"

// KeyInfo is one Redis key shown in the cache admin list.
type KeyInfo struct {
	Key        string `json:"key"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

// Stats is GET /system/cache/stats.
type Stats struct {
	Healthy  bool             `json:"healthy"`
	Pool     cache.PoolHealth `json:"pool" swaggertype:"object"`
	L1Hits   int64            `json:"l1_hits"`
	L1Misses int64            `json:"l1_misses"`
	L1Size   int              `json:"l1_size"`
	Keys     []KeyInfo        `json:"keys"`
	KeyCount int              `json:"key_count"`
	Pattern  string           `json:"pattern"`
}

// DeleteResult is returned by key/pattern delete.
type DeleteResult struct {
	Deleted int64    `json:"deleted"`
	Keys    []string `json:"keys"`
}

// WarmupEntry is one key written during warmup.
type WarmupEntry struct {
	Key        string `json:"key" binding:"required"`
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttl_seconds"`
}

// WarmupRequest is POST /system/cache/warmup.
type WarmupRequest struct {
	Entries    []WarmupEntry `json:"entries"`
	ScanPrefix string        `json:"scan_prefix"`
}

// WarmupResult reports how many keys were loaded into L1/L2.
type WarmupResult struct {
	Loaded int      `json:"loaded"`
	Keys   []string `json:"keys"`
}
