package domain

import "context"

// ProcessingTask represents a video processing task
type ProcessingTask struct {
	ID            int
	MessageID     int
	ChatID        int64
	VideoFileID   string
	StatusMsgID   int
	QueuePosition int
	CancelContext context.Context
	CancelFunc    context.CancelFunc
}

