package domain

// Locale represents localized strings for a language
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

// GetLocales returns all available locales
func GetLocales() map[string]*Locale {
	return map[string]*Locale{
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
}

