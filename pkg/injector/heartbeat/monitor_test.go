package heartbeat_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewHeartbeatMonitor(t *testing.T) {
	bh, err := heartbeat.NewMonitor(5*time.Second, 10*time.Second)
	if assert.NoError(t, err) {
		assert.IsType(t, &heartbeat.Monitor{}, bh)
	}
}

func TestNewHeartbeatMonitorWithInvalidInput(t *testing.T) {
	_, err := heartbeat.NewMonitor(10*time.Second, 5*time.Second)
	assert.Error(t, err)
}

func TestNewHeartbeatMonitorWithInvalidInput2(t *testing.T) {
	_, err := heartbeat.NewMonitor(5*time.Second, 5*time.Second)
	assert.Error(t, err)
}

func TestNewHeartbeatMonitorWithoutBeatTimeout(t *testing.T) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	tempCtx, cancelFunc := context.WithCancel(context.TODO())
	ctx := logger.WithContext(tempCtx)

	bh, _ := heartbeat.NewMonitor(10*time.Millisecond, 20*time.Millisecond)
	childCtx := bh.Start(ctx)
	childContextCanceled := false

	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				bh.IncomingBeat()

			case <-ctx.Done():
				return
			}
		}
	}()

	statusChangeFuncWG := sync.WaitGroup{}
	go func() {
		statusChangeFuncWG.Add(1)
		<-childCtx.Done()
		logger.Info().Msg("Child context canceled, changing variable status")
		childContextCanceled = true
		// logger.Info().Bool("childContextCanceled", childContextCanceled).Msg("Variable status")
		statusChangeFuncWG.Done()
	}()

	time.Sleep(50 * time.Millisecond)
	assert.False(t, childContextCanceled)
	cancelFunc()
	bh.Wait()
	statusChangeFuncWG.Wait()
	assert.True(t, childContextCanceled)
}

func TestNewHeartbeatMonitorWithBeatTimeout(t *testing.T) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	tempCtx, cancelFunc := context.WithCancel(context.TODO())
	ctx := logger.WithContext(tempCtx)

	bh, _ := heartbeat.NewMonitor(10*time.Millisecond, 20*time.Millisecond)
	childCtx := bh.Start(ctx)
	childContextCanceled := false

	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				bh.IncomingBeat()

			case <-ctx.Done():
				return
			}
		}
	}()

	statusChangeFuncWG := sync.WaitGroup{}
	go func() {
		statusChangeFuncWG.Add(1)
		<-childCtx.Done()
		logger.Info().Msg("Child context canceled, changing variable status")
		childContextCanceled = true
		// logger.Info().Bool("childContextCanceled", childContextCanceled).Msg("Variable status")
		statusChangeFuncWG.Done()
	}()

	time.Sleep(50 * time.Millisecond)
	cancelFunc()
	assert.True(t, childContextCanceled)
}

func TestExternalStop(t *testing.T) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	ctx := logger.WithContext(context.TODO())

	bh, _ := heartbeat.NewMonitor(10*time.Millisecond, 20*time.Millisecond)
	childCtx := bh.Start(ctx)
	childContextCanceled := false

	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				bh.IncomingBeat()

			case <-ctx.Done():
				return
			}
		}
	}()

	statusChangeFuncWG := sync.WaitGroup{}
	go func() {
		statusChangeFuncWG.Add(1)
		<-childCtx.Done()
		logger.Info().Msg("Child context canceled, changing variable status")
		childContextCanceled = true
		// logger.Info().Bool("childContextCanceled", childContextCanceled).Msg("Variable status")
		statusChangeFuncWG.Done()
	}()

	time.Sleep(50 * time.Millisecond)
	assert.False(t, childContextCanceled)
	bh.Stop()
	bh.Wait()
	statusChangeFuncWG.Wait()
	assert.True(t, childContextCanceled)
}
