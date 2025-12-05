package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot struct {
		Token string `yaml:"token"`
	} `yaml:"bot"`
	GIF struct {
		Quality string `yaml:"quality"`
		FPS     int    `yaml:"fps"`
		Width   int    `yaml:"width"`
		Colors  int    `yaml:"colors"`
	} `yaml:"gif"`
	Processing struct {
		MaxConcurrent    int `yaml:"max_concurrent"`
		MaxVideoDuration int `yaml:"max_video_duration"`
	} `yaml:"processing"`
}

type ProcessingTask struct {
	MessageID     int
	ChatID        int64
	VideoFileID   string
	VideoFilePath string
	StatusMsgID   int
	QueuePosition int
	CancelContext context.Context
	CancelFunc    context.CancelFunc
}

type ProcessingQueue struct {
	mu            sync.Mutex
	activeTasks   map[int]*ProcessingTask
	waitingQueue  []*ProcessingTask
	nextTaskID    int
	maxConcurrent int
}

type Locale struct {
	StartMessage     string
	HelpMessage      string
	SendVideoMessage string
	VideoTooLong     string
	Processing       string
	SendingGIF       string
	GIFReady         string
	InQueue          string
	InQueuePlural    string
	ErrorGetFile     string
	ErrorDownload    string
	ErrorDuration    string
	ErrorConversion  string
	ErrorCreateGIF   string
	ErrorFileTooBig  string
	ErrorOpenGIF     string
	ErrorReadGIF     string
	ErrorSendGIF     string
	ErrorSendVideo   string
	LanguageChanged  string
	SelectLanguage   string
	HelpTitle        string
	HelpDescription  string
	HelpUsage        string
	HelpLimits       string
	HelpLanguage     string
}

var locales = map[string]*Locale{
	"ru": {
		StartMessage:     "👋 Привет! Отправьте мне видео файл (до 20 секунд), и я конвертирую его в GIF.",
		HelpMessage:      "📖 Справка",
		SendVideoMessage: "Пожалуйста, отправьте видео файл",
		VideoTooLong:     "Видео слишком длинное. Максимальная длительность: %d секунд",
		Processing:       "Обрабатываю видео...",
		SendingGIF:       "Отправляю GIF...",
		GIFReady:         "Ваш GIF готов!",
		InQueue:          "⏳ Вы ожидаете в очереди, перед вами %d файл",
		InQueuePlural:    "⏳ Вы ожидаете в очереди, перед вами %d файлов",
		ErrorGetFile:     "Не удалось получить файл видео",
		ErrorDownload:    "Не удалось скачать видео",
		ErrorDuration:    "Не удалось определить длительность видео",
		ErrorConversion:  "Ошибка при конвертации видео в GIF",
		ErrorCreateGIF:   "Ошибка при создании GIF файла",
		ErrorFileTooBig:  "Полученный GIF файл слишком большой. Попробуйте видео с меньшей длительностью или разрешением.",
		ErrorOpenGIF:     "Ошибка при открытии GIF файла",
		ErrorReadGIF:     "Ошибка при чтении GIF файла",
		ErrorSendGIF:     "Ошибка при отправке GIF",
		ErrorSendVideo:   "Пожалуйста, отправьте видео файл, а не GIF",
		LanguageChanged:  "✅ Язык изменен на русский",
		SelectLanguage:   "Выберите язык / Select language:",
		HelpTitle:        "📖 Справка по использованию бота",
		HelpDescription:  "Этот бот конвертирует видео файлы в GIF анимации.",
		HelpUsage:        "📹 Отправьте видео файл длительностью до 20 секунд, и бот автоматически создаст из него GIF.",
		HelpLimits:       "⚙️ Ограничения:\n• Максимальная длительность: 20 секунд\n• Если пользователей много, то вы попадете в очередь ожидания\n• Размер GIF не должен превышать 20 МБ",
		HelpLanguage:     "🌐 Для смены языка используйте кнопку \"Язык / Language\"",
	},
	"en": {
		StartMessage:     "👋 Hello! Send me a video file (up to 20 seconds), and I'll convert it to a GIF.",
		HelpMessage:      "📖 Help",
		SendVideoMessage: "Please send a video file",
		VideoTooLong:     "Video is too long. Maximum duration: %d seconds",
		Processing:       "Processing video...",
		SendingGIF:       "Sending GIF...",
		GIFReady:         "Your GIF is ready!",
		InQueue:          "⏳ You are waiting in queue, %d file ahead",
		InQueuePlural:    "⏳ You are waiting in queue, %d files ahead",
		ErrorGetFile:     "Failed to get video file",
		ErrorDownload:    "Failed to download video",
		ErrorDuration:    "Failed to determine video duration",
		ErrorConversion:  "Error converting video to GIF",
		ErrorCreateGIF:   "Error creating GIF file",
		ErrorFileTooBig:  "The resulting GIF file is too large. Try a video with shorter duration or lower resolution.",
		ErrorOpenGIF:     "Error opening GIF file",
		ErrorReadGIF:     "Error reading GIF file",
		ErrorSendGIF:     "Error sending GIF",
		ErrorSendVideo:   "Please send a video file, not a GIF",
		LanguageChanged:  "✅ Language changed to English",
		SelectLanguage:   "Select language / Выберите язык:",
		HelpTitle:        "📖 Bot Usage Guide",
		HelpDescription:  "This bot converts video files to GIF animations.",
		HelpUsage:        "📹 Send a video file up to 20 seconds long, and the bot will automatically create a GIF from it.",
		HelpLimits:       "⚙️ Limits:\n• Maximum duration: 20 seconds\n• If users are many, you will be in the waiting queue\n• GIF size must not exceed 20 MB",
		HelpLanguage:     "🌐 To change language, use the \"Language / Язык\" button",
	},
}

