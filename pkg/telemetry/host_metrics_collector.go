package telemetry

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"google.golang.org/protobuf/types/known/timestamppb"
	"runtime"
	"time"
)

type HostMetricsCollector struct {
	pollInterval    time.Duration
	measureInterval time.Duration

	saver HostMetricsSaver
}

func NewHostMetricsCollector(pollInterval time.Duration, measureInterval time.Duration) *HostMetricsCollector {
	mc := new(HostMetricsCollector)
	mc.pollInterval = pollInterval
	mc.measureInterval = measureInterval

	return mc
}

func (c *HostMetricsCollector) Start(ctx context.Context, saver HostMetricsSaver) {
	c.saver = saver
	ticker := time.NewTicker(c.pollInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				err := c.gatherMetrics(ctx)
				if err != nil {
					continue
				}
			}
		}
	}()
}

func (c *HostMetricsCollector) gatherMetrics(ctx context.Context) error {
	cpuPercent, err := cpu.PercentWithContext(ctx, c.measureInterval, false)
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
	time.Sleep(c.measureInterval)
	netUsageAfter, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return err
	}

	sample := &HostMetricsSample{
		Timestamp: timestamppb.Now(),
		Cpu:       cpuPercent[0],
		Memory: &HostMetricsSample_Memory{
			Total: memUsage.Total,
			Used:  memUsage.Used,
		},
		Storage: &HostMetricsSample_Storage{
			Total: diskUsage.Total,
			Used:  diskUsage.Used,
		},
		Network: &HostMetricsSample_Network{
			UpSpeed:   float64(netUsageAfter[0].BytesSent-netUsageBefore[0].BytesSent) / c.measureInterval.Seconds(),
			DownSpeed: float64(netUsageAfter[0].BytesRecv-netUsageBefore[0].BytesRecv) / c.measureInterval.Seconds(),
		},
	}

	c.saver.SaveHostMetricsSample(sample)
	return nil
}

func (c *HostMetricsCollector) getRootDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\"
	default:
		return "/"
	}
}
