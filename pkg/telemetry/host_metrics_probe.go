package telemetry

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"runtime"
	"time"
)

type HostMetricsProbe struct {
	pollInterval    time.Duration
	measureInterval time.Duration

	storer HostMetricsStorer
	logger *zerolog.Logger
}

func NewHostMetricsProbe(storer HostMetricsStorer) *HostMetricsProbe {
	mc := new(HostMetricsProbe)
	mc.storer = storer

	return mc
}

func (p *HostMetricsProbe) Serve(ctx context.Context, pollInterval time.Duration, measureInterval time.Duration) {
	p.pollInterval = pollInterval
	p.measureInterval = measureInterval
	p.setLogger(zerolog.Ctx(ctx))
	p.logger.Info().Dur(
		"pollInterval", p.pollInterval,
	).Dur("measureInterval", p.measureInterval).Msg("Host metrics probe started")

	ticker := time.NewTicker(p.pollInterval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			p.logger.Info().Msg("Host metrics probe stopped")
			return

		case <-ticker.C:
			p.logger.Trace().Msg("Gather metrics loop started")
			err := p.gatherMetrics(ctx)
			if err != nil {
				p.logger.Warn().AnErr("metricsGatheringError", err).Msg("Encountered an error while gathering metrics, continuing...")
				continue
			}
			p.logger.Trace().Msg("Gather metrics loop ended")
		}
	}
}

func (p *HostMetricsProbe) setLogger(logger *zerolog.Logger) {
	mainLogger := logger.With().Str(log.ComponentKey, "Host metrics probe").Logger()
	p.logger = &mainLogger
}

func (p *HostMetricsProbe) gatherMetrics(ctx context.Context) error {
	cpuPercent, err := cpu.PercentWithContext(ctx, p.measureInterval, false)
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
	time.Sleep(p.measureInterval)
	netUsageAfter, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return err
	}

	sample := &HostMetrics{
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
			UpSpeed:   float64(netUsageAfter[0].BytesSent-netUsageBefore[0].BytesSent) / p.measureInterval.Seconds(),
			DownSpeed: float64(netUsageAfter[0].BytesRecv-netUsageBefore[0].BytesRecv) / p.measureInterval.Seconds(),
		},
	}

	return p.storer.StoreHostMetrics(sample)
}

func (p *HostMetricsProbe) getRootDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\"
	default:
		return "/"
	}
}
