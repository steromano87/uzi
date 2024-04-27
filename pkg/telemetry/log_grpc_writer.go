package telemetry

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"google.golang.org/grpc"
	"sync/atomic"
)

type LogGrpcWriter struct {
	grpcClient LogClient
	hostId     string
	sendStream Log_SaveClient
	logger     zerolog.Logger

	noOp atomic.Bool
}

func NewLogGrpcWriter(grpcClient LogClient) *LogGrpcWriter {
	writer := new(LogGrpcWriter)
	writer.grpcClient = grpcClient
	writer.noOp.Store(true)

	return writer
}

func (w *LogGrpcWriter) Start(ctx context.Context, opts ...grpc.CallOption) error {
	w.logger = zerolog.Ctx(ctx).With().Str(log.ComponentKey, "Log GRPC Writer").Logger()
	w.retrieveHostIdFromCtx(ctx)

	w.logger.Info().Msg("Opening log save stream")
	sendStream, err := w.grpcClient.Save(ctx, opts...)
	if err != nil {
		w.logger.Error().Err(err).Msg("Error when opening log save stream")
		return err
	}

	w.logger.Info().Msg("Log save stream opened, logs streaming started")
	w.sendStream = sendStream
	w.noOp.Store(false)

	go w.handleStreamClosure(ctx)
	return nil
}

func (w *LogGrpcWriter) retrieveHostIdFromCtx(ctx context.Context) {
	hostId := ctx.Value(HostIdCtxKey)

	// If no host id has been passed to the context, return and leave the default empty value
	if hostId == nil {
		return
	}

	w.hostId = hostId.(string)
}

func (w *LogGrpcWriter) handleStreamClosure(ctx context.Context) {
	select {
	case <-ctx.Done():
		w.logger.Info().AnErr("reason", context.Cause(ctx)).Msg("Closing log save stream")
		w.noOp.Store(true)
		reply, err := w.sendStream.CloseAndRecv()
		if err != nil {
			w.logger.Error().Err(err).Msg("Error when closing log save stream")
		}
		w.logger.Info().Uint64("savedLogs", reply.GetSavedLogs()).Msg("Log save stream successfully closed")
	}
}

func (w *LogGrpcWriter) Write(p []byte) (n int, err error) {
	// If the sending is not enabled (either because the GRPC client is not yet connected
	// or because the send stream has already been closed, go no-op and return immediately
	if w.noOp.Load() {
		return len(p), nil
	}

	rawLog := RawLog{
		HostId: w.hostId,
		Entry:  p,
	}
	if err := w.sendStream.Send(&rawLog); err != nil {
		return 0, err
	}
	return len(p), nil
}
