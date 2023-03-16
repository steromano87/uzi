package heartbeat_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type HeartbeatServerTestSuite struct {
	suite.Suite
	ctx                    context.Context
	cancelFunc             context.CancelCauseFunc
	heartbeatClient        *heartbeat.GrpcClient
	heartbeatInProcChannel *inprocgrpc.Channel
}

func (s *HeartbeatServerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	tempCtx, cancelFunc := context.WithCancelCause(context.TODO())
	s.ctx = logger.WithContext(tempCtx)
	s.cancelFunc = cancelFunc

	s.heartbeatInProcChannel = &inprocgrpc.Channel{}

	s.heartbeatClient = heartbeat.NewGrpcClient(s.heartbeatInProcChannel)
}

func (s *HeartbeatServerTestSuite) TestServerStatusGrpcRequest() {
	server, _, err := heartbeat.NewServer(s.ctx, 10*time.Millisecond, 15*time.Millisecond)
	if assert.NoError(s.T(), err) {
		heartbeat.RegisterHeartbeatServer(s.heartbeatInProcChannel, server)

		response, err := s.heartbeatClient.GetStatus(context.TODO(), &heartbeat.StatusRequest{})
		if assert.NoError(s.T(), err) {
			assert.IsType(s.T(), &heartbeat.StatusResponse{}, response)
			assert.Equal(s.T(), heartbeat.LockStatus_AVAILABLE, response.GetStatus())
			assert.Equal(s.T(), version.Version, response.GetVersion())
		}
	}
}

func (s *HeartbeatServerTestSuite) TestServerLockGrpcRequest() {
	server, _, err := heartbeat.NewServer(s.ctx, 10*time.Millisecond, 15*time.Millisecond)

	if assert.NoError(s.T(), err) {
		heartbeat.RegisterHeartbeatServer(s.heartbeatInProcChannel, server)

		go s.heartbeatClient.Monitor(s.ctx, 10*time.Millisecond, 15*time.Millisecond)
		defer s.cancelFunc(context.Canceled)
		time.Sleep(50 * time.Millisecond)
		statusResponse, err := s.heartbeatClient.GetStatus(context.TODO(), &heartbeat.StatusRequest{})

		if assert.NoError(s.T(), err) {
			status := statusResponse.GetStatus()
			assert.Equal(s.T(), heartbeat.LockStatus_LOCKED, status)
		}

		s.cancelFunc(context.Canceled)
		time.Sleep(20 * time.Millisecond)
		statusResponse, err = s.heartbeatClient.GetStatus(context.TODO(), &heartbeat.StatusRequest{})
		if assert.NoError(s.T(), err) {
			status := statusResponse.GetStatus()
			assert.Equal(s.T(), heartbeat.LockStatus_AVAILABLE, status)
		}
	}
}

func TestHeartbeatServerTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatServerTestSuite))
}
