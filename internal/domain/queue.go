package domain

import "sync"

// ProcessingQueue manages video processing tasks
type ProcessingQueue struct {
	mu            sync.Mutex
	activeTasks   map[int]*ProcessingTask
	waitingQueue  []*ProcessingTask
	nextTaskID    int
	maxConcurrent int
}

// NewProcessingQueue creates a new processing queue
func NewProcessingQueue(maxConcurrent int) *ProcessingQueue {
	return &ProcessingQueue{
		activeTasks:   make(map[int]*ProcessingTask),
		waitingQueue:  make([]*ProcessingTask, 0),
		maxConcurrent: maxConcurrent,
		nextTaskID:    1,
	}
}

// AddTask adds a task to the queue
func (pq *ProcessingQueue) AddTask(task *ProcessingTask) int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	taskID := pq.nextTaskID
	pq.nextTaskID++

	// Calculate queue position
	queuePos := len(pq.waitingQueue) + len(pq.activeTasks)
	task.QueuePosition = queuePos
	task.ID = taskID

	if len(pq.activeTasks) < pq.maxConcurrent {
		pq.activeTasks[taskID] = task
		task.QueuePosition = 0
		return taskID
	}

	pq.waitingQueue = append(pq.waitingQueue, task)
	return taskID
}

// StartTask marks a task as active
func (pq *ProcessingQueue) StartTask(taskID int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if task, ok := pq.activeTasks[taskID]; ok {
		task.QueuePosition = 0
	}
}

// CompleteTask removes a task from active tasks and starts the next one
func (pq *ProcessingQueue) CompleteTask(taskID int) *ProcessingTask {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	delete(pq.activeTasks, taskID)

	// Start next task from queue if available
	var nextTask *ProcessingTask
	if len(pq.waitingQueue) > 0 {
		nextTask = pq.waitingQueue[0]
		pq.waitingQueue = pq.waitingQueue[1:]
		nextTask.QueuePosition = 0
		pq.activeTasks[nextTask.ID] = nextTask
	}

	// Update queue positions for waiting tasks
	for i, t := range pq.waitingQueue {
		t.QueuePosition = i + 1
	}

	return nextTask
}

// GetWaitingTasks returns all waiting tasks
func (pq *ProcessingQueue) GetWaitingTasks() []*ProcessingTask {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	result := make([]*ProcessingTask, len(pq.waitingQueue))
	copy(result, pq.waitingQueue)
	return result
}

// GetActiveCount returns the number of active tasks
func (pq *ProcessingQueue) GetActiveCount() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.activeTasks)
}

// GetQueueSize returns the total queue size
func (pq *ProcessingQueue) GetQueueSize() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.waitingQueue) + len(pq.activeTasks)
}

