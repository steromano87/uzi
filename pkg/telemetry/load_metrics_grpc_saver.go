package telemetry

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"google.golang.org/grpc"
	"sync/atomic"
)

type LoadMetricsGrpcSaver struct {
	grpcClient            LoadMetricsClient
	hostId                string
	sampleSendStream      LoadMetrics_SaveSampleClient
	transactionSendStream LoadMetrics_SaveTransactionClient
	logger                zerolog.Logger

	noOp atomic.Bool
}

func NewLoadMetricsGrpcSaver(grpcClient LoadMetricsClient) *LoadMetricsGrpcSaver {
	saver := new(LoadMetricsGrpcSaver)
	saver.grpcClient = grpcClient
	saver.noOp.Store(true)

	return saver
}

func (l *LoadMetricsGrpcSaver) Start(ctx context.Context, opts ...grpc.CallOption) error {
	l.logger = zerolog.Ctx(ctx).With().Str(log.ComponentKey, "Load metrics GRPC Writer").Logger()
	l.retrieveHostIdFromCtx(ctx)

	l.logger.Info().Msg("Opening sample save stream")
	sampleSendStream, err := l.grpcClient.SaveSample(ctx, opts...)
	if err != nil {
		l.logger.Error().Err(err).Msg("Error when opening sample save stream")
		return err
	}

	l.logger.Info().Msg("Sample save stream opened, samples streaming started")
	l.sampleSendStream = sampleSendStream

	l.logger.Info().Msg("Opening transaction save stream")
	transactionSendStream, err := l.grpcClient.SaveTransaction(ctx, opts...)
	if err != nil {
		l.logger.Error().Err(err).Msg("Error when opening transaction save stream")
		return err
	}

	l.logger.Info().Msg("Transaction save stream opened, transactions streaming started")
	l.transactionSendStream = transactionSendStream
	l.noOp.Store(false)

	go l.handleStreamClosure(ctx)
	return nil
}

func (l *LoadMetricsGrpcSaver) retrieveHostIdFromCtx(ctx context.Context) {
	hostId := ctx.Value(HostIdCtxKey)

	// If no host id has been passed to the context, return and leave the default empty value
	if hostId == nil {
		return
	}

	l.hostId = hostId.(string)
}

func (l *LoadMetricsGrpcSaver) handleStreamClosure(ctx context.Context) {
	select {
	case <-ctx.Done():
		l.noOp.Store(true)

		l.logger.Info().AnErr("reason", context.Cause(ctx)).Msg("Closing samples save stream")
		sampleReply, err := l.sampleSendStream.CloseAndRecv()
		if err != nil {
			l.logger.Error().Err(err).Msg("Error when closing sample save stream")
		}
		l.logger.Info().Uint64(
			"savedSamples", sampleReply.GetSavedSamples()).Msg("Samples save stream successfully closed")

		l.logger.Info().AnErr("reason", context.Cause(ctx)).Msg("Closing transaction save stream")
		transactionReply, err := l.transactionSendStream.CloseAndRecv()
		if err != nil {
			l.logger.Error().Err(err).Msg("Error when closing sample save stream")
		}
		l.logger.Info().Uint64(
			"savedTransactions", transactionReply.GetSavedTransactions()).Msg(
			"Transactions save stream successfully closed")
	}
}

func (l *LoadMetricsGrpcSaver) SaveSample(sample *Sample) {
	if l.noOp.Load() {
		return
	}

	sample.HostId = l.hostId
	if err := l.sampleSendStream.Send(sample); err != nil {
		l.logger.Error().Err(err).Str("content", sample.String()).Msg("Error when sending sample")
	}
}

func (l *LoadMetricsGrpcSaver) SaveTransaction(transaction *Transaction) {
	if l.noOp.Load() {
		return
	}

	transaction.HostId = l.hostId
	if err := l.transactionSendStream.Send(transaction); err != nil {
		l.logger.Error().Err(err).Str("content", transaction.String()).Msg("Error when sending transaction")
	}
}
