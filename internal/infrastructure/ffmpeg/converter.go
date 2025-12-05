package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"

	"gifmaker-bot/internal/domain"
)

// Converter handles video to GIF conversion using FFmpeg
type Converter struct{}

// NewConverter creates a new FFmpeg converter
func NewConverter() *Converter {
	return &Converter{}
}

// GetVideoDuration returns the duration of a video file in seconds
func (c *Converter) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	var duration float64
	_, err = fmt.Sscanf(string(output), "%f", &duration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// ConvertToGIF converts a video file to GIF
func (c *Converter) ConvertToGIF(videoPath, outputPath string, config *domain.Config) error {
	// Build scale filter based on width setting
	var scaleFilter string
	if config.GIF.Width > 0 {
		scaleFilter = fmt.Sprintf("scale=%d:-1:flags=lanczos", config.GIF.Width)
	} else {
		scaleFilter = "scale=-1:-1:flags=lanczos"
	}

	// Add palette generation for better quality
	palettePath := outputPath + ".palette.png"
	paletteFilter := fmt.Sprintf("fps=%d,%s,palettegen=max_colors=%d",
		config.GIF.FPS, scaleFilter, config.GIF.Colors)

	paletteArgs := []string{
		"-i", videoPath,
		"-vf", paletteFilter,
		"-y", palettePath,
	}

	// Generate palette
	paletteCmd := exec.Command("ffmpeg", paletteArgs...)
	if err := paletteCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate palette: %w", err)
	}
	defer func() {
		_ = os.Remove(palettePath)
	}()

	// Convert to GIF using palette
	videoFilter := fmt.Sprintf("fps=%d,%s[x]", config.GIF.FPS, scaleFilter)
	paletteUseFilter := "[x][1:v]paletteuse"

	args := []string{
		"-i", videoPath,
		"-i", palettePath,
		"-lavfi", fmt.Sprintf("%s;%s", videoFilter, paletteUseFilter),
		"-y", outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert video to GIF: %w", err)
	}

	return nil
}

// CheckFFmpeg checks if FFmpeg is available
func CheckFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run()
}

