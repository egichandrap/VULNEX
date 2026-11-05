package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

const (
	stateFileName = "vulnex.json"
	filePerms     = 0600
	dirPerms      = 0700
)

// FileStorage implements atomic persistence for configuration and installation state.
type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// LoadConfig reads and deserializes an InstallConfig from the filesystem.
func (s *FileStorage) LoadConfig(ctx context.Context, path string) (*domain.InstallConfig, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, domain.WrapError(domain.ExitInvalidConfig, "failed to read config file", err)
	}

	var cfg domain.InstallConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, domain.WrapError(domain.ExitInvalidConfig, "failed to parse config JSON", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, domain.WrapError(domain.ExitInvalidConfig, "invalid configuration values", err)
	}

	return &cfg, nil
}

// SaveConfig atomically writes an InstallConfig with restricted 0600 permissions.
func (s *FileStorage) SaveConfig(ctx context.Context, path string, cfg *domain.InstallConfig) error {
	if cfg == nil {
		return domain.NewError(domain.ExitInvalidConfig, "cannot save nil configuration")
	}

	cleanPath := filepath.Clean(path)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, dirPerms); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to create config directory", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to serialize config JSON", err)
	}

	return s.atomicWriteFile(cleanPath, data)
}

// LoadState reads the active InstallationState from the deployment data directory.
func (s *FileStorage) LoadState(ctx context.Context, dataDir string) (*domain.InstallationState, error) {
	statePath := filepath.Join(filepath.Clean(dataDir), stateFileName)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, domain.WrapError(domain.ExitGeneralFailure, "failed to read installation state", err)
	}

	var state domain.InstallationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, domain.WrapError(domain.ExitGeneralFailure, "failed to parse installation state JSON", err)
	}

	return &state, nil
}

// SaveState atomically writes the InstallationState to dataDir/vulnex.json with 0600 permissions.
func (s *FileStorage) SaveState(ctx context.Context, dataDir string, state *domain.InstallationState) error {
	if state == nil {
		return domain.NewError(domain.ExitGeneralFailure, "cannot save nil installation state")
	}

	cleanDir := filepath.Clean(dataDir)
	if err := os.MkdirAll(cleanDir, dirPerms); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to create data directory", err)
	}

	statePath := filepath.Join(cleanDir, stateFileName)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to serialize state JSON", err)
	}

	return s.atomicWriteFile(statePath, data)
}

// atomicWriteFile writes data to a temporary file in the same directory, syncs, and renames.
func (s *FileStorage) atomicWriteFile(destPath string, data []byte) error {
	dir := filepath.Dir(destPath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-vulnex-*")
	if err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to create temporary file for atomic write", err)
	}
	tmpName := tmpFile.Name()

	// Ensure cleanup if rename fails
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if err := os.Chmod(tmpName, filePerms); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to set file permissions", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to write temporary file data", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to sync temporary file", err)
	}

	if err := tmpFile.Close(); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to close temporary file", err)
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to atomically replace target file", err)
	}

	return nil
}
