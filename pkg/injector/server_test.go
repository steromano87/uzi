package injector_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
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

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
