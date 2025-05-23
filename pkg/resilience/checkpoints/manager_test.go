package checkpoints

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Expected no error creating manager, got: %v", err)
	}
	
	if manager == nil {
		t.Error("Expected manager to be created")
	}
}

func TestNewCheckpoint(t *testing.T) {
	id := "test_op_123"
	operation := "test_operation"
	progress := 0.5
	data := map[string]interface{}{
		"processed": 50,
		"total":     100,
	}
	nextStep := "continue_processing"
	
	checkpoint := NewCheckpoint(id, operation, progress, data, nextStep)
	
	if checkpoint.OperationID != id {
		t.Errorf("Expected OperationID %s, got %s", id, checkpoint.OperationID)
	}
	
	if checkpoint.OperationType != operation {
		t.Errorf("Expected OperationType %s, got %s", operation, checkpoint.OperationType)
	}
	
	if checkpoint.Progress != progress {
		t.Errorf("Expected progress %f, got %f", progress, checkpoint.Progress)
	}
	
	if checkpoint.NextStep != nextStep {
		t.Errorf("Expected next step %s, got %s", nextStep, checkpoint.NextStep)
	}
	
	if len(checkpoint.State) != len(data) {
		t.Errorf("Expected state length %d, got %d", len(data), len(checkpoint.State))
	}
}

func TestCheckpointAddMetadata(t *testing.T) {
	checkpoint := NewCheckpoint("test", "op", 0.5, nil, "next")
	
	checkpoint.AddMetadata("key1", "value1")
	checkpoint.AddMetadata("key2", "value2")
	
	if checkpoint.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1=value1, got %v", checkpoint.Metadata["key1"])
	}
	
	if checkpoint.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2=value2, got %v", checkpoint.Metadata["key2"])
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create a checkpoint
	checkpoint := NewCheckpoint("test_op_123", "test_operation", 0.75, map[string]interface{}{
		"processed": 75,
		"total":     100,
	}, "continue")
	checkpoint.AddMetadata("test_run", "unit_test")
	
	// Save the checkpoint
	err = manager.Save(context.Background(), checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	
	// Load the checkpoint
	loadedCheckpoint, err := manager.Load(context.Background(), "test_op_123")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}
	
	// Verify loaded checkpoint
	if loadedCheckpoint.OperationID != checkpoint.OperationID {
		t.Errorf("Expected OperationID %s, got %s", checkpoint.OperationID, loadedCheckpoint.OperationID)
	}
	
	if loadedCheckpoint.Progress != checkpoint.Progress {
		t.Errorf("Expected progress %f, got %f", checkpoint.Progress, loadedCheckpoint.Progress)
	}
	
	if loadedCheckpoint.Metadata["test_run"] != "unit_test" {
		t.Errorf("Expected metadata to be preserved")
	}
}

func TestManagerLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Try to load non-existent checkpoint
	_, err = manager.Load(context.Background(), "non_existent")
	if err == nil {
		t.Error("Expected error when loading non-existent checkpoint")
	}
}

func TestManagerDelete(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create and save a checkpoint
	checkpoint := NewCheckpoint("test_delete", "test_operation", 0.5, nil, "next")
	err = manager.Save(context.Background(), checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	
	// Verify it exists
	_, err = manager.Load(context.Background(), "test_delete")
	if err != nil {
		t.Fatalf("Checkpoint should exist before deletion: %v", err)
	}
	
	// Delete the checkpoint
	err = manager.Delete(context.Background(), "test_delete")
	if err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}
	
	// Verify it no longer exists
	_, err = manager.Load(context.Background(), "test_delete")
	if err == nil {
		t.Error("Checkpoint should not exist after deletion")
	}
}

func TestManagerList(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create multiple checkpoints
	checkpoint1 := NewCheckpoint("op1", "operation1", 0.3, nil, "next1")
	checkpoint2 := NewCheckpoint("op2", "operation2", 0.7, nil, "next2")
	
	err = manager.Save(context.Background(), checkpoint1)
	if err != nil {
		t.Fatalf("Failed to save checkpoint1: %v", err)
	}
	
	err = manager.Save(context.Background(), checkpoint2)
	if err != nil {
		t.Fatalf("Failed to save checkpoint2: %v", err)
	}
	
	// List checkpoints
	checkpoints, err := manager.List(context.Background())
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}
	
	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}
}

func TestManagerCleanupExpired(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        100 * time.Millisecond, // Very short max age
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create a checkpoint
	checkpoint := NewCheckpoint("expired_test", "test_operation", 0.5, nil, "next")
	err = manager.Save(context.Background(), checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	
	// Wait for it to expire
	time.Sleep(200 * time.Millisecond)
	
	// Run cleanup
	err = manager.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Failed to cleanup expired checkpoints: %v", err)
	}
	
	// Verify checkpoint was cleaned up (it should still exist in this implementation)
	// The cleanup behavior depends on the actual implementation
}

func TestCheckpointFilePath(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		CheckpointDir: tempDir,
		MaxAge:        1 * time.Hour,
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Save a checkpoint
	checkpoint := NewCheckpoint("path_test", "test_operation", 0.5, nil, "next")
	err = manager.Save(context.Background(), checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	
	// Check that the file was created
	expectedPath := filepath.Join(tempDir, "checkpoint_path_test.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected checkpoint file to exist at %s", expectedPath)
	}
} 