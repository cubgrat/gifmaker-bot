package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// FileStorage handles file operations
type FileStorage struct{}

// NewFileStorage creates a new file storage
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// DownloadFile downloads a file from URL to local path
func (fs *FileStorage) DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads a file and returns its contents
func (fs *FileStorage) ReadFile(filepath string) ([]byte, error) {
	return os.ReadFile(filepath)
}

// FileExists checks if a file exists
func (fs *FileStorage) FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

// GetFileSize returns the size of a file
func (fs *FileStorage) GetFileSize(filepath string) (int64, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// RemoveFile removes a file
func (fs *FileStorage) RemoveFile(filepath string) error {
	return os.Remove(filepath)
}

// RemoveDir removes a directory and all its contents
func (fs *FileStorage) RemoveDir(dirpath string) error {
	return os.RemoveAll(dirpath)
}

// CreateTempDir creates a temporary directory
func (fs *FileStorage) CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

