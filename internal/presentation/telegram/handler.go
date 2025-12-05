package telegram

import (
	"fmt"
	"path/filepath"
	"strings"

	"gifmaker-bot/internal/application/service"
	"gifmaker-bot/internal/application/usecase"
	"gifmaker-bot/internal/domain"
	"gifmaker-bot/internal/infrastructure/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler handles Telegram bot updates
type Handler struct {
	bot         *telegram.Bot
	queueMgr    *usecase.QueueManager
	localeSvc   *service.LocaleService
	config      *domain.Config
}

// NewHandler creates a new Telegram handler
func NewHandler(
	bot *telegram.Bot,
	queueMgr *usecase.QueueManager,
	localeSvc *service.LocaleService,
	config *domain.Config,
) *Handler {
	return &Handler{
		bot:       bot,
		queueMgr:  queueMgr,
		localeSvc: localeSvc,
		config:    config,
	}
}

// HandleUpdate handles a Telegram update
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	locale := h.localeSvc.GetLocale(chatID)

	// Handle text commands and buttons
	if update.Message.Text != "" {
		h.handleTextMessage(chatID, update.Message.Text, locale)
		return
	}

	// Handle video messages
	if update.Message.Video != nil {
		h.handleVideoMessage(update.Message, locale)
		return
	}

	// Handle document messages
	if update.Message.Document != nil {
		h.handleDocumentMessage(update.Message, locale)
		return
	}
}

func (h *Handler) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID

	if strings.HasPrefix(callback.Data, "lang_") {
		lang := strings.TrimPrefix(callback.Data, "lang_")
		h.localeSvc.SetLanguage(chatID, lang)
		locale := h.localeSvc.GetLocale(chatID)

		// Answer callback
		_ = h.bot.AnswerCallback(callback.ID)

		// Send confirmation
		keyboard := CreateMainKeyboard()
		_, _ = h.bot.SendMessage(chatID, locale.LanguageChanged, keyboard)

		// Delete language selection message
		_ = h.bot.DeleteMessage(chatID, callback.Message.MessageID)
	}
}

func (h *Handler) handleTextMessage(chatID int64, text string, locale *domain.Locale) {
	switch text {
	case "/start":
		keyboard := CreateMainKeyboard()
		_, _ = h.bot.SendMessage(chatID, locale.StartMessage, keyboard)

	case "🌐 Язык / Language", "/language", "/lang":
		keyboard := CreateLanguageKeyboard()
		_, _ = h.bot.SendMessage(chatID, locale.SelectLanguage, keyboard)

	case "📖 Справка / Help", "/help":
		helpText := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			locale.HelpTitle,
			locale.HelpDescription,
			locale.HelpUsage,
			locale.HelpLimits,
			locale.HelpLanguage)
		keyboard := CreateMainKeyboard()
		_, _ = h.bot.SendMessage(chatID, helpText, keyboard)
	}
}

func (h *Handler) handleVideoMessage(message *tgbotapi.Message, locale *domain.Locale) {
	fileID := message.Video.FileID
	h.processVideoFile(message.Chat.ID, message.MessageID, fileID, locale)
}

func (h *Handler) handleDocumentMessage(message *tgbotapi.Message, locale *domain.Locale) {
	// Check if document is a video
	mimeType := message.Document.MimeType
	fileName := message.Document.FileName

	isVideo := false
	if mimeType != "" {
		isVideo = mimeType == "video/mp4" || mimeType == "video/quicktime" ||
			mimeType == "video/x-msvideo" || mimeType == "video/webm" ||
			mimeType == "video/x-matroska" || mimeType == "video/x-ms-wmv"
	}

	// Check by file extension if MIME type is not available
	if !isVideo && fileName != "" {
		ext := strings.ToLower(filepath.Ext(fileName))
		isVideo = ext == ".mp4" || ext == ".mov" || ext == ".avi" ||
			ext == ".webm" || ext == ".mkv" || ext == ".wmv" || ext == ".flv"
	}

	if isVideo {
		fileID := message.Document.FileID
		h.processVideoFile(message.Chat.ID, message.MessageID, fileID, locale)
	} else {
		keyboard := CreateMainKeyboard()
		_, _ = h.bot.SendMessage(message.Chat.ID, locale.SendVideoMessage, keyboard)
	}
}

func (h *Handler) processVideoFile(chatID int64, messageID int, fileID string, locale *domain.Locale) {
	// Determine queue position and send status
	queuePos := h.queueMgr.GetQueuePosition(chatID)

	var statusMsgID int
	var err error
	if queuePos >= h.config.Processing.MaxConcurrent {
		statusMsgID, err = h.sendStatusMessage(chatID, queuePos-h.config.Processing.MaxConcurrent+1, locale)
	} else {
		statusMsgID, err = h.sendStatusMessage(chatID, 0, locale)
	}
	if err != nil {
		// Log error but continue
		return
	}

	// Create task
	task := &domain.ProcessingTask{
		MessageID: messageID,
		ChatID:    chatID,
		VideoFileID: fileID,
		StatusMsgID: statusMsgID,
	}

	h.queueMgr.AddTask(task)
}

func (h *Handler) sendStatusMessage(chatID int64, position int, locale *domain.Locale) (int, error) {
	var text string
	if position == 0 {
		text = locale.Processing
	} else {
		if position == 1 {
			text = fmt.Sprintf(locale.InQueue, position)
		} else {
			text = fmt.Sprintf(locale.InQueuePlural, position)
		}
	}
	return h.bot.SendMessage(chatID, text, nil)
}

