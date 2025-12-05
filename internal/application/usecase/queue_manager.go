package usecase

import (
	"context"
	"fmt"
	"time"

	"gifmaker-bot/internal/application/service"
	"gifmaker-bot/internal/domain"
	"gifmaker-bot/internal/infrastructure/telegram"
)

// QueueManager manages the processing queue
type QueueManager struct {
	queue      *domain.ProcessingQueue
	processor  *VideoProcessor
	bot        *telegram.Bot
	localeSvc  *service.LocaleService
	config     *domain.Config
}

// NewQueueManager creates a new queue manager
func NewQueueManager(
	queue *domain.ProcessingQueue,
	processor *VideoProcessor,
	bot *telegram.Bot,
	localeSvc *service.LocaleService,
	config *domain.Config,
) *QueueManager {
	return &QueueManager{
		queue:     queue,
		processor: processor,
		bot:       bot,
		localeSvc: localeSvc,
		config:    config,
	}
}

// AddTask adds a task to the queue and starts processing if possible
func (qm *QueueManager) AddTask(task *domain.ProcessingTask) {
	taskID := qm.queue.AddTask(task)
	task.ID = taskID

	// If task can start immediately, process it
	if task.QueuePosition == 0 {
		go qm.processTask(task)
	}
}

// processTask processes a task
func (qm *QueueManager) processTask(task *domain.ProcessingTask) {
	defer func() {
		// Complete task and start next one
		nextTask := qm.queue.CompleteTask(task.ID)
		if nextTask != nil {
			go qm.processTask(nextTask)
		}
	}()

	// Process the video
	ctx := context.Background()
	if err := qm.processor.ProcessVideo(ctx, task); err != nil {
		// Error already sent to user in ProcessVideo
		return
	}
}

// StartQueueUpdater starts a goroutine that updates queue status messages
func (qm *QueueManager) StartQueueUpdater() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		waitingTasks := qm.queue.GetWaitingTasks()
		for _, task := range waitingTasks {
			locale := qm.localeSvc.GetLocale(task.ChatID)
			position := task.QueuePosition

			var text string
			if position == 1 {
				text = fmt.Sprintf(locale.InQueue, position)
			} else {
				text = fmt.Sprintf(locale.InQueuePlural, position)
			}

			_ = qm.bot.EditMessageText(task.ChatID, task.StatusMsgID, text)
		}
	}
}

// GetQueuePosition returns the queue position for a message
func (qm *QueueManager) GetQueuePosition(chatID int64) int {
	queueSize := qm.queue.GetQueueSize()
	activeCount := qm.queue.GetActiveCount()

	if queueSize <= activeCount {
		return 0
	}

	return queueSize - activeCount
}

