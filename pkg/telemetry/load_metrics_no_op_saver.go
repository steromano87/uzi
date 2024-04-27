package telemetry

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
)

type LoadMetricsNoOpSaver struct {
	logger zerolog.Logger
}

func NewLoadMetricsNoOpSaver(logger zerolog.Logger) LoadMetricsNoOpSaver {
	noOpLogger := logger.With().Str(log.ComponentKey, "NoOp Load Metrics Saver").Logger()
	return LoadMetricsNoOpSaver{
		logger: noOpLogger,
	}
}

func (l LoadMetricsNoOpSaver) SaveSample(sample *Sample) {
	l.logger.Trace().Str("sample", sample.String()).Msg("Sending sample to sink")
}

func (l LoadMetricsNoOpSaver) SaveTransaction(transaction *Transaction) {
	l.logger.Trace().Str("transaction", transaction.String()).Msg("Sending transaction to sink")
}
