package internal

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileOperations handles file system operations such as copy, move, delete, etc.
type FileOperations struct {
	model *Model
}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations(model *Model) *FileOperations {
	return &FileOperations{
		model: model,
	}
}

// CopyFile copies the selected file to the clipboard
func (fo *FileOperations) CopyFile() error {
	selectedEntry, err := fo.getSelectedEntry()
	if err != nil {
		return err
	}

	fo.model.Clipboard = filepath.Join(fo.model.Cwd, selectedEntry.Name())
	fo.model.ClipboardOp = "copy"
	fo.model.Err = fmt.Errorf("Copied: %s", selectedEntry.Name())
	return nil
}

// CutFile marks the selected file for cutting (moving)
func (fo *FileOperations) CutFile() error {
	selectedEntry, err := fo.getSelectedEntry()
	if err != nil {
		return err
	}

	fo.model.Clipboard = filepath.Join(fo.model.Cwd, selectedEntry.Name())
	fo.model.ClipboardOp = "cut"
	fo.model.Err = fmt.Errorf("Cut: %s", selectedEntry.Name())
	return nil
}

// PasteFile pastes the file from clipboard to the current directory
func (fo *FileOperations) PasteFile() error {
	if fo.model.Clipboard == "" {
		return fmt.Errorf("Clipboard is empty")
	}

	// Get the base name of the source file/directory
	sourceName := filepath.Base(fo.model.Clipboard)
	destPath := filepath.Join(fo.model.Cwd, sourceName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		if fo.model.Config.Behavior.ConfirmFileOperations {
			fo.model.ConfirmOperation = true
			fo.model.ConfirmPrompt = fmt.Sprintf("Overwrite %s?", sourceName)
			fo.model.ConfirmAction = func() error {
				return fo.executePaste(destPath)
			}
			return nil
		}
		return fmt.Errorf("Destination already exists: %s", destPath)
	}

	return fo.executePaste(destPath)
}

// executePaste performs the actual paste operation
func (fo *FileOperations) executePaste(destPath string) error {
	var err error
	source := fo.model.Clipboard
	operation := fo.model.ClipboardOp

	// Check if source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Error accessing source: %w", err)
	}

	if sourceInfo.IsDir() {
		if operation == "copy" {
			err = fo.copyDir(source, destPath)
		} else {
			err = os.Rename(source, destPath)
		}
	} else {
		if operation == "copy" {
			err = fo.copyFile(source, destPath)
		} else {
			err = os.Rename(source, destPath)
		}
	}

	if err != nil {
		return fmt.Errorf("Error during %s operation: %w", operation, err)
	}

	// Clear clipboard if it was a cut operation
	if operation == "cut" {
		fo.model.Clipboard = ""
		fo.model.ClipboardOp = ""
	}

	// Refresh the directory
	fo.model.ReadDir(fo.model.Cwd)
	return nil
}

// DeleteFile deletes the selected file or directory
func (fo *FileOperations) DeleteFile() error {
	selectedEntry, err := fo.getSelectedEntry()
	if err != nil {
		return err
	}

	targetPath := filepath.Join(fo.model.Cwd, selectedEntry.Name())

	// Check if the target is a directory
	if selectedEntry.IsDir() {
		// Confirm directory deletion
		if fo.model.Config.Behavior.ConfirmFileOperations {
			fo.model.ConfirmOperation = true
			fo.model.ConfirmPrompt = fmt.Sprintf("Delete directory: %s?", selectedEntry.Name())
			fo.model.ConfirmAction = func() error {
				return os.RemoveAll(targetPath)
			}
			return nil
		}
		return os.RemoveAll(targetPath)
	} else {
		// Confirm file deletion
		if fo.model.Config.Behavior.ConfirmFileOperations {
			fo.model.ConfirmOperation = true
			fo.model.ConfirmPrompt = fmt.Sprintf("Delete file: %s?", selectedEntry.Name())
			fo.model.ConfirmAction = func() error {
				return os.Remove(targetPath)
			}
			return nil
		}
		return os.Remove(targetPath)
	}
}

// RenameFile renames the selected file or directory
func (fo *FileOperations) RenameFile(newName string) error {
	selectedEntry, err := fo.getSelectedEntry()
	if err != nil {
		return err
	}

	// Check if the new name is empty
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("New name cannot be empty")
	}

	oldPath := filepath.Join(fo.model.Cwd, selectedEntry.Name())
	newPath := filepath.Join(fo.model.Cwd, newName)

	// Check if destination already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("File already exists: %s", newPath)
	}

	// Rename the file
	return os.Rename(oldPath, newPath)
}

// CreateDirectory creates a new directory in the current working directory
func (fo *FileOperations) CreateDirectory(name string) error {
	// Check if the name is empty
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("Directory name cannot be empty")
	}

	dirPath := filepath.Join(fo.model.Cwd, name)

	// Check if directory already exists
	if _, err := os.Stat(dirPath); err == nil {
		return fmt.Errorf("Directory already exists: %s", dirPath)
	}

	// Create the directory with default permissions
	return os.MkdirAll(dirPath, 0755)
}

// CreateFile creates a new empty file in the current working directory
func (fo *FileOperations) CreateFile(name string) error {
	// Check if the name is empty
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("File name cannot be empty")
	}

	filePath := filepath.Join(fo.model.Cwd, name)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("File already exists: %s", filePath)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

// Helper functions

// getSelectedEntry returns the currently selected entry
func (fo *FileOperations) getSelectedEntry() (fs.DirEntry, error) {
	return getSelectedEntry(*fo.model)
}

// copyFile copies a single file from src to dst
func (fo *FileOperations) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file mode
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// copyDir recursively copies a directory from src to dst
func (fo *FileOperations) copyDir(src, dst string) error {
	// Get source info
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return err
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := fo.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := fo.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddBookmark adds the current directory to bookmarks
func (fo *FileOperations) AddBookmark() error {
	currentPath := fo.model.Cwd

	// Check if already bookmarked
	for _, bookmark := range fo.model.Bookmarks {
		if bookmark == currentPath {
			return fmt.Errorf("Already bookmarked: %s", currentPath)
		}
	}

	// Add to bookmarks
	fo.model.Bookmarks = append(fo.model.Bookmarks, currentPath)

	// Save to persistent state
	state, _ := LoadState()
	state.Bookmarks = fo.model.Bookmarks
	if err := SaveState(&state); err != nil {
		return fmt.Errorf("Error saving bookmarks: %w", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark
func (fo *FileOperations) RemoveBookmark(index int) error {
	// Check bounds
	if index < 0 || index >= len(fo.model.Bookmarks) {
		return fmt.Errorf("Invalid bookmark index")
	}

	// Remove bookmark
	fo.model.Bookmarks = append(fo.model.Bookmarks[:index], fo.model.Bookmarks[index+1:]...)

	// Save to persistent state
	state, _ := LoadState()
	state.Bookmarks = fo.model.Bookmarks
	if err := SaveState(&state); err != nil {
		return fmt.Errorf("Error saving bookmarks: %w", err)
	}

	return nil
}