package heartbeat

import (
	"context"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type GrpcClient struct {
	logger *zerolog.Logger
	HeartbeatClient
	monitor *Monitor
}

func NewGrpcClient(grpcInterface grpc.ClientConnInterface) *GrpcClient {
	client := new(GrpcClient)
	client.HeartbeatClient = NewHeartbeatClient(grpcInterface)

	return client
}

func (c *GrpcClient) Monitor(ctx context.Context, beatInterval, beatTimeout time.Duration) error {
	c.logger = zerolog.Ctx(ctx)
	c.contextualizedLogger().Info().Msg("Starting heartbeat monitor")
	monitor, err := NewClientMonitor(beatInterval, beatTimeout)
	if err != nil {
		c.contextualizedLogger().Error().Err(err).Msg("Error creating the heartbeat monitor")
		return err
	}
	c.monitor = monitor

	heartbeatStream, err := c.HeartbeatClient.Lock(ctx)
	if err != nil {
		c.contextualizedLogger().Error().Err(err).Msg("Error initializing control stream")
		return err
	}

	controlCtx := c.monitor.Start(ctx)

	incomingBeatChan := make(chan error)
	outgoingBeatChan := make(chan error)
	go func() {
		incomingBeatChan <- c.incomingBeatLoop(heartbeatStream)
	}()
	go func() {
		outgoingBeatChan <- c.outgoingBeatLoop(controlCtx, heartbeatStream)
	}()

	select {
	case err := <-incomingBeatChan:
		c.contextualizedLogger().Info().Err(err).Msg("Stopping lock loop due to incoming beat loop termination")
		return err

	case err := <-outgoingBeatChan:
		c.contextualizedLogger().Info().Err(err).Msg("Stopping lock loop due to outgoing beat loop termination")
		return err
	}
}

func (c *GrpcClient) incomingBeatLoop(heartbeatStream Heartbeat_LockClient) error {
	for {
		_, err := heartbeatStream.Recv()
		if err == io.EOF {
			c.contextualizedLogger().Info().Msg("Lock channel remotely closed, stopping...")
			return nil
		}

		// Check if the stream was remotely closed by checking the GRPC of the error code
		if status.Convert(err).Code() == codes.Canceled {
			c.contextualizedLogger().Info().Msg("Remote context gracefully canceled, stopping...")
			return nil
		}

		if err != nil {
			c.contextualizedLogger().Error().Err(err).Msg("Error when receiving heartbeat, exiting")
			return err
		}

		c.contextualizedLogger().Debug().Msg("Received heartbeat")
		c.monitor.IncomingBeat()
	}
}

func (c *GrpcClient) outgoingBeatLoop(controlCtx context.Context, heartbeatStream Heartbeat_LockClient) error {
	defer func() {
		c.contextualizedLogger().Info().Msg("Closing lock stream")
		err := heartbeatStream.CloseSend()
		if err != nil {
			c.contextualizedLogger().Error().Err(err).Msg("Error closing lock stream")
		}
	}()

	for {
		select {
		case <-controlCtx.Done():
			err := context.Cause(controlCtx)
			if err == ErrHeartbeatRequestTimeout {
				c.contextualizedLogger().Error().Err(err).Msg("Heartbeat timeout reached, exiting")
				return err
			}

			c.contextualizedLogger().Info().Err(err).Msg("Stop requested, exiting")
			return nil

		case <-c.monitor.OutgoingBeat():
			c.contextualizedLogger().Debug().Msg("Sending heartbeat")
			if err := heartbeatStream.Send(&HeartbeatRequest{Timestamp: timestamppb.Now()}); err != nil {
				return err
			}
		}
	}
}

func (c *GrpcClient) contextualizedLogger() *zerolog.Logger {
	logger := c.logger.With().Str("component", "Heartbeat GrpcClient").Logger()
	return &logger
}