type UserLanguage struct {
	mu    sync.RWMutex
	langs map[int64]string // chatID -> language code
}

func NewUserLanguage() *UserLanguage {
	return &UserLanguage{
		langs: make(map[int64]string),
	}
}

func (ul *UserLanguage) Get(chatID int64) string {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	lang, ok := ul.langs[chatID]
	if !ok {
		return "ru" // default language
	}
	return lang
}

func (ul *UserLanguage) Set(chatID int64, lang string) {
	ul.mu.Lock()
	defer ul.mu.Unlock()
	ul.langs[chatID] = lang
}

func getLocale(chatID int64, userLang *UserLanguage) *Locale {
	lang := userLang.Get(chatID)
	locale, ok := locales[lang]
	if !ok {
		locale = locales["ru"] // fallback to Russian
	}
	return locale
}

func NewProcessingQueue(maxConcurrent int) *ProcessingQueue {
	return &ProcessingQueue{
		activeTasks:   make(map[int]*ProcessingTask),
		waitingQueue:  make([]*ProcessingTask, 0),
		maxConcurrent: maxConcurrent,
		nextTaskID:    1,
	}
}

func (pq *ProcessingQueue) AddTask(task *ProcessingTask) int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	taskID := pq.nextTaskID
	pq.nextTaskID++

	// Calculate queue position
	queuePos := len(pq.waitingQueue) + len(pq.activeTasks)
	task.QueuePosition = queuePos

	if len(pq.activeTasks) < pq.maxConcurrent {
		pq.activeTasks[taskID] = task
		task.QueuePosition = 0
		go pq.processTask(taskID, task)
	} else {
		pq.waitingQueue = append(pq.waitingQueue, task)
	}

	return taskID
}

func (pq *ProcessingQueue) processTask(taskID int, task *ProcessingTask) {
	defer func() {
		pq.mu.Lock()
		delete(pq.activeTasks, taskID)

		// Start next task from queue if available
		if len(pq.waitingQueue) > 0 {
			nextTask := pq.waitingQueue[0]
			pq.waitingQueue = pq.waitingQueue[1:]
			nextTask.QueuePosition = 0
			pq.activeTasks[pq.nextTaskID] = nextTask
			go pq.processTask(pq.nextTaskID, nextTask)
			pq.nextTaskID++
		}

		// Update queue positions for waiting tasks
		for i, t := range pq.waitingQueue {
			t.QueuePosition = i + 1
		}

		pq.mu.Unlock()
	}()

	// Process the task
	processVideoToGIF(task)
}

func (pq *ProcessingQueue) GetQueuePosition(messageID int) int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for _, task := range pq.activeTasks {
		if task.MessageID == messageID {
			return 0
		}
	}

	for i, task := range pq.waitingQueue {
		if task.MessageID == messageID {
			return i + 1
		}
	}

	return -1
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run()
}

func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	outputStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func convertVideoToGIF(videoPath, outputPath string, config *Config) error {
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
		return fmt.Errorf("failed to generate palette: %v", err)
	}
	defer os.Remove(palettePath)

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
	return cmd.Run()
}

