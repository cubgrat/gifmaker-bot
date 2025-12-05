package domain

import "sync"

// UserLanguage stores user language preferences
type UserLanguage struct {
	mu    sync.RWMutex
	langs map[int64]string // chatID -> language code
}

// NewUserLanguage creates a new UserLanguage instance
func NewUserLanguage() *UserLanguage {
	return &UserLanguage{
		langs: make(map[int64]string),
	}
}

// Get returns the language for a chat ID
func (ul *UserLanguage) Get(chatID int64) string {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	lang, ok := ul.langs[chatID]
	if !ok {
		return "ru" // default language
	}
	return lang
}

// Set sets the language for a chat ID
func (ul *UserLanguage) Set(chatID int64, lang string) {
	ul.mu.Lock()
	defer ul.mu.Unlock()
	ul.langs[chatID] = lang
}

