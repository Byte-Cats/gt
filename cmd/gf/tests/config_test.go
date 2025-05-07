package tests

import (
	"os"
	"path/filepath"
	"testing"

	"gf/internal"
)

func TestDefaultConfig(t *testing.T) {
	// Save current working directory
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a temporary directory without config file
	tempDir, err := os.MkdirTemp("", "config_test_default")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origCwd) // Restore original directory when done

	// Initialize model (should use default config)
	model := internal.NewModel()

	// Test default style values
	if model.Config.Style.DirColor != "blue" { // Default blue
		t.Errorf("Expected default dir color to be blue, got %s", model.Config.Style.DirColor)
	}

	if model.Config.Style.FileColor != "white" { // Default white
		t.Errorf("Expected default file color to be white, got %s", model.Config.Style.FileColor)
	}

	// Test default behavior values
	if model.ShowHidden != false {
		t.Errorf("Expected default ShowHidden to be false, got %t", model.ShowHidden)
	}

	if model.Config.Behavior.PreviewMaxSizeKB != 100 {
		t.Errorf("Expected default PreviewMaxSizeKB to be 100, got %d", model.Config.Behavior.PreviewMaxSizeKB)
	}
}

func TestCustomConfig(t *testing.T) {
	// Save current working directory
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a temporary directory for custom config
	tempDir, err := os.MkdirTemp("", "config_test_custom")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origCwd)

	// Create a custom config file
	configContent := `
[style]
dir_color = "#FF0000" # Red
file_color = "#00FF00" # Green
selected_prefix = ">> "
selected_color = "#0000FF" # Blue
error_color = "#FF00FF" # Magenta
header_color = "#FFFF00" # Yellow
footer_color = "#00FFFF" # Cyan
border_color = "#FFFFFF" # White
border_type = "double"

[behavior]
show_hidden_by_default = true
confirm_file_operations = false
preview_max_size_kb = 200
remember_last_directory = false
`

	if err := os.WriteFile("config.toml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	// Initialize model (should use custom config)
	model := internal.NewModel()

	// Test custom style values
	if model.Config.Style.DirColor != "#FF0000" {
		t.Errorf("Expected custom dir color to be #FF0000, got %s", model.Config.Style.DirColor)
	}

	if model.Config.Style.FileColor != "#00FF00" {
		t.Errorf("Expected custom file color to be #00FF00, got %s", model.Config.Style.FileColor)
	}

	if model.Config.Style.SelectedPrefix != ">> " {
		t.Errorf("Expected custom selected prefix to be '>> ', got %s", model.Config.Style.SelectedPrefix)
	}

	if model.Config.Style.BorderType != "double" {
		t.Errorf("Expected custom border type to be 'double', got %s", model.Config.Style.BorderType)
	}

	// Test custom behavior values
	if model.ShowHidden != true {
		t.Errorf("Expected custom ShowHidden to be true, got %t", model.ShowHidden)
	}

	if model.Config.Behavior.PreviewMaxSizeKB != 200 {
		t.Errorf("Expected custom PreviewMaxSizeKB to be 200, got %d", model.Config.Behavior.PreviewMaxSizeKB)
	}

	if model.Config.Behavior.ConfirmFileOperations != false {
		t.Errorf("Expected custom ConfirmFileOperations to be false, got %t", model.Config.Behavior.ConfirmFileOperations)
	}

	if model.Config.Behavior.RememberLastDirectory != false {
		t.Errorf("Expected custom RememberLastDirectory to be false, got %t", model.Config.Behavior.RememberLastDirectory)
	}
}

func TestInvalidConfig(t *testing.T) {
	// Save current working directory
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a temporary directory for invalid config
	tempDir, err := os.MkdirTemp("", "config_test_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(origCwd)

	// Create an invalid config file
	invalidConfig := `
[style]
dir_color = "#FF0000" # Red
file_color = true # Invalid type
invalid_field = "test" # Unknown field

[behavior]
show_hidden_by_default = "not_a_bool" # Invalid type
`

	if err := os.WriteFile("config.toml", []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Initialize model (should use default config due to invalid file)
	model := internal.NewModel()

	// Should use default style values despite invalid config
	if model.Styles.SelectedPrefix != "> " { // Default prefix
		t.Errorf("Expected default selected prefix '> ' when config is invalid, got %s", model.Styles.SelectedPrefix)
	}

	// Check behavior defaults
	if model.ShowHidden != false {
		t.Errorf("Expected default ShowHidden to be false when config is invalid, got %t", model.ShowHidden)
	}
}

// TestConfigPersistence tests state persistence features
// Disabled for CI runs since it modifies system files
func TestConfigPersistence(t *testing.T) {
	t.Skip("Skipping persistence test that modifies system files")
	// Create a temporary directory for state
	tempDir, err := os.MkdirTemp("", "config_state_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Get home directory path and create .config/gf directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Could not get home directory for testing: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gf")
	origConfigDir := configDir + ".orig"

	// Backup existing config directory if it exists
	if _, err := os.Stat(configDir); err == nil {
		if err := os.Rename(configDir, origConfigDir); err != nil {
			t.Fatalf("Failed to backup existing config directory: %v", err)
		}
		defer os.Rename(origConfigDir, configDir) // Restore at end
	} else {
		defer os.RemoveAll(configDir) // Remove test directory at end
	}

	// Create test config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Initialize model with test paths and bookmarks
	model := internal.NewModel()
	model.Cwd = tempDir

	testBookmarks := []string{
		"/test/path1",
		"/test/path2",
		tempDir,
	}

	model.Bookmarks = testBookmarks

	// Save state using model's functionality
	state := internal.PersistentState{
		LastDirectory: model.Cwd,
		Bookmarks:     model.Bookmarks,
		RecentDirs:    []string{tempDir, "/previous/dir"},
	}

	// Use the exported SaveState function
	if err := internal.SaveState(&state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Initialize a new model, which should load the saved state
	newModel := internal.NewModel()

	// Check if state was properly loaded
	if newModel.Config.Behavior.RememberLastDirectory && newModel.Cwd != tempDir {
		t.Errorf("Expected current directory to be %s, got %s", tempDir, newModel.Cwd)
	}

	if len(newModel.Bookmarks) != len(testBookmarks) {
		t.Errorf("Expected %d bookmarks, got %d", len(testBookmarks), len(newModel.Bookmarks))
	}

	// Check specific bookmarks
	found := false
	for _, bookmark := range newModel.Bookmarks {
		if bookmark == tempDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find bookmark %s in loaded state", tempDir)
	}
}