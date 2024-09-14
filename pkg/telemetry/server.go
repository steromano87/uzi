package telemetry

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	harkErrors "github.com/steromano87/harkonnen/v1/pkg/errors"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
)

const semaphoreWeight = 5

var ErrFullBuffer = errors.New("full buffer")

type Server struct {
	UnimplementedMetricsServer
	UnimplementedLogsServer

	activeSession atomic.Bool

	samplesBuffer           chan *Sample
	transactionsBuffer      chan *Transaction
	iterationCountersBuffer chan *IterationCounters
	logsBuffer              chan *LogEntry
	hostMetricsBuffer       chan *HostMetrics
	buffersMu               sync.RWMutex

	logger zerolog.Logger

	controlCtx      context.Context
	streamSemaphore *semaphore.Weighted
}

func NewServer() *Server {
	server := new(Server)
	config := configuration.MustNewDefault()
	server.samplesBuffer = make(chan *Sample, config.Telemetry.LoadMetrics.BufferCapacity)
	server.transactionsBuffer = make(chan *Transaction, config.Telemetry.LoadMetrics.BufferCapacity)
	server.iterationCountersBuffer = make(chan *IterationCounters, config.Telemetry.LoadMetrics.BufferCapacity)
	server.logsBuffer = make(chan *LogEntry, config.Telemetry.Logs.BufferCapacity)
	server.hostMetricsBuffer = make(chan *HostMetrics, config.Telemetry.HostMetrics.BufferCapacity)
	server.streamSemaphore = semaphore.NewWeighted(semaphoreWeight)

	server.setLogger(zerolog.Nop())

	return server
}

func (s *Server) Reconfigure(config *configuration.Manifest) error {
	if s.activeSession.Load() {
		return harkErrors.SessionAlreadyInProgress
	}

	s.buffersMu.Lock()
	defer s.buffersMu.Unlock()
	s.samplesBuffer = make(chan *Sample, config.Telemetry.LoadMetrics.BufferCapacity)
	s.transactionsBuffer = make(chan *Transaction, config.Telemetry.LoadMetrics.BufferCapacity)
	s.iterationCountersBuffer = make(chan *IterationCounters, config.Telemetry.LoadMetrics.BufferCapacity)
	s.logsBuffer = make(chan *LogEntry, config.Telemetry.Logs.BufferCapacity)
	s.hostMetricsBuffer = make(chan *HostMetrics, config.Telemetry.HostMetrics.BufferCapacity)

	return nil
}

func (s *Server) ServeSession(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	s.setLogger(*logger)
	s.logger.Info().Msg("Telemetry server started")
	defer s.logger.Info().Msg("Telemetry server stopped")
	s.controlCtx = ctx

	s.activeSession.Store(true)
	defer s.activeSession.Store(false)

	<-s.controlCtx.Done()
	return s.streamSemaphore.Acquire(context.Background(), semaphoreWeight)
}

func (s *Server) setLogger(logger zerolog.Logger) {
	s.logger = logger.With().Str(log.ComponentKey, "Telemetry server").Logger()
}

///////////////////////////
// Storer implementation //
///////////////////////////

