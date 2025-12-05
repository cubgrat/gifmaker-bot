package service

import (
	"gifmaker-bot/internal/domain"
)

// LocaleService handles locale operations
type LocaleService struct {
	userLang *domain.UserLanguage
	locales  map[string]*domain.Locale
}

// NewLocaleService creates a new locale service
func NewLocaleService(userLang *domain.UserLanguage) *LocaleService {
	return &LocaleService{
		userLang: userLang,
		locales:  domain.GetLocales(),
	}
}

// GetLocale returns the locale for a chat ID
func (s *LocaleService) GetLocale(chatID int64) *domain.Locale {
	lang := s.userLang.Get(chatID)
	locale, ok := s.locales[lang]
	if !ok {
		locale = s.locales["ru"] // fallback to Russian
	}
	return locale
}

// SetLanguage sets the language for a chat ID
func (s *LocaleService) SetLanguage(chatID int64, lang string) {
	s.userLang.Set(chatID, lang)
}

