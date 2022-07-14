package loading

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/io"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"runtime"
)

// L is, like T for the testing package, the object holding the state of the load test.
// Each instance of L is assigned to a specific script runner
type L struct {
	Logger       *zerolog.Logger
	Config       *project.Config
	Variables    Variables
	SampleWriter io.AnyWriter[model.Sample]
}

func (l L) Render(tpl string) (string, error) {
	return l.Variables.Render(tpl)
}

func (l L) OnNewSample(sample model.Sample) {
	l.contextLogger().Trace().Str("sampleName", sample.Name).Msg("New sample generated")
	err := l.SampleWriter.Write(sample)

	if err != nil {
		l.OnError(err)
	}
}

func (l L) OnError(err error) {
	l.contextLogger().Error().Stack().Err(err)
}

func (l L) FailIteration(err error) {
	l.OnError(err)
	runtime.Goexit()
}

func (l L) OnUnrecoverableError(err error) {
	l.contextLogger().Panic().Stack().Err(err).Msg("Caught an unrecoverable error")
}

func (l L) OnIterationStart(counter int) {
	l.contextLogger().Debug().Int("iteration", counter).Msg("Started new iteration")
}

func (l L) OnIterationEnd(counter int) {
	l.contextLogger().Debug().Int("iteration", counter).Msg("Iteration ended")
}

func (l L) contextLogger() *zerolog.Logger {
	contextLogger := l.Logger.With().Str("component", "Runtime").Logger()
	return &contextLogger
}
