package telemetry_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
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
	server := telemetry.NewServer(s.ctx, s.configuration)
	assert.IsType(s.T(), &telemetry.Server{}, server)
}

func (s *ServerTestSuite) TestStoreSample_NoError() {
	server := telemetry.NewServer(s.ctx, s.configuration)
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
	server := telemetry.NewServer(s.ctx, s.configuration)
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
	server := telemetry.NewServer(s.ctx, s.configuration)
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
	server := telemetry.NewServer(s.ctx, s.configuration)
	transaction := &telemetry.Transaction{}
	err := server.StoreTransaction(transaction)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreTransaction_Error() {
	s.configuration.Telemetry.LoadMetrics.BufferCapacity = 1
	server := telemetry.NewServer(s.ctx, s.configuration)
	transaction := &telemetry.Transaction{}
	if assert.NoError(s.T(), server.StoreTransaction(transaction)) {
		err := server.StoreTransaction(transaction)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreTransaction_Retrieve() {
	server := telemetry.NewServer(s.ctx, s.configuration)
	transaction := &telemetry.Transaction{}

	// GRPC setup
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
	server := telemetry.NewServer(s.ctx, s.configuration)
	counters := &telemetry.IterationCounters{}
	err := server.StoreIterationCounters(counters)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreIterationCounters_Error() {
	s.configuration.Telemetry.LoadMetrics.BufferCapacity = 1
	server := telemetry.NewServer(s.ctx, s.configuration)
	counters := &telemetry.IterationCounters{}
	if assert.NoError(s.T(), server.StoreIterationCounters(counters)) {
		err := server.StoreIterationCounters(counters)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreIterationCounters_Retrieve() {
	server := telemetry.NewServer(s.ctx, s.configuration)
	counters := &telemetry.IterationCounters{}

	// GRPC setup
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
	server := telemetry.NewServer(s.ctx, s.configuration)
	hostMetrics := &telemetry.HostMetrics{}
	err := server.StoreHostMetrics(hostMetrics)
	assert.NoError(s.T(), err)
}

func (s *ServerTestSuite) TestStoreHostMetrics_Error() {
	s.configuration.Telemetry.HostMetrics.BufferCapacity = 1
	server := telemetry.NewServer(s.ctx, s.configuration)
	hostMetrics := &telemetry.HostMetrics{}
	if assert.NoError(s.T(), server.StoreHostMetrics(hostMetrics)) {
		err := server.StoreHostMetrics(hostMetrics)
		if assert.Error(s.T(), err) {
			assert.ErrorIs(s.T(), err, telemetry.ErrFullBuffer)
		}
	}
}

func (s *ServerTestSuite) TestStoreHostMetrics_Retrieve() {
	server := telemetry.NewServer(s.ctx, s.configuration)
	hostMetrics := &telemetry.HostMetrics{}

	// GRPC setup
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

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
