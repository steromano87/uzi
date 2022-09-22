package telemetry

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"runtime"
	"time"
)

type HostMetricsSender struct {
	ctx       context.Context
	messenger messaging.Messenger
}

func NewHostMetricsSender(messenger messaging.Messenger) *HostMetricsSender {
	mc := new(HostMetricsSender)
	mc.messenger = messenger

	return mc
}

func (p HostMetricsSender) Start(ctx context.Context, pollInterval time.Duration, measureInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				err := p.gatherMetrics(ctx, measureInterval)
				if err != nil {
					continue
				}
			}
		}
	}()
}

func (p HostMetricsSender) gatherMetrics(ctx context.Context, measureInterval time.Duration) error {
	cpuPercent, err := cpu.PercentWithContext(ctx, measureInterval, false)
	if err != nil {
		return err
	}

	memUsage, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}

	diskUsage, err := disk.UsageWithContext(ctx, p.getRootDir())
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

	messagePayload := messaging.HostMetricsPayload{
		CPU: cpuPercent[0],
		Memory: struct {
			Total uint64
			Used  uint64
		}{
			Total: memUsage.Total,
			Used:  memUsage.Used,
		},
		Disk: struct {
			Total uint64
			Used  uint64
		}{
			Total: diskUsage.Total,
			Used:  diskUsage.Used,
		},
		Network: struct {
			UpSpeed   float64
			DownSpeed float64
		}{
			UpSpeed:   float64(netUsageAfter[0].BytesSent-netUsageBefore[0].BytesSent) / measureInterval.Seconds(),
			DownSpeed: float64(netUsageAfter[0].BytesRecv-netUsageBefore[0].BytesRecv) / measureInterval.Seconds(),
		},
	}

	p.messenger.Send(messaging.NewRawMessage(messaging.HostMetricsMsgId, &messagePayload))
	return nil
}

func (p HostMetricsSender) getRootDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\"
	default:
		return "/"
	}
}
