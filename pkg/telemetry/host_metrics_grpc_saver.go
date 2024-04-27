package telemetry

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"google.golang.org/grpc"
	"sync/atomic"
)

type HostMetricsGrpcSaver struct {
	grpcClient       HostMetricsClient
	hostId           string
	sampleSendStream HostMetrics_SaveClient
	logger           zerolog.Logger

	noOp atomic.Bool
}

func NewHostMetricsGrpcSaver(grpcClient HostMetricsClient) *HostMetricsGrpcSaver {
	saver := new(HostMetricsGrpcSaver)
	saver.grpcClient = grpcClient
	saver.noOp.Store(true)

	return saver
}

func (s *HostMetricsGrpcSaver) Start(ctx context.Context, opts ...grpc.CallOption) error {
	s.logger = zerolog.Ctx(ctx).With().Str(log.ComponentKey, "Host samples GRPC Writer").Logger()
	s.retrieveHostIdFromCtx(ctx)

	s.logger.Info().Msg("Opening host metrics sample save stream")
	sampleSendStream, err := s.grpcClient.Save(ctx, opts...)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error when opening sample save stream")
		return err
	}

	s.logger.Info().Msg("Sample save stream opened, samples streaming started")
	s.sampleSendStream = sampleSendStream
	s.noOp.Store(false)

	go s.handleStreamClosure(ctx)
	return nil
}

func (s *HostMetricsGrpcSaver) retrieveHostIdFromCtx(ctx context.Context) {
	hostId := ctx.Value(HostIdCtxKey)

	// If no host id has been passed to the context, return and leave the default empty value
	if hostId == nil {
		return
	}

	s.hostId = hostId.(string)
}

func (s *HostMetricsGrpcSaver) handleStreamClosure(ctx context.Context) {
	select {
	case <-ctx.Done():
		s.noOp.Store(true)

		s.logger.Info().AnErr("reason", context.Cause(ctx)).Msg("Closing samples save stream")
		sampleReply, err := s.sampleSendStream.CloseAndRecv()
		if err != nil {
			s.logger.Error().Err(err).Msg("Error when closing sample save stream")
		}
		s.logger.Info().Uint64(
			"savedSamples", sampleReply.GetSavedHostMetrics()).Msg("Samples save stream successfully closed")
	}
}

func (s *HostMetricsGrpcSaver) SaveHostMetricsSample(sample *HostMetricsSample) {
	if s.noOp.Load() {
		return
	}

	sample.HostId = s.hostId
	if err := s.sampleSendStream.Send(sample); err != nil {
		s.logger.Error().Err(err).Str("content", sample.String()).Msg("Error when sending sample")
	}
}
