package telemetry

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"google.golang.org/protobuf/types/known/timestamppb"
	"runtime"
	"sync"
	"time"
)

type HostMetricsCollector struct {
	metrics []*HostMetrics
	mu      sync.Mutex
}

func NewHostMetricsCollector() *HostMetricsCollector {
	mc := new(HostMetricsCollector)
	mc.metrics = make([]*HostMetrics, 0)

	return mc
}

func (c *HostMetricsCollector) StartHostMetricsCollection(ctx context.Context, pollInterval time.Duration, measureInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				err := c.gatherMetrics(ctx, measureInterval)
				if err != nil {
					continue
				}
			}
		}
	}()
}

func (c *HostMetricsCollector) gatherMetrics(ctx context.Context, measureInterval time.Duration) error {
	cpuPercent, err := cpu.PercentWithContext(ctx, measureInterval, false)
	if err != nil {
		return err
	}

	memUsage, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}

	diskUsage, err := disk.UsageWithContext(ctx, c.getRootDir())
	if err != nil {
		return err
	}

	netUsageBefore, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return err
	}
	time.Sleep(measureInterval)
	netUsageAfter, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return err
	}

	metrics := HostMetrics{
		Timestamp: timestamppb.Now(),
		Cpu:       cpuPercent[0],
		Memory: &HostMetrics_Memory{
			Total: memUsage.Total,
			Used:  memUsage.Used,
		},
		Storage: &HostMetrics_Storage{
			Total: diskUsage.Total,
			Used:  diskUsage.Used,
		},
		Network: &HostMetrics_Network{
			UpSpeed:   float64(netUsageAfter[0].BytesSent-netUsageBefore[0].BytesSent) / measureInterval.Seconds(),
			DownSpeed: float64(netUsageAfter[0].BytesRecv-netUsageBefore[0].BytesRecv) / measureInterval.Seconds(),
		},
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, &metrics)

	return nil
}

func (c *HostMetricsCollector) GetHostMetrics() []*HostMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	var output []*HostMetrics
	copy(output, c.metrics)
	c.metrics = make([]*HostMetrics, 0)
	return output
}

func (c *HostMetricsCollector) getRootDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\"
	default:
		return "/"
	}
}
