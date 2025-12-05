package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// CreateMainKeyboard creates the main keyboard with language and help buttons
func CreateMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🌐 Язык / Language"),
			tgbotapi.NewKeyboardButton("📖 Справка / Help"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// CreateLanguageKeyboard creates the language selection keyboard
func CreateLanguageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en"),
		),
	)
}

