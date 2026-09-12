package service

import (
	"context"
	"runtime"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

const rootDiskPath = "/"

// HostSampler reads CPU / memory / disk / load. Tests inject a fake.
type HostSampler interface {
	CPUPercent(ctx context.Context) (float64, error)
	VirtualMemory() (used, total uint64, usedPercent float64, err error)
	DiskUsage(path string) (used, total uint64, usedPercent float64, err error)
	LoadAvg() (load1, load5, load15 float64, err error)
}

type gopsutilHost struct {
	cpuInterval time.Duration
}

func newGopsutilHost() HostSampler {
	return &gopsutilHost{cpuInterval: 150 * time.Millisecond}
}

func (h *gopsutilHost) CPUPercent(ctx context.Context) (float64, error) {
	vals, err := cpu.PercentWithContext(ctx, h.cpuInterval, false)
	if err != nil {
		return 0, err
	}
	if len(vals) == 0 {
		return 0, nil
	}
	return vals[0], nil
}

func (h *gopsutilHost) VirtualMemory() (uint64, uint64, float64, error) {
	st, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0, err
	}
	return st.Used, st.Total, st.UsedPercent, nil
}

func (h *gopsutilHost) DiskUsage(path string) (uint64, uint64, float64, error) {
	st, err := disk.Usage(path)
	if err != nil {
		return 0, 0, 0, err
	}
	return st.Used, st.Total, st.UsedPercent, nil
}

func (h *gopsutilHost) LoadAvg() (float64, float64, float64, error) {
	st, err := load.Avg()
	if err != nil {
		return 0, 0, 0, err
	}
	return st.Load1, st.Load5, st.Load15, nil
}

func collectServer(ctx context.Context, host HostSampler, now time.Time) (*dto.ServerStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	out := &dto.ServerStatus{
		DiskPath:    rootDiskPath,
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		CollectedAt: now.UTC().Format(time.RFC3339),
	}

	cpuPct, err := host.CPUPercent(ctx)
	if err != nil {
		return nil, collectFail("采集 CPU 失败")
	}
	out.CPUPercent = round2(cpuPct)

	used, total, pct, err := host.VirtualMemory()
	if err != nil {
		return nil, collectFail("采集内存失败")
	}
	out.MemUsed, out.MemTotal, out.MemPercent = used, total, round2(pct)

	dUsed, dTotal, dPct, err := host.DiskUsage(rootDiskPath)
	if err != nil {
		return nil, collectFail("采集磁盘失败")
	}
	out.DiskUsed, out.DiskTotal, out.DiskPercent = dUsed, dTotal, round2(dPct)

	l1, l5, l15, err := host.LoadAvg()
	if err != nil {
		return nil, collectFail("采集系统负载失败")
	}
	out.Load1, out.Load5, out.Load15 = round2(l1), round2(l5), round2(l15)
	return out, nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
