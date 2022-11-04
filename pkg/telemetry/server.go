package telemetry

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
)

//go:generate sh -c "protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto"

type Server struct {
	UnimplementedTelemetryServer

	*LogCollector
	*SampleCollector
	*HostMetricsCollector
	*TransactionCollector
	*IterationCountersCollector
}

func NewServer(config *configuration.Configuration) *Server {
	server := new(Server)
	server.LogCollector = NewLogCollector()
	server.SampleCollector = NewSampleCollector()
	server.HostMetricsCollector = NewHostMetricsCollector(
		config.Telemetry.HostMetrics.PollInterval,
		config.Telemetry.HostMetrics.MeasureInterval,
	)
	server.TransactionCollector = NewTransactionCollector()
	server.IterationCountersCollector = NewIterationCountersCollector(config.Load.MaxIterations)

	return server
}

func (s *Server) GetSamples(_ *SampleRequest, server Telemetry_GetSamplesServer) error {
	samplesToSend := s.SampleCollector.GetSamples()
	for _, sample := range samplesToSend {
		if err := server.Send(sample); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) GetLogs(_ *LogRequest, server Telemetry_GetLogsServer) error {
	logsToSend := s.LogCollector.GetLogs()
	for _, log := range logsToSend {
		if err := server.Send(log); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) GetTransactions(_ *TransactionRequest, server Telemetry_GetTransactionsServer) error {
	transactionsToSend := s.TransactionCollector.GetTransactions()
	for _, transaction := range transactionsToSend {
		if err := server.Send(transaction); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) GetHostMetrics(_ *HostMetricsRequest, server Telemetry_GetHostMetricsServer) error {
	hostMetricsToSend := s.HostMetricsCollector.GetHostMetrics()
	for _, hostMetrics := range hostMetricsToSend {
		if err := server.Send(hostMetrics); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) GetIterationCounters(_ context.Context, _ *IterationCountersRequest) (*IterationCounters, error) {
	return s.IterationCountersCollector.GetCounters(), nil
}
