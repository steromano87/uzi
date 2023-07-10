package injector

import (
	"context"
	"errors"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//////////////////////////
// GRPC implementation  //
//////////////////////////

func (s *Server) Initialize(ctx context.Context, request *InitializationRequest) (*emptypb.Empty, error) {
	statusResponse, err := s.GetStatus(ctx, &heartbeat.StatusRequest{})
	if err != nil {
		return nil, err
	}

	lockStatus := statusResponse.GetStatus()

	if lockStatus != heartbeat.LockStatus_LOCKED {
		err := errors.New("injector must be in LOCKED status before initializing it")
		s.contextualizedLogger().Error().Err(err).Msg("InjectorConfiguration cannot accept initialization request")
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	resultChan := make(chan error)

	// Working folder unzipping
	s.contextualizedLogger().Info().Msg("Received initialization request")
	s.contextualizedLogger().Info().Msg("Unzipping working folder...")
	go func() {
		workingFolder := request.GetWorkingFolder()
		resultChan <- utils.UnzipFolder(workingFolder.GetCompressedWorkingFolder(), s.WorkingFolder())
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			s.contextualizedLogger().Error().Err(err).Msg("Error unzipping the working folder")
			return nil, status.Error(codes.Unknown, err.Error())
		}
		s.contextualizedLogger().Info().Msg("Working folder unzip completed")

	case <-ctx.Done():
		s.contextualizedLogger().Warn().Err(context.Cause(ctx)).Msg("Initialization interrupted due to context cancellation")
		return nil, status.Error(codes.Canceled, context.Cause(ctx).Error())
	}

	// Configuration setup
	s.contextualizedLogger().Info().Msg("Parsing configuration")
	go func() {
		var err error
		configuration, err = workingfolder.ParseConfiguration(s.WorkingFolder())
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			s.contextualizedLogger().Error().Err(err).Msg("Error while parsing configuration")
			return nil, status.Error(codes.Unknown, err.Error())
		}
		s.contextualizedLogger().Info().Msg("Configuration successfully parsed")

	case <-ctx.Done():
		s.contextualizedLogger().Warn().Err(context.Cause(ctx)).Msg("Initialization interrupted due to context cancellation")
		return nil, status.Error(codes.Canceled, context.Cause(ctx).Error())
	}

	// Pipeline parse
	s.contextualizedLogger().Info().Str("pipelinePath", configuration.Load.Script).Msg("Parsing load pipeline")
	var referencePipeline pipeline.Pipeline
	go func() {
		var err error
		pipContent, pipFile, err := workingfolder.ParsePipeline(s.WorkingFolder())
		if err != nil {
			resultChan <- err
			return
		}

		referencePipeline, err = pipeline.Decode(pipContent, pipFile)
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			s.contextualizedLogger().Error().Err(err).Msg("Error while parsing load pipeline")
			return nil, status.Error(codes.Unknown, err.Error())
		}
		s.contextualizedLogger().Info().Msg("LoadConfiguration pipeline successfully parsed")

	case <-ctx.Done():
		s.contextualizedLogger().Warn().Err(context.Cause(ctx)).Msg("Initialization interrupted due to context cancellation")
		return nil, status.Error(codes.Canceled, context.Cause(ctx).Error())
	}

	// Spawner initialization
	s.contextualizedLogger().Info().Uint64("maxUsersQuota", request.GetMaxUsersQuota()).Msg("Initializing spawner")
	go func() {
		s.spawner = NewSpawner(request.GetMaxUsersQuota())
		resultChan <- nil
	}()

	select {
	case <-resultChan:
		s.spawner.Start(s.spawnerContext, referencePipeline)
		s.contextualizedLogger().Info().Msg("Spawner initialized and started")
		return nil, nil

	case <-ctx.Done():
		s.contextualizedLogger().Warn().Err(context.Cause(ctx)).Msg("Initialization interrupted due to context cancellation")
		return nil, context.Cause(ctx)
	}
}
