package telemetry

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/steromano87/harkonnen/v1/pkg/message"
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

	payload := message.Envelope_HostMetrics{HostMetrics: &message.HostMetrics{
		Cpu: cpuPercent[0],
		Memory: &message.HostMetrics_Memory{
			Total: memUsage.Total,
			Used:  memUsage.Used,
		},
		Storage: &message.HostMetrics_Storage{
			Total: diskUsage.Total,
			Used:  diskUsage.Used,
		},
		Network: &message.HostMetrics_Network{
			UpSpeed:   float64(netUsageAfter[0].BytesSent-netUsageBefore[0].BytesSent) / measureInterval.Seconds(),
			DownSpeed: float64(netUsageAfter[0].BytesRecv-netUsageBefore[0].BytesRecv) / measureInterval.Seconds(),
		},
	}}

	msg := message.NewEnvelope(&payload)
	p.messenger.Send(msg)
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