func processVideoToGIF(task *ProcessingTask) {
	bot := task.CancelContext.Value("bot").(*tgbotapi.BotAPI)
	config := task.CancelContext.Value("config").(*Config)
	userLang := task.CancelContext.Value("userLang").(*UserLanguage)
	locale := getLocale(task.ChatID, userLang)

	// Create temp directory for this task
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("gifbot_%d_%d", task.ChatID, task.MessageID))
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	videoPath := filepath.Join(tempDir, "video.mp4")
	gifPath := filepath.Join(tempDir, "output.gif")

	// Download video
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: task.VideoFileID})
	if err != nil {
		sendError(bot, task.ChatID, locale.ErrorGetFile, locale)
		return
	}

	fileURL := file.Link(bot.Token)
	if err := downloadFile(fileURL, videoPath); err != nil {
		sendError(bot, task.ChatID, locale.ErrorDownload, locale)
		return
	}

	// Check video duration
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		sendError(bot, task.ChatID, locale.ErrorDuration, locale)
		return
	}

	if duration > float64(config.Processing.MaxVideoDuration) {
		sendError(bot, task.ChatID, fmt.Sprintf(locale.VideoTooLong,
			config.Processing.MaxVideoDuration), locale)
		return
	}

	// Update status: processing
	updateStatus(bot, task.ChatID, task.StatusMsgID, locale.Processing)

	// Convert to GIF
	if err := convertVideoToGIF(videoPath, gifPath, config); err != nil {
		sendError(bot, task.ChatID, locale.ErrorConversion, locale)
		return
	}

	// Check if file exists and get size
	fileInfo, err := os.Stat(gifPath)
	if err != nil {
		sendError(bot, task.ChatID, locale.ErrorCreateGIF, locale)
		return
	}

	// Telegram has a 50MB limit for files, but for GIFs it's usually 20MB
	if fileInfo.Size() > 20*1024*1024 {
		sendError(bot, task.ChatID, locale.ErrorFileTooBig, locale)
		return
	}

	// Send GIF
	updateStatus(bot, task.ChatID, task.StatusMsgID, locale.SendingGIF)

	gifFile, err := os.Open(gifPath)
	if err != nil {
		sendError(bot, task.ChatID, locale.ErrorOpenGIF, locale)
		return
	}
	defer gifFile.Close()

	fileBytes := tgbotapi.FileBytes{
		Name:  "animation.gif",
		Bytes: make([]byte, fileInfo.Size()),
	}

	if _, err := gifFile.Read(fileBytes.Bytes); err != nil {
		sendError(bot, task.ChatID, locale.ErrorReadGIF, locale)
		return
	}

	msg := tgbotapi.NewAnimation(task.ChatID, fileBytes)
	msg.Caption = locale.GIFReady

	if _, err := bot.Send(msg); err != nil {
		sendError(bot, task.ChatID, locale.ErrorSendGIF, locale)
		return
	}

	// Delete status message
	bot.Request(tgbotapi.NewDeleteMessage(task.ChatID, task.StatusMsgID))
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func sendError(bot *tgbotapi.BotAPI, chatID int64, message string, locale *Locale) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ %s", message))
	bot.Send(msg)
}

func sendStatusMessage(bot *tgbotapi.BotAPI, chatID int64, position int, locale *Locale) (int, error) {
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
	msg := tgbotapi.NewMessage(chatID, text)
	resp, err := bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return resp.MessageID, nil
}

func updateStatus(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	bot.Send(msg)
}

func createMainKeyboard(locale *Locale) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🌐 Язык / Language"),
			tgbotapi.NewKeyboardButton("📖 Справка / Help"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

func createLanguageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en"),
		),
	)
}

func startQueueUpdater(bot *tgbotapi.BotAPI, queue *ProcessingQueue, config *Config, userLang *UserLanguage) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		queue.mu.Lock()

		// Update status messages for waiting tasks
		for _, task := range queue.waitingQueue {
			position := task.QueuePosition
			locale := getLocale(task.ChatID, userLang)
			var text string
			if position == 1 {
				text = fmt.Sprintf(locale.InQueue, position)
			} else {
				text = fmt.Sprintf(locale.InQueuePlural, position)
			}
			msg := tgbotapi.NewEditMessageText(task.ChatID, task.StatusMsgID, text)
			bot.Send(msg)
		}

		queue.mu.Unlock()
	}
}

