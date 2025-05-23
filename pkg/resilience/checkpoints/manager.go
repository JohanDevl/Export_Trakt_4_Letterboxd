package checkpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
)

// Checkpoint represents a saved state during operation
type Checkpoint struct {
	OperationID   string                 `json:"operation_id"`
	OperationType string                 `json:"operation_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Progress      float64                `json:"progress"`
	State         map[string]interface{} `json:"state"`
	NextStep      string                 `json:"next_step"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// Manager handles checkpoint operations
type Manager struct {
	checkpointDir string
	maxAge        time.Duration
}

// Config represents checkpoint manager configuration
type Config struct {
	CheckpointDir string
	MaxAge        time.Duration
}

// DefaultConfig returns default checkpoint configuration
func DefaultConfig() *Config {
	return &Config{
		CheckpointDir: "./checkpoints",
		MaxAge:        24 * time.Hour, // Keep checkpoints for 24 hours
	}
}

// NewManager creates a new checkpoint manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Create checkpoint directory if it doesn't exist
	if err := os.MkdirAll(config.CheckpointDir, 0755); err != nil {
		return nil, types.NewAppError(
			types.ErrFileSystem,
			"failed to create checkpoint directory",
			err,
		)
	}
	
	return &Manager{
		checkpointDir: config.CheckpointDir,
		maxAge:        config.MaxAge,
	}, nil
}

// Save saves a checkpoint to disk
func (m *Manager) Save(ctx context.Context, checkpoint *Checkpoint) error {
	if checkpoint.OperationID == "" {
		return types.NewAppError(
			types.ErrInvalidInput,
			"checkpoint operation ID cannot be empty",
			nil,
		)
	}
	
	// Set timestamp if not provided
	if checkpoint.Timestamp.IsZero() {
		checkpoint.Timestamp = time.Now()
	}
	
	// Marshal checkpoint to JSON
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return types.NewAppError(
			types.ErrProcessingFailed,
			"failed to marshal checkpoint",
			err,
		)
	}
	
	// Generate filename
	filename := fmt.Sprintf("checkpoint_%s.json", checkpoint.OperationID)
	filepath := filepath.Join(m.checkpointDir, filename)
	
	// Write to file
	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return types.NewAppError(
			types.ErrFileSystem,
			"failed to write checkpoint file",
			err,
		)
	}
	
	return nil
}

// Load loads a checkpoint from disk
func (m *Manager) Load(ctx context.Context, operationID string) (*Checkpoint, error) {
	if operationID == "" {
		return nil, types.NewAppError(
			types.ErrInvalidInput,
			"operation ID cannot be empty",
			nil,
		)
	}
	
	filename := fmt.Sprintf("checkpoint_%s.json", operationID)
	filepath := filepath.Join(m.checkpointDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, types.NewAppError(
			types.ErrDataMissing,
			"checkpoint not found",
			err,
		)
	}
	
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, types.NewAppError(
			types.ErrFileSystem,
			"failed to read checkpoint file",
			err,
		)
	}
	
	// Unmarshal checkpoint
	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, types.NewAppError(
			types.ErrDataCorrupted,
			"failed to unmarshal checkpoint",
			err,
		)
	}
	
	// Check if checkpoint is expired
	if time.Since(checkpoint.Timestamp) > m.maxAge {
		// Delete expired checkpoint
		if deleteErr := m.Delete(ctx, operationID); deleteErr != nil {
			// Log deletion error but don't fail the load operation
		}
		
		return nil, types.NewAppError(
			types.ErrDataMissing,
			"checkpoint has expired",
			nil,
		)
	}
	
	return &checkpoint, nil
}

// Delete removes a checkpoint from disk
func (m *Manager) Delete(ctx context.Context, operationID string) error {
	if operationID == "" {
		return types.NewAppError(
			types.ErrInvalidInput,
			"operation ID cannot be empty",
			nil,
		)
	}
	
	filename := fmt.Sprintf("checkpoint_%s.json", operationID)
	filepath := filepath.Join(m.checkpointDir, filename)
	
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return types.NewAppError(
			types.ErrFileSystem,
			"failed to delete checkpoint file",
			err,
		)
	}
	
	return nil
}

// List returns all available checkpoints
func (m *Manager) List(ctx context.Context) ([]*Checkpoint, error) {
	files, err := filepath.Glob(filepath.Join(m.checkpointDir, "checkpoint_*.json"))
	if err != nil {
		return nil, types.NewAppError(
			types.ErrFileSystem,
			"failed to list checkpoint files",
			err,
		)
	}
	
	var checkpoints []*Checkpoint
	
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}
		
		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip corrupted files
		}
		
		// Skip expired checkpoints
		if time.Since(checkpoint.Timestamp) > m.maxAge {
			continue
		}
		
		checkpoints = append(checkpoints, &checkpoint)
	}
	
	return checkpoints, nil
}

// Cleanup removes expired checkpoints
func (m *Manager) Cleanup(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join(m.checkpointDir, "checkpoint_*.json"))
	if err != nil {
		return types.NewAppError(
			types.ErrFileSystem,
			"failed to list checkpoint files for cleanup",
			err,
		)
	}
	
	var errors []error
	
	for _, file := range files {
		// Read file to get timestamp
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			// Delete corrupted files
			os.Remove(file)
			continue
		}
		
		// Delete expired checkpoints
		if time.Since(checkpoint.Timestamp) > m.maxAge {
			if err := os.Remove(file); err != nil {
				errors = append(errors, err)
			}
		}
	}
	
	if len(errors) > 0 {
		return types.NewAppError(
			types.ErrFileSystem,
			fmt.Sprintf("failed to cleanup %d checkpoint files", len(errors)),
			errors[0],
		)
	}
	
	return nil
}

// NewCheckpoint creates a new checkpoint with the given parameters
func NewCheckpoint(operationID, operationType string, progress float64, state map[string]interface{}, nextStep string) *Checkpoint {
	return &Checkpoint{
		OperationID:   operationID,
		OperationType: operationType,
		Timestamp:     time.Now(),
		Progress:      progress,
		State:         state,
		NextStep:      nextStep,
		Metadata:      make(map[string]string),
	}
}

// UpdateProgress updates the progress of a checkpoint
func (c *Checkpoint) UpdateProgress(progress float64, nextStep string) {
	c.Progress = progress
	c.NextStep = nextStep
	c.Timestamp = time.Now()
}

// AddMetadata adds metadata to the checkpoint
func (c *Checkpoint) AddMetadata(key, value string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
}

// GetState retrieves a value from the checkpoint state
func (c *Checkpoint) GetState(key string) (interface{}, bool) {
	if c.State == nil {
		return nil, false
	}
	value, exists := c.State[key]
	return value, exists
}

// SetState sets a value in the checkpoint state
func (c *Checkpoint) SetState(key string, value interface{}) {
	if c.State == nil {
		c.State = make(map[string]interface{})
	}
	c.State[key] = value
} 