package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gifmaker-bot/internal/application/service"
	"gifmaker-bot/internal/application/usecase"
	"gifmaker-bot/internal/domain"
	"gifmaker-bot/internal/infrastructure/config"
	"gifmaker-bot/internal/infrastructure/ffmpeg"
	"gifmaker-bot/internal/infrastructure/storage"
	"gifmaker-bot/internal/infrastructure/telegram"
	telegramhandler "gifmaker-bot/internal/presentation/telegram"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Bot.Token == "YOUR_BOT_TOKEN_HERE" {
		log.Fatal("Please set bot token in config.yaml")
	}

	// Check FFmpeg
	if err := ffmpeg.CheckFFmpeg(); err != nil {
		log.Fatal("FFmpeg not found. Please install FFmpeg (see README.md)")
	}

	// Initialize infrastructure
	bot, err := telegram.NewBot(cfg.Bot.Token)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	log.Printf("Bot started: @%s", bot.GetSelf().UserName)

	converter := ffmpeg.NewConverter()
	fileStore := storage.NewFileStorage()

	// Initialize domain
	userLang := domain.NewUserLanguage()
	queue := domain.NewProcessingQueue(cfg.Processing.MaxConcurrent)

	// Initialize services
	localeSvc := service.NewLocaleService(userLang)

	// Initialize use cases
	videoProcessor := usecase.NewVideoProcessor(
		bot,
		converter,
		fileStore,
		cfg,
		localeSvc,
	)

	queueMgr := usecase.NewQueueManager(
		queue,
		videoProcessor,
		bot,
		localeSvc,
		cfg,
	)

	// Start queue updater
	go queueMgr.StartQueueUpdater()

	// Initialize handler
	handler := telegramhandler.NewHandler(
		bot,
		queueMgr,
		localeSvc,
		cfg,
	)

	// Setup update channel
	updates := bot.GetUpdatesChan(60)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping bot...")
		cancel()
		bot.StopReceivingUpdates()
	}()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopped")
			return
		case update := <-updates:
			handler.HandleUpdate(update)
		}
	}
}

