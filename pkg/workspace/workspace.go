package workspace

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"os"
	"path"
	"path/filepath"
)

const (
	FolderPerms = 0755
	FilePerms   = 0644

	TempFolderPrefix = "harkonnen_"
	ScriptsFolder    = "scripts"
	RunsFolder       = "runs"
	VariablesFolder  = "variables"
	DataFolder       = "data"

	ManifestFile        = "harkonnen.yaml"
	GlobalVariablesFile = "globals.yaml"
)

type Workspace struct {
	location   string
	currentRun *Run
	logger     zerolog.Logger
}

func New(location string) Workspace {
	work := Workspace{
		location: location,
		logger:   zerolog.Nop(),
	}
	return work
}

func NewTemp() Workspace {
	return New("")
}

func (w *Workspace) SetLogger(logger zerolog.Logger) {
	w.logger = logger.With().Str(log.ComponentKey, "workspace").Logger()
}

func (w *Workspace) Hydrate() error {
	if err := w.EnsureWorkspace(); err != nil {
		return err
	}

	if err := w.createManifestFile(); err != nil {
		return err
	}

	if err := w.createGlobalVariablesFile(); err != nil {
		return err
	}

	if err := w.createVariablesFolder(); err != nil {
		return err
	}

	if err := w.createRunsFolder(); err != nil {
		return err
	}

	if err := w.createScriptsFolder(); err != nil {
		return err
	}

	return nil
}

func (w *Workspace) EnsureWorkspace() error {
	// If location is unset, create a temporary folder and set it as workspace location
	if w.location == "" {
		dir, err := os.MkdirTemp("", TempFolderPrefix)
		if err != nil {
			return err
		}
		w.location = dir
		w.logger.Info().Str("location", w.Location()).Msg("Temporary workspace created")
		return nil
	}

	// If the folder is set, but does not exist, create it
	if _, err := os.Stat(w.location); os.IsNotExist(err) {
		if err := os.MkdirAll(w.location, FolderPerms); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (w *Workspace) createManifestFile() error {
	configFilePath := path.Join(w.location, ManifestFile)
	return os.WriteFile(configFilePath, []byte(configuration.DefaultManifestContent), FilePerms)
}

func (w *Workspace) createGlobalVariablesFile() error {
	file, err := os.Create(path.Join(w.location, GlobalVariablesFile))
	if err != nil {
		return err
	}
	return file.Close()
}

func (w *Workspace) createVariablesFolder() error {
	return os.MkdirAll(path.Join(w.location, VariablesFolder), FolderPerms)
}

func (w *Workspace) createScriptsFolder() error {
	if err := os.MkdirAll(path.Join(w.location, ScriptsFolder), FolderPerms); err != nil {
		return err
	}

	pipelinePath, err := w.PipelinePath()
	if err != nil {
		return err
	}

	return os.WriteFile(pipelinePath, []byte(DefaultPipelineContent), FilePerms)
}

func (w *Workspace) createRunsFolder() error {
	return os.MkdirAll(path.Join(w.location, RunsFolder), FolderPerms)
}

func (w *Workspace) createDataFolder() error {
	return os.MkdirAll(path.Join(w.location, DataFolder), FolderPerms)
}

func (w *Workspace) DeleteContent() error {
	// Snippet taken from https://stackoverflow.com/a/33451503
	// to delete only the content of a folder, not the folder itself
	workingFolderContent, err := os.Open(w.location)
	if err != nil {
		return err
	}

	defer func() {
		_ = workingFolderContent.Close()
	}()

	names, err := workingFolderContent.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(w.location, name))
		if err != nil {
			return err
		}
	}

	w.logger.Info().Str("location", w.Location()).Msg("Workspace content cleaned")

	return nil
}

func (w *Workspace) Delete() error {
	oldLocation := w.Location()
	if err := os.RemoveAll(w.location); err != nil {
		return err
	}
	w.location = ""
	w.logger.Info().Str("location", oldLocation).Msg("Workspace deleted")
	return nil
}

func (w *Workspace) CompressToArchive() ([]byte, error) {
	archiver := ZipArchive{}
	return archiver.Archive(w.location)
}

func (w *Workspace) ExtractFromArchive(archiveContent []byte) error {
	extractor := ZipArchive{}
	return extractor.Extract(archiveContent, w.location)
}

func (w *Workspace) Location() string {
	return w.location
}

func (w *Workspace) Configuration() (*configuration.Manifest, error) {
	config, err := configuration.New(path.Join(w.location, ManifestFile))
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (w *Workspace) PipelinePath() (string, error) {
	config, err := w.Configuration()
	if err != nil {
		return "", err
	}
	return path.Join(w.location, ScriptsFolder, config.Load.Script), nil
}

func (w *Workspace) Pipeline() ([]byte, string, error) {
	pipelinePath, err := w.PipelinePath()
	if err != nil {
		return nil, "", err
	}

	pipelineContent, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, "", err
	}

	return pipelineContent, pipelinePath, nil
}

func (w *Workspace) Runs() ([]string, error) {
	entries, err := os.ReadDir(w.location)
	if err != nil {
		return nil, err
	}

	var runs []string

	for _, entry := range entries {
		if entry.IsDir() {
			runs = append(runs, entry.Name())
		}
	}
	return runs, nil
}

func (w *Workspace) CreateRun(ctx context.Context, name string) error {
	run, err := NewRun(ctx, name, w.location)
	if err != nil {
		return err
	}

	w.currentRun = run
	return nil
}

func (w *Workspace) SwitchRun(name string) error {
	if _, err := os.Stat(filepath.Join(w.location, name)); os.IsNotExist(err) {
		return err
	}

	w.currentRun = LoadRun(name, w.location)
	return nil
}

func (w *Workspace) DeleteRun(name string) error {
	if w.currentRun != nil && w.currentRun.name == name {
		w.currentRun = nil
	}
	return os.RemoveAll(path.Join(w.location, name))
}

func (w *Workspace) CurrentRun() *Run {
	return w.currentRun
}