func main() {
	// Load config
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	if config.Bot.Token == "YOUR_BOT_TOKEN_HERE" {
		log.Fatal("Пожалуйста, установите токен бота в config.yaml")
	}

	// Check FFmpeg
	if err := checkFFmpeg(); err != nil {
		log.Fatal("FFmpeg не найден. Пожалуйста, установите FFmpeg (см. README.md)")
	}

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)

	// Initialize queue
	queue := NewProcessingQueue(config.Processing.MaxConcurrent)

	// Initialize user language storage
	userLang := NewUserLanguage()

	// Start queue updater
	go startQueueUpdater(bot, queue, config, userLang)

	// Setup update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, останавливаю бота...")
		cancel()
		bot.StopReceivingUpdates()
	}()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			log.Println("Бот остановлен")
			return
		case update := <-updates:
			// Handle callback queries (button presses)
			if update.CallbackQuery != nil {
				callback := update.CallbackQuery
				chatID := callback.Message.Chat.ID

				if strings.HasPrefix(callback.Data, "lang_") {
					lang := strings.TrimPrefix(callback.Data, "lang_")
					userLang.Set(chatID, lang)
					locale := getLocale(chatID, userLang)

					// Answer callback
					bot.Request(tgbotapi.NewCallback(callback.ID, ""))

					// Send confirmation
					msg := tgbotapi.NewMessage(chatID, locale.LanguageChanged)
					msg.ReplyMarkup = createMainKeyboard(locale)
					bot.Send(msg)

					// Delete language selection message
					bot.Request(tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID))
				}
				continue
			}

			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			locale := getLocale(chatID, userLang)

			// Handle text commands and buttons
			if update.Message.Text != "" {
				text := update.Message.Text

				if text == "/start" {
					msg := tgbotapi.NewMessage(chatID, locale.StartMessage)
					msg.ReplyMarkup = createMainKeyboard(locale)
					bot.Send(msg)
					continue
				}

				if text == "🌐 Язык / Language" || text == "/language" || text == "/lang" {
					msg := tgbotapi.NewMessage(chatID, locale.SelectLanguage)
					msg.ReplyMarkup = createLanguageKeyboard()
					bot.Send(msg)
					continue
				}

				if text == "📖 Справка / Help" || text == "/help" {
					helpText := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
						locale.HelpTitle,
						locale.HelpDescription,
						locale.HelpUsage,
						locale.HelpLimits,
						locale.HelpLanguage)
					msg := tgbotapi.NewMessage(chatID, helpText)
					msg.ReplyMarkup = createMainKeyboard(locale)
					bot.Send(msg)
					continue
				}
			}

			// Handle video messages
			if update.Message.Video != nil {
				fileID := update.Message.Video.FileID

				// Create task context
				taskCtx, taskCancel := context.WithCancel(context.Background())
				taskCtx = context.WithValue(taskCtx, "bot", bot)
				taskCtx = context.WithValue(taskCtx, "config", config)
				taskCtx = context.WithValue(taskCtx, "userLang", userLang)

				// Determine queue position and send status
				queue.mu.Lock()
				queuePos := len(queue.waitingQueue) + len(queue.activeTasks)
				queue.mu.Unlock()

				var statusMsgID int
				var err error
				if queuePos >= config.Processing.MaxConcurrent {
					statusMsgID, err = sendStatusMessage(bot, chatID, queuePos-config.Processing.MaxConcurrent+1, locale)
				} else {
					statusMsgID, err = sendStatusMessage(bot, chatID, 0, locale)
				}
				if err != nil {
					log.Printf("Ошибка отправки статуса: %v", err)
				}

				task := &ProcessingTask{
					MessageID:     update.Message.MessageID,
					ChatID:        chatID,
					VideoFileID:   fileID,
					StatusMsgID:   statusMsgID,
					CancelContext: taskCtx,
					CancelFunc:    taskCancel,
				}

				queue.AddTask(task)
			} else if update.Message.Document != nil {
				// Check if document is a video
				mimeType := update.Message.Document.MimeType
				fileName := update.Message.Document.FileName

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
					// Create task context
					taskCtx, taskCancel := context.WithCancel(context.Background())
					taskCtx = context.WithValue(taskCtx, "bot", bot)
					taskCtx = context.WithValue(taskCtx, "config", config)
					taskCtx = context.WithValue(taskCtx, "userLang", userLang)

					// Determine queue position and send status
					queue.mu.Lock()
					queuePos := len(queue.waitingQueue) + len(queue.activeTasks)
					queue.mu.Unlock()

					var statusMsgID int
					var err error
					if queuePos >= config.Processing.MaxConcurrent {
						statusMsgID, err = sendStatusMessage(bot, chatID, queuePos-config.Processing.MaxConcurrent+1, locale)
					} else {
						statusMsgID, err = sendStatusMessage(bot, chatID, 0, locale)
					}
					if err != nil {
						log.Printf("Ошибка отправки статуса: %v", err)
					}

					task := &ProcessingTask{
						MessageID:     update.Message.MessageID,
						ChatID:        chatID,
						VideoFileID:   update.Message.Document.FileID,
						StatusMsgID:   statusMsgID,
						CancelContext: taskCtx,
						CancelFunc:    taskCancel,
					}

					queue.AddTask(task)
				} else {
					msg := tgbotapi.NewMessage(chatID, locale.SendVideoMessage)
					msg.ReplyMarkup = createMainKeyboard(locale)
					bot.Send(msg)
				}
			}
		}
	}
}
