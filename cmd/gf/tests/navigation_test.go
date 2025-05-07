package tests

import (
	"os"
	"path/filepath"
	"testing"

	"gf/internal"
)

func TestDirectoryNavigation(t *testing.T) {
	// Create a temporary directory structure for testing
	baseDir, err := os.MkdirTemp("", "navigation_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// Create subdirectories
	subDir1 := filepath.Join(baseDir, "subdir1")
	subDir2 := filepath.Join(baseDir, "subdir2")
	subSubDir := filepath.Join(subDir1, "subsubdir")

	for _, dir := range []string{subDir1, subDir2, subSubDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string][]string{
		baseDir:  {"basefile1.txt", "basefile2.txt"},
		subDir1:  {"sub1file1.txt", "sub1file2.txt"},
		subDir2:  {"sub2file1.txt", "sub2file2.txt"},
		subSubDir: {"subsubfile1.txt", "subsubfile2.txt"},
	}

	for dir, files := range testFiles {
		for _, file := range files {
			filePath := filepath.Join(dir, file)
			if _, err := os.Create(filePath); err != nil {
				t.Fatalf("Failed to create file %s: %v", filePath, err)
			}
		}
	}

	// Create model and initialize with base directory
	model := internal.NewModel()
	model.ReadDir(baseDir)

	// Test case 1: Initial directory loading
	if model.Cwd != baseDir {
		t.Errorf("Expected current directory to be %s, got %s", baseDir, model.Cwd)
	}

	expectedBaseEntries := 4 // subdir1, subdir2, basefile1.txt, basefile2.txt
	if len(model.Entries) != expectedBaseEntries {
		t.Errorf("Expected %d entries in base dir, got %d", expectedBaseEntries, len(model.Entries))
	}

	// Find subdir1 index
	subdir1Index := -1
	for i, entry := range model.Entries {
		if entry.Name() == "subdir1" && entry.IsDir() {
			subdir1Index = i
			break
		}
	}

	if subdir1Index == -1 {
		t.Fatalf("Could not find subdir1 in entries")
	}

	// Test case 2: Navigate into subdirectory
	model.Cursor = subdir1Index
	// Simulate entering the directory (normally done through key handling)
	model.ReadDir(filepath.Join(model.Cwd, model.Entries[subdir1Index].Name()))

	if model.Cwd != subDir1 {
		t.Errorf("Expected current directory to be %s, got %s", subDir1, model.Cwd)
	}

	expectedSubDir1Entries := 3 // subsubdir, sub1file1.txt, sub1file2.txt
	if len(model.Entries) != expectedSubDir1Entries {
		t.Errorf("Expected %d entries in subDir1, got %d", expectedSubDir1Entries, len(model.Entries))
	}

	// Test case 3: Navigate deeper
	subsubdirIndex := -1
	for i, entry := range model.Entries {
		if entry.Name() == "subsubdir" && entry.IsDir() {
			subsubdirIndex = i
			break
		}
	}

	if subsubdirIndex == -1 {
		t.Fatalf("Could not find subsubdir in entries")
	}

	model.Cursor = subsubdirIndex
	// Simulate entering the directory
	model.ReadDir(filepath.Join(model.Cwd, model.Entries[subsubdirIndex].Name()))

	if model.Cwd != subSubDir {
		t.Errorf("Expected current directory to be %s, got %s", subSubDir, model.Cwd)
	}

	expectedSubSubDirEntries := 2 // subsubfile1.txt, subsubfile2.txt
	if len(model.Entries) != expectedSubSubDirEntries {
		t.Errorf("Expected %d entries in subSubDir, got %d", expectedSubSubDirEntries, len(model.Entries))
	}

	// Test case 4: Navigate back to parent
	parentDir := filepath.Dir(model.Cwd)
	model.ReadDir(parentDir)

	if model.Cwd != subDir1 {
		t.Errorf("Expected current directory to be %s, got %s", subDir1, model.Cwd)
	}

	// Test case 5: Navigate back to root
	parentDir = filepath.Dir(model.Cwd)
	model.ReadDir(parentDir)

	if model.Cwd != baseDir {
		t.Errorf("Expected current directory to be %s, got %s", baseDir, model.Cwd)
	}
}

func TestCursorNavigation(t *testing.T) {
	// Create a temporary directory with files
	tempDir, err := os.MkdirTemp("", "cursor_nav_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create 10 files
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tempDir, "file"+string(rune(i+48))+".txt")
		if _, err := os.Create(filename); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Initialize model
	model := internal.NewModel()
	model.ReadDir(tempDir)

	// Test initial cursor position
	if model.Cursor != 0 {
		t.Errorf("Initial cursor position should be 0, got %d", model.Cursor)
	}

	// Test cursor movement down
	origCursor := model.Cursor
	for i := 0; i < 5; i++ {
		// Simulate cursor down (normally done through key handling)
		if model.Cursor < len(model.Entries)-1 {
			model.Cursor++
		}
	}

	if model.Cursor != origCursor+5 {
		t.Errorf("Expected cursor at position %d after moving down 5 times, got %d", origCursor+5, model.Cursor)
	}

	// Test cursor movement up
	for i := 0; i < 2; i++ {
		// Simulate cursor up
		if model.Cursor > 0 {
			model.Cursor--
		}
	}

	if model.Cursor != origCursor+3 {
		t.Errorf("Expected cursor at position %d after moving up 2 times, got %d", origCursor+3, model.Cursor)
	}

	// Test cursor out of bounds (attempt to move past last item)
	model.Cursor = len(model.Entries) - 1
	startPos := model.Cursor

	// Try to move beyond end
	if model.Cursor < len(model.Entries)-1 {
		model.Cursor++
	}

	if model.Cursor != startPos {
		t.Errorf("Cursor should not move beyond last entry, expected %d, got %d", startPos, model.Cursor)
	}

	// Test cursor at start (attempt to move before first item)
	model.Cursor = 0
	
	// Try to move before beginning
	if model.Cursor > 0 {
		model.Cursor--
	}

	if model.Cursor != 0 {
		t.Errorf("Cursor should not move before first entry, expected 0, got %d", model.Cursor)
	}

	// Test cursor jump to bottom
	model.Cursor = len(model.Entries) - 1
	if model.Cursor != 9 {
		t.Errorf("Expected cursor at last position (9), got %d", model.Cursor)
	}

	// Test cursor jump to top
	model.Cursor = 0
	if model.Cursor != 0 {
		t.Errorf("Expected cursor at first position (0), got %d", model.Cursor)
	}
}