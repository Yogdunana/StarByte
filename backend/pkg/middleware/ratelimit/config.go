package ratelimit

// Bucket is one token-bucket: Rate tokens per second, Burst capacity.
type Bucket struct {
	Rate  float64
	Burst float64
}

// Config is IP / user / route token buckets plus ACL and gray settings.
type Config struct {
	IP    Bucket
	User  Bucket
	Route Bucket

	IPWhitelist   map[string]struct{}
	IPBlacklist   map[string]struct{}
	UserWhitelist map[string]struct{}
	UserBlacklist map[string]struct{}

	GrayHeader  string // default X-Gray-Group
	GrayAllow   map[string]struct{}
	GrayPercent int // 0 = off; 1-100 hash(user/ip) < percent is gray
}

func DefaultConfig() Config {
	return Config{
		IP:          Bucket{Rate: 100.0 / 60.0, Burst: 20},
		User:        Bucket{Rate: 2, Burst: 10},
		Route:       Bucket{Rate: 50, Burst: 50},
		GrayHeader:  "X-Gray-Group",
		GrayPercent: 0,
	}
}

func inSet(set map[string]struct{}, key string) bool {
	if key == "" || set == nil {
		return false
	}
	_, ok := set[key]
	return ok
}