func (s *Server) StoreSample(sample *Sample) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	s.buffersMu.RLock()
	defer s.buffersMu.RUnlock()

	select {
	case s.samplesBuffer <- sample:
		s.logger.Trace().Msg("Saved sample")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreTransaction(transaction *Transaction) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	s.buffersMu.RLock()
	defer s.buffersMu.RUnlock()

	select {
	case s.transactionsBuffer <- transaction:
		s.logger.Trace().Msg("Saved transaction")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreIterationCounters(counters *IterationCounters) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	s.buffersMu.RLock()
	defer s.buffersMu.RUnlock()

	select {
	case s.iterationCountersBuffer <- counters:
		s.logger.Trace().Msg("Saved iteration counters")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreHostMetrics(agentMetrics *HostMetrics) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	s.buffersMu.RLock()
	defer s.buffersMu.RUnlock()

	select {
	case s.hostMetricsBuffer <- agentMetrics:
		s.logger.Trace().Msg("Saved host metrics")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreLogEntry(entry *LogEntry) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	s.buffersMu.RLock()
	defer s.buffersMu.RUnlock()

	select {
	case s.logsBuffer <- entry:
		s.logger.Trace().Msg("Saved log entry")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) Write(p []byte) (n int, err error) {
	if !s.activeSession.Load() {
		return 0, harkErrors.NoSessionsInProgress
	}

	entry := &LogEntry{
		RawData: p,
	}
	if err := s.StoreLogEntry(entry); err != nil {
		return 0, err
	}

	return len(p), nil
}

/////////////////////////
// GRPC implementation //
/////////////////////////

func (s *Server) GetSamples(_ *SampleStreamRequest, g grpc.ServerStreamingServer[Sample]) error {
	if !s.activeSession.Load() {
		s.logger.Error().Err(harkErrors.NoSessionsInProgress).Msg("Cannot start sample streaming")
		return harkErrors.NoSessionsInProgress
	}

	if err := s.streamSemaphore.Acquire(s.controlCtx, 1); err != nil {
		s.logger.Error().Err(err).Msg("Cannot start sample streaming")
		return err
	}
	defer s.streamSemaphore.Release(1)

	s.logger.Info().Msg("Serve sample streaming")

	for {
		select {
		case sample := <-s.samplesBuffer:
			s.logger.Trace().Msg("Streaming sample")
			if err := g.Send(sample); err != nil {
				s.logger.Error().Err(err).Msg("Encountered an error while streaming samples")
				return err
			}

		case <-s.controlCtx.Done():
			s.logger.Info().Msg("Shutdown request received, closing samples stream")
			return nil
		}
	}
}

func (s *Server) GetTransactions(_ *TransactionStreamRequest, g grpc.ServerStreamingServer[Transaction]) error {
	if !s.activeSession.Load() {
		s.logger.Error().Err(harkErrors.NoSessionsInProgress).Msg("Cannot start transaction streaming")
		return harkErrors.NoSessionsInProgress
	}

	if err := s.streamSemaphore.Acquire(s.controlCtx, 1); err != nil {
		s.logger.Error().Err(err).Msg("Cannot start transaction streaming")
		return err
	}
	defer s.streamSemaphore.Release(1)

	s.logger.Info().Msg("Serve transaction streaming")

	for {
		select {
		case transaction := <-s.transactionsBuffer:
			s.logger.Trace().Msg("Streaming transaction")
			if err := g.Send(transaction); err != nil {
				s.logger.Error().Err(err).Msg("Encountered an error while streaming transactions")
				return err
			}

		case <-s.controlCtx.Done():
			s.logger.Info().Msg("Shutdown request received, closing transactions stream")
			return nil
		}
	}
}

func (s *Server) GetIterationCounters(_ *IterationCountersStreamRequest, g grpc.ServerStreamingServer[IterationCounters]) error {
	if !s.activeSession.Load() {
		s.logger.Error().Err(harkErrors.NoSessionsInProgress).Msg("Cannot start iteration counters streaming")
		return harkErrors.NoSessionsInProgress
	}

	if err := s.streamSemaphore.Acquire(s.controlCtx, 1); err != nil {
		s.logger.Error().Err(err).Msg("Cannot start iteration counters streaming")
		return err
	}
	defer s.streamSemaphore.Release(1)

	s.logger.Info().Msg("Serve iteration counters streaming")

	for {
		select {
		case iterationCounter := <-s.iterationCountersBuffer:
			s.logger.Trace().Msg("Streaming iteration counters")
			if err := g.Send(iterationCounter); err != nil {
				s.logger.Error().Err(err).Msg("Encountered an error while streaming iteration counters")
				return err
			}

		case <-s.controlCtx.Done():
			s.logger.Info().Msg("Shutdown request received, closing iteration counters stream")
			return nil
		}
	}
}

func (s *Server) GetHostMetrics(_ *HostMetricsStreamRequest, g grpc.ServerStreamingServer[HostMetrics]) error {
	if !s.activeSession.Load() {
		s.logger.Error().Err(harkErrors.NoSessionsInProgress).Msg("Cannot start host metrics streaming")
		return harkErrors.NoSessionsInProgress
	}

	if err := s.streamSemaphore.Acquire(s.controlCtx, 1); err != nil {
		s.logger.Error().Err(err).Msg("Cannot start host metrics streaming")
		return err
	}
	defer s.streamSemaphore.Release(1)

	s.logger.Info().Msg("Serve host metrics streaming")

	for {
		select {
		case hostMetric := <-s.hostMetricsBuffer:
			s.logger.Trace().Msg("Streaming host metrics")
			if err := g.Send(hostMetric); err != nil {
				s.logger.Error().Err(err).Msg("Encountered an error while streaming host metrics")
				return err
			}

		case <-s.controlCtx.Done():
			s.logger.Info().Msg("Shutdown request received, closing host metrics stream")
			return nil
		}
	}
}

func (s *Server) GetLogEntries(_ *LogEntriesStreamRequest, g grpc.ServerStreamingServer[LogEntry]) error {
	if !s.activeSession.Load() {
		s.logger.Error().Err(harkErrors.NoSessionsInProgress).Msg("Cannot start logs streaming")
		return harkErrors.NoSessionsInProgress
	}

	if err := s.streamSemaphore.Acquire(s.controlCtx, 1); err != nil {
		s.logger.Error().Err(err).Msg("Cannot start logs streaming")
		return err
	}
	defer s.streamSemaphore.Release(1)

	s.logger.Info().Msg("Serve logs streaming")

	for {
		select {
		case logEntry := <-s.logsBuffer:
			s.logger.Trace().Msg("Streaming log entry")
			if err := g.Send(logEntry); err != nil {
				s.logger.Error().Err(err).Msg("Encountered an error while streaming log entry")
				return err
			}

		case <-s.controlCtx.Done():
			s.logger.Info().Msg("Shutdown request received, closing logs stream")
			return nil
		}
	}
}
