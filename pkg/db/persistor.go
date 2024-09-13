package db

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"io"
)

type Persistor struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewPersistor(db *gorm.DB) *Persistor {
	persistor := new(Persistor)
	persistor.db = db
	return persistor
}

func (p *Persistor) Serve(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient, logsClient telemetry.LogsClient) error {
	p.setLogger(zerolog.Ctx(ctx))

	clientErrGroup, _ := errgroup.WithContext(ctx)
	clientErrGroup.Go(func() error {
		return p.readSamples(ctx, agentId, metricsClient)
	})
	clientErrGroup.Go(func() error {
		return p.readTransactions(ctx, agentId, metricsClient)
	})
	clientErrGroup.Go(func() error {
		return p.readIterationCounters(ctx, agentId, metricsClient)
	})
	clientErrGroup.Go(func() error {
		return p.readHostMetrics(ctx, agentId, metricsClient)
	})
	clientErrGroup.Go(func() error {
		return p.readRawLogs(ctx, agentId, logsClient)
	})

	return clientErrGroup.Wait()
}

func (p *Persistor) setLogger(logger *zerolog.Logger) {
	p.logger = logger.With().Str(log.ComponentKey, "DB Persistor").Logger()
}

func (p *Persistor) readSamples(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	serverStream, err := metricsClient.GetSamples(ctx, &telemetry.SampleStreamRequest{})
	if err != nil {
		return err
	}

	for {
		sample, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		dbSample := NewSampleFromGrpc(sample)
		dbSample.AgentId = agentId
		p.logger.Trace().Str("name", dbSample.Name).Msg("Persisting sample")
		if result := p.db.Save(&dbSample); result.Error != nil {
			p.logger.Error().Err(result.Error).Msg("Failed to persist sample")
		}
	}
}

func (p *Persistor) readTransactions(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	serverStream, err := metricsClient.GetTransactions(ctx, &telemetry.TransactionStreamRequest{})
	if err != nil {
		return err
	}

	for {
		transaction, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		dbTransaction := NewTransactionFromGrpc(transaction)
		dbTransaction.AgentId = agentId
		p.logger.Trace().Str("name", dbTransaction.Name).Msg("Persisting transaction")
		if result := p.db.Save(&dbTransaction); result.Error != nil {
			p.logger.Error().Err(result.Error).Msg("Failed to persist transaction")
		}
	}
}

func (p *Persistor) readIterationCounters(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	serverStream, err := metricsClient.GetIterationCounters(ctx, &telemetry.IterationCountersStreamRequest{})
	if err != nil {
		return err
	}

	for {
		counters, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		dbCounters := NewIterationCountersFromGrpc(counters)
		dbCounters.AgentId = agentId
		p.logger.Trace().Str(log.AgentId, dbCounters.AgentId).Msg("Persisting iteration counter")
		if result := p.db.Save(&dbCounters); result.Error != nil {
			p.logger.Error().Err(result.Error).Msg("Failed to persist iteration counter")
		}
	}
}

func (p *Persistor) readHostMetrics(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	serverStream, err := metricsClient.GetHostMetrics(ctx, &telemetry.HostMetricsStreamRequest{})
	if err != nil {
		return err
	}

	for {
		metrics, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		dbMetrics := NewHostMetricFromGrpc(metrics)
		dbMetrics.AgentId = agentId
		p.logger.Trace().Str(log.AgentId, dbMetrics.AgentId).Msg("Persisting host metrics")
		if result := p.db.Save(&dbMetrics); result.Error != nil {
			p.logger.Error().Err(result.Error).Msg("Failed to persist host metrics")
		}
	}
}

func (p *Persistor) readRawLogs(ctx context.Context, agentId string, logsClient telemetry.LogsClient) error {
	serverStream, err := logsClient.GetLogEntries(ctx, &telemetry.LogEntriesStreamRequest{})
	if err != nil {
		return err
	}

	for {
		logEntry, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		dbLog := NewLogFromGrpc(logEntry)
		dbLog.AgentId = agentId
		p.logger.Trace().Str(log.AgentId, dbLog.AgentId).Msg("Persisting raw log")
		if result := p.db.Save(&dbLog); result.Error != nil {
			p.logger.Error().Err(result.Error).Msg("Failed to persist raw log")
		}
	}
}
