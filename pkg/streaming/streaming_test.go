package streaming

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestStreamProcessor(t *testing.T) {
	// Create a test file
	testFile := "/tmp/test_streaming.txt"
	content := "line1\nline2\nline3\n"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test file:", err)
	}
	defer os.Remove(testFile)
	
	// Test streaming
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatal("Failed to open test file:", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineCount := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "line") {
			t.Errorf("Unexpected line content: %s", line)
		}
		lineCount++
	}
	
	if lineCount != 3 {
		t.Errorf("Expected 3 lines, got %d", lineCount)
	}
	
	if err := scanner.Err(); err != nil {
		t.Fatal("Scanner error:", err)
	}
}

func TestStreamingBatch(t *testing.T) {
	// Test batch processing
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	batchSize := 2
	batches := make([][]string, 0)
	
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		batches = append(batches, batch)
	}
	
	// Should have 3 batches: [item1,item2], [item3,item4], [item5]
	if len(batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(batches))
	}
	
	if len(batches[0]) != 2 || len(batches[1]) != 2 || len(batches[2]) != 1 {
		t.Error("Unexpected batch sizes")
	}
}