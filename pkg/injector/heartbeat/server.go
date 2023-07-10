package heartbeat

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type Server struct {
	mainCtx                context.Context
	controlCancelCauseFunc context.CancelCauseFunc

	logger  *zerolog.Logger
	monitor *Monitor
	status  LockStatus

	UnimplementedHeartbeatServer
}

func NewServer(ctx context.Context, beatInterval, beatTimeout time.Duration) (*Server, context.Context, error) {
	server := new(Server)
	server.mainCtx = ctx
	server.logger = zerolog.Ctx(ctx)
	monitor, err := NewServerMonitor(beatInterval, beatTimeout)
	if err != nil {
		return nil, nil, err
	}
	server.monitor = monitor
	server.status = LockStatus_AVAILABLE
	controlCtx, controlCancelCauseFunc := context.WithCancelCause(ctx)
	server.controlCancelCauseFunc = controlCancelCauseFunc

	return server, controlCtx, nil
}

func (s *Server) GetStatus(_ context.Context, _ *StatusRequest) (*StatusResponse, error) {
	s.contextualizedLogger().Info().Msg("Received status request")
	statusResponse := &StatusResponse{
		Status:  s.status,
		Version: version.Version,
	}
	s.contextualizedLogger().Debug().Msgf("Current status is %s", s.status.String())
	return statusResponse, nil
}

func (s *Server) Lock(stream Heartbeat_LockServer) error {
	s.contextualizedLogger().Info().Msg("Received lock request")
	if s.status == LockStatus_LOCKED {
		err := errors.New("injector cannot be locked because is is already in LOCKED status")
		s.contextualizedLogger().Error().Err(err).Msg("InjectorConfiguration cannot accept lock request")
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}

	s.status = LockStatus_LOCKED
	defer func() {
		s.monitor.Stop()
		s.status = LockStatus_AVAILABLE
		s.contextualizedLogger().Info().Msg("Lock released")
	}()
	s.contextualizedLogger().Info().Msg("Lock acquired")

	controlCtx := s.monitor.Start(s.mainCtx)

	incomingBeatChan := make(chan error)
	outgoingBeatChan := make(chan error)
	go func() {
		incomingBeatChan <- s.incomingBeatLoop(stream)
	}()
	go func() {
		outgoingBeatChan <- s.outgoingBeatLoop(controlCtx, stream)
	}()

	select {
	case err := <-incomingBeatChan:
		s.contextualizedLogger().Info().Err(err).Msg("Releasing lock due to incoming beat loop termination")
		return err

	case err := <-outgoingBeatChan:
		s.contextualizedLogger().Info().Err(err).Msg("Releasing lock due to outgoing beat loop termination")
		return err
	}
}

func (s *Server) incomingBeatLoop(stream Heartbeat_LockServer) error {
	for {
		_, err := stream.Recv()

		if err == io.EOF {
			s.contextualizedLogger().Info().Msg("Lock channel remotely closed, stopping...")
			return nil
		}

		if err == context.Canceled {
			s.contextualizedLogger().Info().Msg("Remote context gracefully canceled, stopping...")
			return nil
		}

		if err != nil {
			s.contextualizedLogger().Error().Err(err).Msg("Error when receiving heartbeat")
			return err
		}

		s.contextualizedLogger().Debug().Msg("Received heartbeat")
		s.monitor.IncomingBeat()
	}
}

func (s *Server) outgoingBeatLoop(controlCtx context.Context, stream Heartbeat_LockServer) error {
	for {
		select {
		case <-controlCtx.Done():
			if context.Cause(controlCtx) == ErrHeartbeatRequestTimeout {
				s.contextualizedLogger().Error().Err(context.Cause(controlCtx)).Msg("Timeout receiving heartbeat, releasing lock...")
				s.controlCancelCauseFunc(ErrHeartbeatRequestTimeout)
				return status.Error(codes.Aborted, "Heartbeat timeout exceeded")
			}

			if context.Cause(controlCtx) == context.Canceled {
				s.contextualizedLogger().Info().Err(context.Cause(controlCtx)).Msg("Graceful termination requested")
				s.controlCancelCauseFunc(context.Canceled)
				return nil
			}

			s.contextualizedLogger().Warn().Err(context.Cause(controlCtx)).Msg("Context canceled for unhandled reason, exiting...")
			s.controlCancelCauseFunc(context.Cause(controlCtx))
			return context.Cause(controlCtx)

		case <-s.monitor.OutgoingBeat():
			s.contextualizedLogger().Debug().Msg("Sending heartbeat")
			if err := stream.Send(&HeartbeatResponse{Timestamp: timestamppb.Now()}); err != nil {
				s.contextualizedLogger().Error().Err(err).Msg("Error sending outgoing beat")
				return status.Error(codes.Unknown, err.Error())
			}
		}
	}
}

func (s *Server) contextualizedLogger() *zerolog.Logger {
	logger := s.logger.With().Str("component", "Heartbeat Server").Logger()
	return &logger
}
