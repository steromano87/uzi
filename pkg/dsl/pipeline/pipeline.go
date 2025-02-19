package pipeline

import (
	"context"
	"errors"
	"github.com/steromano87/uzi/v1/pkg/dsl"
	"github.com/steromano87/uzi/v1/pkg/telemetry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync/atomic"
	"time"
)

type Pipeline struct {
	Setup    StepContainer
	Main     StepContainer
	Teardown StepContainer

	IterationCounter
	gracefulShutdownRequested atomic.Bool
}

func Nop() *Pipeline {
	return &Pipeline{
		Setup:    StepContainer{steps: make([]dsl.Step, 0)},
		Main:     StepContainer{steps: make([]dsl.Step, 0)},
		Teardown: StepContainer{steps: make([]dsl.Step, 0)},
	}
}

func (p *Pipeline) RunSetup(ctx dsl.Context) error {
	start := time.Now()
	err := p.Setup.Run(ctx)
	end := time.Now()

	transaction := &telemetry.Transaction{
		SyntheticUserId: ctx.SynthUserId,
		Name:            "Setup",
		Start:           timestamppb.New(start),
		End:             timestamppb.New(end),
		Duration:        durationpb.New(end.Sub(start)),
		Global:          true,
		Successful:      err == nil,
	}

	defer func() {
		if err := ctx.StoreTransaction(transaction); err != nil {
			ctx.Logger.Error().Err(err).Msg("Encountered an error while saving setup transaction")
		}
	}()

	return err
}

func (p *Pipeline) RunMain(ctx dsl.Context) error {
	// Run the main block until a graceful shutdown is requested
	for !p.gracefulShutdownRequested.Load() && !p.IterationCounter.MaxIterationsReached() {
		p.IterationCounter.AddInProgressIteration()
		err := p.RunMainOnce(ctx)

		switch {
		case errors.Is(err, context.Canceled):
			return err

		case err == nil:
			p.IterationCounter.AddPassedIteration()

		default:
			p.IterationCounter.AddFailedIteration()
		}
	}

	ctx.Logger.Debug().Msg("Gracefully exited main loop")
	p.gracefulShutdownRequested.Store(false)
	return nil
}

func (p *Pipeline) RunMainOnce(ctx dsl.Context) error {
	start := time.Now()
	err := p.Main.Run(ctx)
	end := time.Now()

	transaction := &telemetry.Transaction{
		SyntheticUserId: ctx.SynthUserId,
		Name:            "Main",
		Start:           timestamppb.New(start),
		End:             timestamppb.New(end),
		Duration:        durationpb.New(end.Sub(start)),
		Global:          true,
		Successful:      err == nil,
	}

	defer func() {
		if err := ctx.StoreTransaction(transaction); err != nil {
			ctx.Logger.Error().Err(err).Msg("Encountered an error while saving setup transaction")
		}
	}()

	return err
}

func (p *Pipeline) RunTeardown(ctx dsl.Context) error {
	start := time.Now()
	err := p.Teardown.Run(ctx)
	end := time.Now()

	transaction := &telemetry.Transaction{
		SyntheticUserId: ctx.SynthUserId,
		Name:            "Teardown",
		Start:           timestamppb.New(start),
		End:             timestamppb.New(end),
		Duration:        durationpb.New(end.Sub(start)),
		Global:          true,
		Successful:      err == nil,
	}

	defer func() {
		if err := ctx.StoreTransaction(transaction); err != nil {
			ctx.Logger.Error().Err(err).Msg("Encountered an error while saving teardown transaction")
		}
	}()

	return err
}

func (p *Pipeline) RequestGracefulShutdown() {
	p.gracefulShutdownRequested.Store(true)
}

func (p *Pipeline) Clone() *Pipeline {
	return &Pipeline{
		Setup:    p.Setup,
		Main:     p.Main,
		Teardown: p.Teardown,
	}
}
