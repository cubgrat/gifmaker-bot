package telegram

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps Telegram Bot API
type Bot struct {
	api *tgbotapi.BotAPI
}

// NewBot creates a new Telegram bot instance
func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{api: api}, nil
}

// GetAPI returns the underlying BotAPI
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}

// GetFileLink returns the download link for a file
func (b *Bot) GetFileLink(fileID string) (string, error) {
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	return file.Link(b.api.Token), nil
}

// SendMessage sends a text message
func (b *Bot) SendMessage(chatID int64, text string, replyMarkup interface{}) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}
	sent, err := b.api.Send(msg)
	if err != nil {
		return 0, err
	}
	return sent.MessageID, nil
}

// EditMessageText edits a message text
func (b *Bot) EditMessageText(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.api.Send(msg)
	return err
}

// SendAnimation sends an animation (GIF)
func (b *Bot) SendAnimation(chatID int64, filePath string, caption string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileBytes := tgbotapi.FileBytes{
		Name:  "animation.gif",
		Bytes: fileData,
	}

	msg := tgbotapi.NewAnimation(chatID, fileBytes)
	msg.Caption = caption
	_, err = b.api.Send(msg)
	return err
}

// DeleteMessage deletes a message
func (b *Bot) DeleteMessage(chatID int64, messageID int) error {
	_, err := b.api.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
	return err
}

// AnswerCallback answers a callback query
func (b *Bot) AnswerCallback(callbackID string) error {
	_, err := b.api.Request(tgbotapi.NewCallback(callbackID, ""))
	return err
}

// StopReceivingUpdates stops receiving updates
func (b *Bot) StopReceivingUpdates() {
	b.api.StopReceivingUpdates()
}

// GetUpdatesChan returns a channel for receiving updates
func (b *Bot) GetUpdatesChan(timeout int) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	return b.api.GetUpdatesChan(u)
}

// GetSelf returns bot information
func (b *Bot) GetSelf() tgbotapi.User {
	return b.api.Self
}

