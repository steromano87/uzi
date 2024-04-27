package telemetry

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
)

type HostMetricsNoOpSaver struct {
	logger zerolog.Logger
}

func NewHostMetricsNoOpSaver(logger zerolog.Logger) HostMetricsNoOpSaver {
	noOpLogger := logger.With().Str(log.ComponentKey, "NoOp Host Metrics Saver").Logger()
	return HostMetricsNoOpSaver{
		logger: noOpLogger,
	}
}

func (h HostMetricsNoOpSaver) SaveHostMetricsSample(hostMetrics *HostMetricsSample) {
	h.logger.Trace().Str("sample", hostMetrics.String()).Msg("Sending sample to sink")
}
