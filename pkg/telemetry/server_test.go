package telemetry_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"testing"
	"time"
)

type logHook struct {
	logEvents []zerolog.Event
}

func (logHook *logHook) Run(logEvent *zerolog.Event, _ zerolog.Level, _ string) {
	logHook.logEvents = append(logHook.logEvents, *logEvent)
}

/////////////////////

type ServerTestSuite struct {
	suite.Suite
	ctx           context.Context
	cancelFunc    context.CancelFunc
	configuration *configuration.Manifest
}

func (s *ServerTestSuite) SetupTest() {
	s.ctx, s.cancelFunc = context.WithCancel(context.TODO())
	s.configuration = configuration.MustNewDefault()
}

func (s *ServerTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *ServerTestSuite) TestNewServer() {
	server := telemetry.NewServer()
	assert.IsType(s.T(), &telemetry.Server{}, server)
}

func (s *ServerTestSuite) TestStoreSample_NoError() {
	server := telemetry.NewServer()
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	sample := &telemetry.Sample{
		Timestamp:       timestamppb.Now(),
		SyntheticUserId: "test",
		Duration:        durationpb.New(50 * time.Millisecond),
		Name:            "My sample",
		SentBytes:       0,
		ReceivedBytes:   0,
		Kind:            "custom",
		SampleData:      nil,
	}
	err := server.StoreSample(sample)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreSample_Error() {
	s.configuration.Telemetry.LoadMetrics.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	sample := &telemetry.Sample{
		Timestamp:       timestamppb.Now(),
		SyntheticUserId: "test",
		Duration:        durationpb.New(50 * time.Millisecond),
		Name:            "My sample",
		SentBytes:       0,
		ReceivedBytes:   0,
		Kind:            "custom",
		SampleData:      nil,
	}

	if assert.NoError(s.T(), server.StoreSample(sample)) {
		err := server.StoreSample(sample)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreSample_Retrieve() {
	server := telemetry.NewServer()
	sample := &telemetry.Sample{
		Timestamp:       timestamppb.Now(),
		SyntheticUserId: "test",
		Duration:        durationpb.New(50 * time.Millisecond),
		Name:            "My sample",
		SentBytes:       0,
		ReceivedBytes:   0,
		Kind:            "custom",
		SampleData:      nil,
	}

	// GRPC setup
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	grpcChannel := &inprocgrpc.Channel{}
	telemetry.RegisterMetricsServer(grpcChannel, server)
	metricsClient := telemetry.NewMetricsClient(grpcChannel)
	retrievedSamples := make([]*telemetry.Sample, 0)

	err := server.StoreSample(sample)
	if assert.NoError(s.T(), err) {
		serverStream, err := metricsClient.GetSamples(context.TODO(), &telemetry.SampleStreamRequest{})
		if assert.NoError(s.T(), err) {
			// Cancel the current context after 500 ms
			cancelFuncTimer := time.NewTimer(250 * time.Millisecond)
			go func() {
				<-cancelFuncTimer.C
				s.cancelFunc()
			}()

			for {
				retrievedSample, err := serverStream.Recv()
				if err == io.EOF {
					break
				}

				if assert.NoError(s.T(), err) {
					retrievedSamples = append(retrievedSamples, retrievedSample)
				}
			}

			if assert.Len(s.T(), retrievedSamples, 1) {
				assert.Equal(s.T(), sample, retrievedSamples[0])
			}
		}
	}
}

func (s *ServerTestSuite) TestStoreTransaction_NoError() {
	server := telemetry.NewServer()
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	transaction := &telemetry.Transaction{}
	err := server.StoreTransaction(transaction)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreTransaction_Error() {
	s.configuration.Telemetry.LoadMetrics.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	transaction := &telemetry.Transaction{}
	if assert.NoError(s.T(), server.StoreTransaction(transaction)) {
		err := server.StoreTransaction(transaction)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreTransaction_Retrieve() {
	server := telemetry.NewServer()
	transaction := &telemetry.Transaction{}

	// GRPC setup
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	grpcChannel := &inprocgrpc.Channel{}
	telemetry.RegisterMetricsServer(grpcChannel, server)
	metricsClient := telemetry.NewMetricsClient(grpcChannel)
	retrievedTransactions := make([]*telemetry.Transaction, 0)

	err := server.StoreTransaction(transaction)
	if assert.NoError(s.T(), err) {
		serverStream, err := metricsClient.GetTransactions(context.TODO(), &telemetry.TransactionStreamRequest{})
		if assert.NoError(s.T(), err) {
			// Cancel the current context after 500 ms
			cancelFuncTimer := time.NewTimer(250 * time.Millisecond)
			go func() {
				<-cancelFuncTimer.C
				s.cancelFunc()
			}()

			for {
				retrievedTransaction, err := serverStream.Recv()
				if err == io.EOF {
					break
				}

				if assert.NoError(s.T(), err) {
					retrievedTransactions = append(retrievedTransactions, retrievedTransaction)
				}
			}

			if assert.Len(s.T(), retrievedTransactions, 1) {
				assert.Equal(s.T(), transaction, retrievedTransactions[0])
			}
		}
	}
}

func (s *ServerTestSuite) TestStoreIterationCounters_NoError() {
	server := telemetry.NewServer()
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	counters := &telemetry.IterationCounters{}
	err := server.StoreIterationCounters(counters)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreIterationCounters_Error() {
	s.configuration.Telemetry.LoadMetrics.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	counters := &telemetry.IterationCounters{}
	if assert.NoError(s.T(), server.StoreIterationCounters(counters)) {
		err := server.StoreIterationCounters(counters)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreIterationCounters_Retrieve() {
	server := telemetry.NewServer()
	counters := &telemetry.IterationCounters{}

	// GRPC setup
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	grpcChannel := &inprocgrpc.Channel{}
	telemetry.RegisterMetricsServer(grpcChannel, server)
	metricsClient := telemetry.NewMetricsClient(grpcChannel)
	retrievedCounters := make([]*telemetry.IterationCounters, 0)

	err := server.StoreIterationCounters(counters)
	if assert.NoError(s.T(), err) {
		serverStream, err := metricsClient.GetIterationCounters(context.TODO(), &telemetry.IterationCountersStreamRequest{})
		if assert.NoError(s.T(), err) {
			// Cancel the current context after 500 ms
			cancelFuncTimer := time.NewTimer(250 * time.Millisecond)
			go func() {
				<-cancelFuncTimer.C
				s.cancelFunc()
			}()

			for {
				retrievedCounter, err := serverStream.Recv()
				if err == io.EOF {
					break
				}

				if assert.NoError(s.T(), err) {
					retrievedCounters = append(retrievedCounters, retrievedCounter)
				}
			}

			if assert.Len(s.T(), retrievedCounters, 1) {
				assert.Equal(s.T(), counters, retrievedCounters[0])
			}
		}
	}
}

func (s *ServerTestSuite) TestStoreHostMetrics_NoError() {
	server := telemetry.NewServer()
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	hostMetrics := &telemetry.HostMetrics{}
	err := server.StoreHostMetrics(hostMetrics)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreHostMetrics_Error() {
	s.configuration.Telemetry.HostMetrics.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	hostMetrics := &telemetry.HostMetrics{}
	if assert.NoError(s.T(), server.StoreHostMetrics(hostMetrics)) {
		err := server.StoreHostMetrics(hostMetrics)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreHostMetrics_Retrieve() {
	server := telemetry.NewServer()
	hostMetrics := &telemetry.HostMetrics{}

	// GRPC setup
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	grpcChannel := &inprocgrpc.Channel{}
	telemetry.RegisterMetricsServer(grpcChannel, server)
	metricsClient := telemetry.NewMetricsClient(grpcChannel)
	retrievedHostMetrics := make([]*telemetry.HostMetrics, 0)

	err := server.StoreHostMetrics(hostMetrics)
	if assert.NoError(s.T(), err) {
		serverStream, err := metricsClient.GetHostMetrics(context.TODO(), &telemetry.HostMetricsStreamRequest{})
		if assert.NoError(s.T(), err) {
			// Cancel the current context after 500 ms
			cancelFuncTimer := time.NewTimer(250 * time.Millisecond)
			go func() {
				<-cancelFuncTimer.C
				s.cancelFunc()
			}()

			for {
				retrievedHostMetric, err := serverStream.Recv()
				if err == io.EOF {
					break
				}

				if assert.NoError(s.T(), err) {
					retrievedHostMetrics = append(retrievedHostMetrics, retrievedHostMetric)
				}
			}

			if assert.Len(s.T(), retrievedHostMetrics, 1) {
				assert.Equal(s.T(), hostMetrics, retrievedHostMetrics[0])
			}
		}
	}
}

func (s *ServerTestSuite) TestStoreLogEntry_NoError() {
	server := telemetry.NewServer()
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	logEntry := &telemetry.LogEntry{}
	err := server.StoreLogEntry(logEntry)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreLogEntry_Error() {
	s.configuration.Telemetry.Logs.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	logEntry := &telemetry.LogEntry{}
	if assert.NoError(s.T(), server.StoreLogEntry(logEntry)) {
		err := server.StoreLogEntry(logEntry)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestWriteLogsAndRetrieveEntries() {
	server := telemetry.NewServer()
	logger := zerolog.New(server)

	// GRPC setup
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	grpcChannel := &inprocgrpc.Channel{}
	telemetry.RegisterLogsServer(grpcChannel, server)
	logsClient := telemetry.NewLogsClient(grpcChannel)
	retrievedLogEntries := make([]*telemetry.LogEntry, 0)

	logger.Info().Msg("test message")
	serverStream, err := logsClient.GetLogEntries(context.TODO(), &telemetry.LogEntriesStreamRequest{})
	if assert.NoError(s.T(), err) {
		// Cancel the current context after 500 ms
		cancelFuncTimer := time.NewTimer(250 * time.Millisecond)
		go func() {
			<-cancelFuncTimer.C
			s.cancelFunc()
		}()

		for {
			retrievedLogEntry, err := serverStream.Recv()
			if err == io.EOF {
				break
			}

			if assert.NoError(s.T(), err) {
				retrievedLogEntries = append(retrievedLogEntries, retrievedLogEntry)
			}
		}

		if assert.Len(s.T(), retrievedLogEntries, 1) {
			assert.NotEmpty(s.T(), retrievedLogEntries[0].GetRawData())
			assert.Contains(s.T(), string(retrievedLogEntries[0].GetRawData()), "test message")
		}
	}
}

func (s *ServerTestSuite) TestWriteLogsWithError() {
	s.configuration.Telemetry.Logs.BufferCapacity = 1
	server := telemetry.NewServer()
	assert.NoError(s.T(), server.Reconfigure(s.configuration))
	go func() {
		_ = server.ServeSession(s.ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	// Logger setup with custom hook to capture log events, thanks to https://stackoverflow.com/a/76851955
	logger := zerolog.New(server)
	logHook := &logHook{}
	logger = logger.Hook(logHook)
	errorsList := make([]error, 0)
	zerolog.ErrorHandler = func(err error) {
		errorsList = append(errorsList, err)
	}

	logger.Info().Msg("first message")
	logger.Info().Msg("second message")

	if assert.Len(s.T(), logHook.logEvents, 2) {
		if assert.Len(s.T(), errorsList, 1) {
			assert.ErrorIs(s.T(), errorsList[0], telemetry.ErrFullBuffer)
		}
	}
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
