package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"reflect"
	"sync"
	"time"
)

type BeatingHeart struct {
	logger *zerolog.Logger

	beatInterval     time.Duration
	beatTimeout      time.Duration
	beatTicker       *time.Ticker
	beatTimeoutTimer *time.Timer

	incomingBeatChan chan struct{}
	outgoingBeatChan chan struct{}

	beatMonitorsWG sync.WaitGroup

	childCancelCauseFunc context.CancelCauseFunc
}

func NewBeatingHeart(beatInterval, beatTimeout time.Duration) (*BeatingHeart, error) {
	if beatTimeout <= beatInterval {
		return nil, errors.New(fmt.Sprintf("beat timeout (%s) cannot be smaller than beat interval (%s)", beatTimeout, beatInterval))
	}

	beatingHeart := new(BeatingHeart)
	beatingHeart.beatInterval = beatInterval
	beatingHeart.beatTimeout = beatTimeout
	beatingHeart.incomingBeatChan = make(chan struct{})
	beatingHeart.outgoingBeatChan = make(chan struct{})

	return beatingHeart, nil
}

func (bh *BeatingHeart) Start(ctx context.Context) context.Context {
	bh.logger = zerolog.Ctx(ctx)
	childCtx, childCancelCauseFunc := context.WithCancelCause(ctx)
	bh.childCancelCauseFunc = childCancelCauseFunc

	bh.beatTicker = time.NewTicker(bh.beatInterval)
	bh.beatTimeoutTimer = time.NewTimer(bh.beatTimeout)

	go bh.incomingBeatMonitorLoop(ctx)
	go bh.outgoingBeatSendLoop(childCtx)

	return childCtx
}

func (bh *BeatingHeart) Wait() {
	bh.beatMonitorsWG.Wait()
}

func (bh *BeatingHeart) IncomingBeat() {
	bh.incomingBeatChan <- struct{}{}
}

func (bh *BeatingHeart) OutgoingBeat() <-chan struct{} {
	return bh.outgoingBeatChan
}

func (bh *BeatingHeart) incomingBeatMonitorLoop(ctx context.Context) {
	bh.beatMonitorsWG.Add(1)
	defer bh.beatMonitorsWG.Done()

	for {
		select {
		case <-bh.incomingBeatChan:
			bh.contextualizedLogger().Debug().Msg("Received incoming heartbeat")
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
			bh.childCancelCauseFunc(context.Cause(ctx))

			return

		case <-bh.beatTimeoutTimer.C:
			bh.contextualizedLogger().Error().Msg("Incoming heartbeat timeout, canceling child context")
			bh.childCancelCauseFunc(errors.New("incoming heartbeat timeout"))
			return
		}
	}
}

func (bh *BeatingHeart) outgoingBeatSendLoop(ctx context.Context) {
	bh.beatMonitorsWG.Add(1)
	defer bh.beatMonitorsWG.Done()

	for {
		select {
		case <-bh.beatTicker.C:
			bh.contextualizedLogger().Debug().Msg("Sending heartbeat")
			// Non-blocking channel send, do not use a goroutine or it may leak
			reflect.ValueOf(bh.outgoingBeatChan).TrySend(reflect.ValueOf(struct{}{}))

		case <-ctx.Done():
			bh.contextualizedLogger().Info().Err(context.Cause(ctx)).Msg(
				"Context canceled, exiting outgoing beats send loop")
			return
		}
	}
}

func (bh *BeatingHeart) contextualizedLogger() *zerolog.Logger {
	logger := bh.logger.With().Str("component", "Beating Heart").Logger()
	return &logger
}
