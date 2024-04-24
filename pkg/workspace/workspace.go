package workspace

import (
	"context"
	"errors"
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

	UnimplementedWorkspaceServer
}

func New(location string) Workspace {
	work := Workspace{
		location: location,
	}
	return work
}

func NewTemp() Workspace {
	return New("")
}

func (w *Workspace) Hydrate() error {
	switch w.location {
	case "":
		dir, err := os.MkdirTemp("", TempFolderPrefix)
		if err != nil {
			return err
		}
		w.location = dir
	default:
		if err := os.MkdirAll(w.location, FolderPerms); err != nil {
			return err
		}
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

	return nil
}

func (w *Workspace) createManifestFile() error {
	configFilePath := path.Join(w.location, ManifestFile)
	return os.WriteFile(configFilePath, []byte(DefaultManifestContent), FilePerms)
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

	return nil
}

func (w *Workspace) Delete() error {
	return os.RemoveAll(w.location)
}

func (w *Workspace) ExtractFromArchive(archiveContent []byte, algorithm CompressionAlgorithm) error {
	switch algorithm {
	case CompressionAlgorithm_ZIP:
		extractor := ZipArchive{}
		return extractor.Extract(archiveContent, w.location)
	default:
		return errors.New("unknown compression algorithm: " + algorithm.String())
	}
}

func (w *Workspace) Location() string {
	return w.location
}

func (w *Workspace) Configuration() (*Configuration, error) {
	config, err := NewConfiguration(path.Join(w.location, ManifestFile))
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

func (w *Workspace) RawPipeline() ([]byte, error) {
	pipelinePath, err := w.PipelinePath()
	if err != nil {
		return nil, err
	}

	pipelineContent, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, err
	}

	return pipelineContent, nil
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
