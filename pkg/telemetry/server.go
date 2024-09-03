package telemetry

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"google.golang.org/grpc"
)

var ErrFullBuffer = errors.New("full buffer")

type Server struct {
	UnimplementedMetricsServer
	UnimplementedLogsServer
	samplesBuffer           chan *Sample
	transactionsBuffer      chan *Transaction
	iterationCountersBuffer chan *IterationCounters
	logsBuffer              chan *LogEntry
	hostMetricsBuffer       chan *HostMetrics
	logger                  zerolog.Logger
	controlCtx              context.Context
}

func NewServer(ctx context.Context, config *configuration.Manifest) *Server {
	server := new(Server)
	server.controlCtx = ctx
	server.samplesBuffer = make(chan *Sample, config.Telemetry.LoadMetrics.BufferCapacity)
	server.transactionsBuffer = make(chan *Transaction, config.Telemetry.LoadMetrics.BufferCapacity)
	server.iterationCountersBuffer = make(chan *IterationCounters, config.Telemetry.LoadMetrics.BufferCapacity)
	server.logsBuffer = make(chan *LogEntry, config.Telemetry.Logs.BufferCapacity)
	server.hostMetricsBuffer = make(chan *HostMetrics, config.Telemetry.HostMetrics.BufferCapacity)
	server.logger = zerolog.Nop()

	return server
}

///////////////////////////
// Storer implementation //
///////////////////////////

func (s *Server) StoreSample(sample *Sample) error {
	select {
	case s.samplesBuffer <- sample:
		s.logger.Trace().Msg("Saved sample")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreTransaction(transaction *Transaction) error {
	select {
	case s.transactionsBuffer <- transaction:
		s.logger.Trace().Msg("Saved transaction")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreIterationCounters(counters *IterationCounters) error {
	select {
	case s.iterationCountersBuffer <- counters:
		s.logger.Trace().Msg("Saved iteration counters")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) StoreLogEntry(entry *LogEntry) error {
	select {
	case s.logsBuffer <- entry:
		s.logger.Trace().Msg("Saved log entry")
		return nil
	default:
		return ErrFullBuffer
	}
}

func (s *Server) Write(p []byte) (n int, err error) {
	entry := &LogEntry{
		RawData: p,
	}
	if err := s.StoreLogEntry(entry); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (s *Server) StoreHostMetrics(agentMetrics *HostMetrics) error {
	select {
	case s.hostMetricsBuffer <- agentMetrics:
		s.logger.Trace().Msg("Saved host metrics")
		return nil
	default:
		return ErrFullBuffer
	}
}

/////////////////////////
// GRPC implementation //
/////////////////////////

func (s *Server) GetSamples(_ *SampleStreamRequest, g grpc.ServerStreamingServer[Sample]) error {
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
