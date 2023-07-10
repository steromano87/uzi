package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"reflect"
	"sync"
	"time"
)

const (
	ServerRole  = "server"
	ClientRole  = "client"
	GenericRole = "generic"
)

type Monitor struct {
	logger *zerolog.Logger
	role   string

	beatInterval     time.Duration
	beatTimeout      time.Duration
	beatTicker       *time.Ticker
	beatTimeoutTimer *time.Timer

	incomingBeatChan chan struct{}
	outgoingBeatChan chan struct{}

	beatMonitorsWG        sync.WaitGroup
	beatMonitorCtx        context.Context
	beatMonitorCancelFunc context.CancelFunc

	externalCtxCancelFunc context.CancelCauseFunc
}

func NewServerMonitor(beatInterval, beatTimeout time.Duration) (*Monitor, error) {
	monitor, err := NewMonitor(beatInterval, beatTimeout)
	if err != nil {
		return nil, err
	}

	monitor.role = ServerRole
	return monitor, nil
}

func NewClientMonitor(beatInterval, beatTimeout time.Duration) (*Monitor, error) {
	monitor, err := NewMonitor(beatInterval, beatTimeout)
	if err != nil {
		return nil, err
	}

	monitor.role = ClientRole
	return monitor, nil
}

func NewMonitor(beatInterval, beatTimeout time.Duration) (*Monitor, error) {
	if beatTimeout <= beatInterval {
		return nil, errors.New(fmt.Sprintf("beat timeout (%s) cannot be smaller than beat interval (%s)", beatTimeout, beatInterval))
	}

	monitor := new(Monitor)
	monitor.beatInterval = beatInterval
	monitor.beatTimeout = beatTimeout
	monitor.incomingBeatChan = make(chan struct{})
	monitor.outgoingBeatChan = make(chan struct{})
	monitor.role = GenericRole

	return monitor, nil
}

func (bh *Monitor) Start(ctx context.Context) context.Context {
	bh.logger = zerolog.Ctx(ctx)
	bh.beatMonitorCtx, bh.beatMonitorCancelFunc = context.WithCancel(ctx)

	outgoingBeatCtx, outgoingBeatCancelCauseFunc := context.WithCancelCause(bh.beatMonitorCtx)
	bh.externalCtxCancelFunc = outgoingBeatCancelCauseFunc

	bh.beatTicker = time.NewTicker(bh.beatInterval)
	bh.beatTimeoutTimer = time.NewTimer(bh.beatTimeout)

	go bh.incomingBeatMonitorLoop(bh.beatMonitorCtx)
	go bh.outgoingBeatSendLoop(outgoingBeatCtx)

	bh.contextualizedLogger().Info().Dur(
		"heartbeatInterval", bh.beatInterval,
	).Dur(
		"heartbeatTimeout", bh.beatTimeout,
	).Msg("Started heartbeat monitor")
	return outgoingBeatCtx
}

func (bh *Monitor) Stop() {
	bh.beatMonitorCancelFunc()
}

func (bh *Monitor) Wait() {
	bh.beatMonitorsWG.Wait()
}

func (bh *Monitor) IncomingBeat() {
	bh.incomingBeatChan <- struct{}{}
}

func (bh *Monitor) OutgoingBeat() <-chan struct{} {
	return bh.outgoingBeatChan
}

func (bh *Monitor) incomingBeatMonitorLoop(ctx context.Context) {
	bh.beatMonitorsWG.Add(1)
	defer bh.beatMonitorsWG.Done()

	for {
		select {
		case <-bh.incomingBeatChan:
			bh.contextualizedLogger().Trace().Msg("Received heartbeat")
			if !bh.beatTimeoutTimer.Stop() {
				<-bh.beatTimeoutTimer.C
			}

			bh.beatTimeoutTimer.Reset(bh.beatTimeout)

		case <-ctx.Done():
			bh.contextualizedLogger().Info().Err(context.Cause(ctx)).Msg(
				"Context canceled, exiting incoming beats monitor loop")
			if !bh.beatTimeoutTimer.Stop() {
				<-bh.beatTimeoutTimer.C
			}
			bh.externalCtxCancelFunc(context.Cause(ctx))

			return

		case <-bh.beatTimeoutTimer.C:
			bh.contextualizedLogger().Error().Msg("Incoming heartbeat timeout, canceling child context")
			bh.externalCtxCancelFunc(ErrHeartbeatRequestTimeout)
			return
		}
	}
}

func (bh *Monitor) outgoingBeatSendLoop(ctx context.Context) {
	bh.beatMonitorsWG.Add(1)
	defer bh.beatMonitorsWG.Done()

	for {
		select {
		case <-bh.beatTicker.C:
			bh.contextualizedLogger().Trace().Msg("Sending heartbeat")
			// Non-blocking channel send, do not use a goroutine, or it may leak
			reflect.ValueOf(bh.outgoingBeatChan).TrySend(reflect.ValueOf(struct{}{}))

		case <-ctx.Done():
			bh.contextualizedLogger().Info().Err(context.Cause(ctx)).Msg(
				"Context canceled, exiting outgoing beats send loop")
			return
		}
	}
}

func (bh *Monitor) contextualizedLogger() *zerolog.Logger {
	logger := bh.logger.With().Str("component", "Heartbeat Monitor").Str("role", bh.role).Logger()
	return &logger
}
