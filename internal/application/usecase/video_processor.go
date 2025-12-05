package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"gifmaker-bot/internal/application/service"
	"gifmaker-bot/internal/domain"
	"gifmaker-bot/internal/infrastructure/ffmpeg"
	"gifmaker-bot/internal/infrastructure/storage"
	"gifmaker-bot/internal/infrastructure/telegram"
)

// VideoProcessor handles video processing use cases
type VideoProcessor struct {
	bot       *telegram.Bot
	converter *ffmpeg.Converter
	fileStore *storage.FileStorage
	config    *domain.Config
	localeSvc *service.LocaleService
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(
	bot *telegram.Bot,
	converter *ffmpeg.Converter,
	fileStore *storage.FileStorage,
	config *domain.Config,
	localeSvc *service.LocaleService,
) *VideoProcessor {
	return &VideoProcessor{
		bot:       bot,
		converter: converter,
		fileStore: fileStore,
		config:    config,
		localeSvc: localeSvc,
	}
}

// ProcessVideo processes a video task and converts it to GIF
func (vp *VideoProcessor) ProcessVideo(ctx context.Context, task *domain.ProcessingTask) error {
	locale := vp.localeSvc.GetLocale(task.ChatID)

	// Create temp directory for this task
	tempDir, err := vp.fileStore.CreateTempDir(fmt.Sprintf("gifbot_%d_%d_", task.ChatID, task.MessageID))
	if err != nil {
		vp.sendError(task.ChatID, locale.ErrorGetFile, locale)
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer vp.fileStore.RemoveDir(tempDir)

	videoPath := filepath.Join(tempDir, "video.mp4")
	gifPath := filepath.Join(tempDir, "output.gif")

	// Download video
	fileURL, err := vp.bot.GetFileLink(task.VideoFileID)
	if err != nil {
		vp.sendError(task.ChatID, locale.ErrorGetFile, locale)
		return fmt.Errorf("failed to get file link: %w", err)
	}

	if err := vp.fileStore.DownloadFile(fileURL, videoPath); err != nil {
		vp.sendError(task.ChatID, locale.ErrorDownload, locale)
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Check video duration
	duration, err := vp.converter.GetVideoDuration(videoPath)
	if err != nil {
		vp.sendError(task.ChatID, locale.ErrorDuration, locale)
		return fmt.Errorf("failed to get duration: %w", err)
	}

	if duration > float64(vp.config.Processing.MaxVideoDuration) {
		errorMsg := fmt.Sprintf(locale.VideoTooLong, vp.config.Processing.MaxVideoDuration)
		vp.sendError(task.ChatID, errorMsg, locale)
		return fmt.Errorf("video too long: %.2f seconds", duration)
	}

	// Update status: processing
	if err := vp.bot.EditMessageText(task.ChatID, task.StatusMsgID, locale.Processing); err != nil {
		// Log error but continue
	}

	// Convert to GIF
	if err := vp.converter.ConvertToGIF(videoPath, gifPath, vp.config); err != nil {
		vp.sendError(task.ChatID, locale.ErrorConversion, locale)
		return fmt.Errorf("failed to convert: %w", err)
	}

	// Check if file exists and get size
	fileSize, err := vp.fileStore.GetFileSize(gifPath)
	if err != nil {
		vp.sendError(task.ChatID, locale.ErrorCreateGIF, locale)
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// Telegram has a 50MB limit for files, but for GIFs it's usually 20MB
	const maxGIFSize = 20 * 1024 * 1024
	if fileSize > maxGIFSize {
		vp.sendError(task.ChatID, locale.ErrorFileTooBig, locale)
		return fmt.Errorf("GIF file too large: %d bytes", fileSize)
	}

	// Send GIF
	if err := vp.bot.EditMessageText(task.ChatID, task.StatusMsgID, locale.SendingGIF); err != nil {
		// Log error but continue
	}

	if err := vp.bot.SendAnimation(task.ChatID, gifPath, locale.GIFReady); err != nil {
		vp.sendError(task.ChatID, locale.ErrorSendGIF, locale)
		return fmt.Errorf("failed to send GIF: %w", err)
	}

	// Delete status message
	_ = vp.bot.DeleteMessage(task.ChatID, task.StatusMsgID)

	return nil
}

func (vp *VideoProcessor) sendError(chatID int64, message string, locale *domain.Locale) {
	_, _ = vp.bot.SendMessage(chatID, fmt.Sprintf("❌ %s", message), nil)
}

