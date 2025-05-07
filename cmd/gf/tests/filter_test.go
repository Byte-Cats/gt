package tests

import (
	"os"
	"reflect"
	"testing"

	"gf/internal"
)

func TestApplyFilter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filter_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.txt",
		"document.md",
		"image.png",
		"testimage.jpg",
	}
	
	for _, file := range testFiles {
		_, err := os.Create(tempDir + "/" + file)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create test directories
	testDirs := []string{
		"dir1",
		"dir2",
		"testdir",
	}
	
	for _, dir := range testDirs {
		err := os.Mkdir(tempDir+"/"+dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create model and set working directory
	model := internal.NewModel()
	model.ReadDir(tempDir)

	// Store original entries count
	originalCount := len(model.Entries)
	if originalCount != len(testFiles)+len(testDirs) {
		t.Fatalf("Expected %d entries, got %d", len(testFiles)+len(testDirs), originalCount)
	}

	// Test case 1: Empty filter should show all entries
	model.FilterInput.SetValue("")
	model.ApplyFilter()
	if len(model.FilteredEntries) != originalCount {
		t.Errorf("Empty filter should show all %d entries, got %d", originalCount, len(model.FilteredEntries))
	}

	// Test case 2: Filter for "file"
	model.FilterInput.SetValue("file")
	model.ApplyFilter()
	expectedFileCount := 2 // file1.txt, file2.txt
	if len(model.FilteredEntries) != expectedFileCount {
		t.Errorf("Filter 'file' should show %d entries, got %d", expectedFileCount, len(model.FilteredEntries))
	}

	// Test case 3: Filter for "test"
	model.FilterInput.SetValue("test")
	model.ApplyFilter()
	expectedTestCount := 2 // testimage.jpg, testdir
	if len(model.FilteredEntries) != expectedTestCount {
		t.Errorf("Filter 'test' should show %d entries, got %d", expectedTestCount, len(model.FilteredEntries))
	}

	// Test case 4: Filter for non-existent term
	model.FilterInput.SetValue("nonexistent")
	model.ApplyFilter()
	if len(model.FilteredEntries) != 0 {
		t.Errorf("Filter 'nonexistent' should show 0 entries, got %d", len(model.FilteredEntries))
	}

	// Test case 5: Case insensitivity
	model.FilterInput.SetValue("FILE")
	model.ApplyFilter()
	if len(model.FilteredEntries) != expectedFileCount {
		t.Errorf("Case-insensitive filter 'FILE' should show %d entries, got %d", expectedFileCount, len(model.FilteredEntries))
	}
}

func TestFilterReset(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filter_reset_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	for i := 0; i < 5; i++ {
		_, err := os.Create(tempDir + "/file" + string(rune(i+48)) + ".txt")
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create model and set working directory
	model := internal.NewModel()
	model.ReadDir(tempDir)

	// Apply a filter
	model.FilterInput.SetValue("file1")
	model.ApplyFilter()
	
	// Store filtered entries
	filteredEntries := model.FilteredEntries
	
	// Verify filter applied correctly
	if len(filteredEntries) != 1 {
		t.Errorf("Expected 1 filtered entry, got %d", len(filteredEntries))
	}

	// Change directory (should reset filter)
	parentDir := tempDir + "/.."
	model.ReadDir(parentDir)
	
	// Check filter was reset
	if model.FilterInput.Value() != "" {
		t.Errorf("Filter should be reset after directory change, got: %s", model.FilterInput.Value())
	}
	
	// Filtered entries should equal all entries after reset
	if !reflect.DeepEqual(model.FilteredEntries, model.Entries) {
		t.Errorf("Filtered entries should equal all entries after filter reset")
	}
}