package db

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/uzi/v1/pkg/log"
	"github.com/steromano87/uzi/v1/pkg/syntheticuser"
	"github.com/steromano87/uzi/v1/pkg/telemetry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"io"
	"time"
)

const (
	syntheticUserPollInterval             = 5 * time.Second
	recordsChanSize                       = 2048
	persistenceTransactionCommitInterval  = 5 * time.Second
	persistenceTransactionCommitBatchSize = 100
)

type Persistor struct {
	db                         *gorm.DB
	logger                     zerolog.Logger
	innerTerminationCancelFunc context.CancelFunc

	samplesChan               chan Sample
	transactionChan           chan Transaction
	iterationCountersChan     chan IterationCounters
	hostMetricsChan           chan HostMetric
	rawLogsChan               chan RawLog
	syntheticUserCountersChan chan SyntheticUserCounters
}

func NewPersistor(db *gorm.DB) *Persistor {
	persistor := new(Persistor)
	persistor.db = db

	persistor.samplesChan = make(chan Sample, recordsChanSize)
	persistor.transactionChan = make(chan Transaction, recordsChanSize)
	persistor.iterationCountersChan = make(chan IterationCounters, recordsChanSize)
	persistor.hostMetricsChan = make(chan HostMetric, recordsChanSize)
	persistor.rawLogsChan = make(chan RawLog, recordsChanSize)
	persistor.syntheticUserCountersChan = make(chan SyntheticUserCounters, recordsChanSize)
	return persistor
}

func (p *Persistor) Serve(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient, logsClient telemetry.LogsClient, spawnerClient syntheticuser.SpawnerClient) error {
	p.setLogger(zerolog.Ctx(ctx))
	var innerTerminationCtx context.Context
	innerTerminationCtx, p.innerTerminationCancelFunc = context.WithCancel(ctx)

	clientErrGroup, _ := errgroup.WithContext(ctx)
	clientErrGroup.Go(func() error {
		return p.persistRecords(innerTerminationCtx)
	})
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
	clientErrGroup.Go(func() error {
		return p.readSyntheticUserCounters(innerTerminationCtx, agentId, spawnerClient)
	})

	return clientErrGroup.Wait()
}

func (p *Persistor) setLogger(logger *zerolog.Logger) {
	p.logger = logger.With().Str(log.ComponentKey, "DB Persistor").Logger()
}

func (p *Persistor) readSamples(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	defer p.innerTerminationCancelFunc()
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
		p.samplesChan <- dbSample
	}
}

func (p *Persistor) readTransactions(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	defer p.innerTerminationCancelFunc()
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
		p.transactionChan <- dbTransaction
	}
}

func (p *Persistor) readIterationCounters(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	defer p.innerTerminationCancelFunc()
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
		p.iterationCountersChan <- dbCounters
	}
}

func (p *Persistor) readHostMetrics(ctx context.Context, agentId string, metricsClient telemetry.MetricsClient) error {
	defer p.innerTerminationCancelFunc()
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
		p.hostMetricsChan <- dbMetrics
	}
}

func (p *Persistor) readRawLogs(ctx context.Context, agentId string, logsClient telemetry.LogsClient) error {
	defer p.innerTerminationCancelFunc()
	serverStream, err := logsClient.GetRawLogs(ctx, &telemetry.RawLogsStreamRequest{})
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
		p.rawLogsChan <- dbLog
	}
}

func (p *Persistor) readSyntheticUserCounters(ctx context.Context, agentId string, spawnerClient syntheticuser.SpawnerClient) error {
	ticker := time.NewTicker(syntheticUserPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("Context canceled, stopping synthetic user counters gathering")
			return nil

		case now := <-ticker.C:
			syntheticUserCounters, err := spawnerClient.GetSyntheticUserCounters(ctx, &emptypb.Empty{})
			if err != nil {
				p.logger.Error().Err(err).Msg("Failed to read synthetic user counters")
				return err
			}
			dbCounters := NewSyntheticUserCountersFromGrpc(syntheticUserCounters)
			dbCounters.AgentId = agentId
			dbCounters.Timestamp = now
			p.syntheticUserCountersChan <- dbCounters
		}
	}
}

func (p *Persistor) persistRecords(ctx context.Context) error {
	p.logger.Debug().Int("commitBatchSize", persistenceTransactionCommitBatchSize).Dur("commitMaxInterval", persistenceTransactionCommitInterval).Msg("Persistence goroutine started")
	commitTicker := time.NewTicker(persistenceTransactionCommitInterval)
	defer commitTicker.Stop()
	transaction := p.db.Begin()
	recordCount := 0
	defer transaction.Commit()

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("Persistence goroutine shutdown requested, committing all remaining records")
			for len(p.samplesChan) > 0 {
				record := <-p.samplesChan
				transaction.Create(&record)
			}

			for len(p.transactionChan) > 0 {
				record := <-p.transactionChan
				transaction.Create(&record)
			}

			for len(p.iterationCountersChan) > 0 {
				record := <-p.iterationCountersChan
				transaction.Create(&record)
			}

			for len(p.hostMetricsChan) > 0 {
				record := <-p.hostMetricsChan
				transaction.Create(&record)
			}

			for len(p.rawLogsChan) > 0 {
				record := <-p.rawLogsChan
				transaction.Create(&record)
			}

			for len(p.syntheticUserCountersChan) > 0 {
				record := <-p.syntheticUserCountersChan
				transaction.Create(&record)
			}

			p.logger.Debug().Msg("All pending records committed, stopping")
			return nil

		case record := <-p.samplesChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case record := <-p.transactionChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case record := <-p.iterationCountersChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case record := <-p.hostMetricsChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case record := <-p.rawLogsChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case record := <-p.syntheticUserCountersChan:
			transaction.Create(&record)
			recordCount++
			if recordCount >= persistenceTransactionCommitBatchSize {
				p.logger.Trace().Int("recordCount", recordCount).Msg("Records have reached the maximum batch size, committing")
				transaction.Commit()
				transaction = p.db.Begin()
				recordCount = 0
			}

		case <-commitTicker.C:
			if recordCount == 0 {
				p.logger.Trace().Msg("Skip transaction commit because no records have been added since last transaction start")
				continue
			}
			transaction.Commit()
			transaction = p.db.Begin()
			recordCount = 0
		}
	}
}
