package cards

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ImageService handles card image operations
type ImageService struct {
	basePath string
}

// NewImageService creates a new ImageService
func NewImageService(basePath string) *ImageService {
	return &ImageService{
		basePath: basePath,
	}
}

// DownloadImage downloads a card image from a URL and saves it locally
func (s *ImageService) DownloadImage(ctx context.Context, imageURL string) (string, error) {
	// Create a unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(imageURL))
	filepath := filepath.Join(s.basePath, filename)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Download the image
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	// Copy the image to the file
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return filename, nil
}

// GetImagePath returns the full path to a card image
func (s *ImageService) GetImagePath(filename string) string {
	return filepath.Join(s.basePath, filename)
}

// DeleteImage deletes a card image
func (s *ImageService) DeleteImage(filename string) error {
	filepath := filepath.Join(s.basePath, filename)
	return os.Remove(filepath)
}

// CleanupOldImages removes images that are no longer referenced by any card
func (s *ImageService) CleanupOldImages(ctx context.Context, referencedImages map[string]bool) error {
	// Read all files in the directory
	files, err := os.ReadDir(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Delete files that are not in the referenced images map
	for _, file := range files {
		if !file.IsDir() && !referencedImages[file.Name()] {
			if err := os.Remove(filepath.Join(s.basePath, file.Name())); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}
