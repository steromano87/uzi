package injector_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ServerTestSuite struct {
	suite.Suite
	ctx                   context.Context
	cancelFunc            context.CancelCauseFunc
	injectorClient        injector.InjectorClient
	heartbeatClient       *heartbeat.GrpcClient
	injectorInProcChannel *inprocgrpc.Channel
}

func (s *ServerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	tempCtx, cancelFunc := context.WithCancelCause(context.TODO())
	s.ctx = logger.WithContext(tempCtx)
	s.cancelFunc = cancelFunc

	s.injectorInProcChannel = &inprocgrpc.Channel{}
	s.injectorClient = injector.NewInjectorClient(s.injectorInProcChannel)

	s.heartbeatClient = heartbeat.NewGrpcClient(s.injectorInProcChannel)
}

func (s *ServerTestSuite) registerLocalGrpcServer(server *injector.Server) injector.InjectorClient {
	grpcChannel := &inprocgrpc.Channel{}
	injector.RegisterInjectorServer(grpcChannel, server)

	return injector.NewInjectorClient(grpcChannel)
}

func (s *ServerTestSuite) TestNewServer() {
	server := injector.NewServer()
	assert.IsType(s.T(), &injector.Server{}, server)
}

func (s *ServerTestSuite) TestServerStartupAndShutdown() {
	server := injector.NewServer()
	ctx, cancelFunc := context.WithCancel(s.ctx)
	server.Run(ctx, 10*time.Millisecond, 20*time.Millisecond)

	if assert.DirExists(s.T(), server.WorkingFolder()) {
		cancelFunc()
		time.Sleep(50 * time.Millisecond)
		assert.NoDirExists(s.T(), server.WorkingFolder())
	}
}

func (s *ServerTestSuite) TestGrpcServer() {
	server := injector.NewServer()
	client := s.registerLocalGrpcServer(server)
	if assert.Implements(s.T(), (*injector.InjectorClient)(nil), client) {
		counters, err := client.GetSyntheticUserCounters(s.ctx, &injector.SyntheticUserCountersRequest{})
		if assert.NoError(s.T(), err) {
			assert.EqualValues(s.T(), 0, counters.GetRequested())
			assert.EqualValues(s.T(), 0, counters.GetMaxQuota())
			assert.EqualValues(s.T(), 0, counters.GetReady())
			assert.EqualValues(s.T(), 0, counters.GetRunning())
		}
	}
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
