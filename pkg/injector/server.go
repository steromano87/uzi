package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/heartbeat"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"os"
	"path/filepath"
	"time"
)

// Global configuration
var configuration = workingfolder.MustNewDefault()

type Server struct {
	*heartbeat.Server
	UnimplementedInjectorServer

	spawner           *Spawner
	spawnerContext    context.Context
	spawnerCancelFunc context.CancelCauseFunc

	mainCtx context.Context

	logger        *zerolog.Logger
	workingFolder string

	heartbeatTimeoutTimer *time.Timer
}

func NewServer() *Server {
	server := new(Server)

	return server
}

func (s *Server) Run(ctx context.Context, heartbeatInterval time.Duration, heartbeatTimeout time.Duration) error {
	s.logger = zerolog.Ctx(ctx)
	bh, heartbeatMonitoredCtx, err := heartbeat.NewServer(ctx, heartbeatInterval, heartbeatTimeout)
	if err != nil {
		return err
	}

	s.mainCtx = heartbeatMonitoredCtx
	s.Server = bh
	s.spawnerContext, s.spawnerCancelFunc = context.WithCancelCause(ctx)

	err = s.initWorkingFolder()
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			s.contextualizedLogger().Info().Msg("Starting server shutdown")
			s.spawnerCancelFunc(context.Cause(ctx))
			err := s.deleteWorkingFolder()
			if err != nil {
				s.contextualizedLogger().Error().Err(err).Msg("Error while shutting down")
			}
			s.contextualizedLogger().Info().Msg("Shutdown completed")
			return
		}
	}()

	return nil
}

func (s *Server) WorkingFolder() string {
	return s.workingFolder
}

func (s *Server) initWorkingFolder() error {
	dir, err := os.MkdirTemp("", "harkonnen_")
	if err != nil {
		return err
	}
	s.workingFolder = dir
	s.contextualizedLogger().Info().Str("workingFolder", s.workingFolder).Msg("Created temporary working folder")
	return err
}

func (s *Server) resetWorkingFolder() error {
	s.contextualizedLogger().Info().Str("workingFolder", s.workingFolder).Msg("Cleaning content of temporary working folder")

	// Snippet taken from https://stackoverflow.com/a/33451503
	// to delete only the content of a folder, not the folder itself
	workingFolderContent, err := os.Open(s.workingFolder)
	if err != nil {
		s.contextualizedLogger().Error().Err(err).Str(
			"workingFolder", s.workingFolder,
		).Msg("Error cleaning temporary working folder content")
		return err
	}

	defer func() {
		err := workingFolderContent.Close()
		if err != nil {
			s.contextualizedLogger().Error().Err(err).Msg("Error during temporary folder cleanup")
		}
	}()

	names, err := workingFolderContent.Readdirnames(-1)
	if err != nil {
		s.contextualizedLogger().Error().Err(err).Str(
			"workingFolder", s.workingFolder,
		).Msg("Error cleaning temporary working folder content")
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(s.workingFolder, name))
		if err != nil {
			s.contextualizedLogger().Error().Err(err).Str(
				"workingFolder", s.workingFolder,
			).Msg("Error cleaning temporary working folder content")
			return err
		}
	}
	s.contextualizedLogger().Info().Str("workingFolder", s.workingFolder).Msg("Temporary folder content cleared")
	return nil
}

func (s *Server) deleteWorkingFolder() error {
	s.contextualizedLogger().Info().Str("workingFolder", s.workingFolder).Msg("Deleting temporary working folder")
	err := os.RemoveAll(s.workingFolder)
	if err != nil {
		s.contextualizedLogger().Error().Err(err).Str(
			"workingFolder", s.workingFolder,
		).Msg("Error cleaning temporary working folder")
		return err
	}
	s.contextualizedLogger().Info().Str("workingFolder", s.workingFolder).Msg("Temporary working folder deleted")

	return nil
}

func (s *Server) contextualizedLogger() *zerolog.Logger {
	logger := s.logger.With().Str("component", "InjectorConfiguration Server").Logger()
	return &logger
}
