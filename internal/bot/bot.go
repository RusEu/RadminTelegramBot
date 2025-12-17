package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"github.com/RusEu/RadminTelegramBot/internal/config"
	"github.com/RusEu/RadminTelegramBot/internal/security"
	"github.com/RusEu/RadminTelegramBot/internal/system"
)

// Bot represents the Telegram bot instance
type Bot struct {
	api      *tgbotapi.BotAPI
	config   *config.Config
	logger   *log.Logger
	auth     *security.AuthManager
	sysInfo  *system.InfoManager
	sysMon   *system.MonitoringManager
	sessions map[int64]*UserSession
	mu       sync.RWMutex
}

// UserSession represents a user session
type UserSession struct {
	UserID      int64
	Username    string
	StartTime   time.Time
	LastCommand time.Time
	CurrentPath string
	IsAdmin     bool
}

// NewBot creates a new bot instance
func NewBot(cfg *config.Config, logger *log.Logger) (*Bot, error) {
	logger.Info("🔑 Initializing Telegram API...")
	
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	logger.Infof("✅ Bot authorized as: @%s", api.Self.UserName)

	// Initialize components
	auth := security.NewAuthManager(cfg, logger)
	sysInfo := system.NewInfoManager(logger)
	sysMon := system.NewMonitoringManager(cfg, logger)

	bot := &Bot{
		api:      api,
		config:   cfg,
		logger:   logger,
		auth:     auth,
		sysInfo:  sysInfo,
		sysMon:   sysMon,
		sessions: make(map[int64]*UserSession),
	}

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("🎯 Starting bot polling...")

	// Start monitoring in background
	go b.sysMon.StartMonitoring(ctx, b.sendAlert)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("🛑 Bot context cancelled, stopping...")
			return nil
		case update := <-updates:
			b.handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	var userID int64
	var username string
	var chatID int64

	if update.Message != nil {
		userID = update.Message.From.ID
		username = update.Message.From.UserName
		chatID = update.Message.Chat.ID
		b.logger.Infof("📨 Received message from @%s (ID: %d): %s", username, userID, update.Message.Text)
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
		username = update.CallbackQuery.From.UserName
		chatID = update.CallbackQuery.Message.Chat.ID
		b.logger.Infof("🖱️ Received callback from @%s (ID: %d): %s", username, userID, update.CallbackQuery.Data)
	}

	// Check authorization
	if !b.auth.IsAuthorized(userID) {
		b.logger.Warnf("🚫 Unauthorized access attempt from @%s (ID: %d)", username, userID)
		msg := tgbotapi.NewMessage(chatID, "❌ Access denied. You are not authorized to use this bot.")
		b.api.Send(msg)
		return
	}

	// Rate limiting
	if !b.auth.CheckRateLimit(userID) {
		b.logger.Warnf("⚡ Rate limit exceeded for @%s (ID: %d)", username, userID)
		msg := tgbotapi.NewMessage(chatID, "⚡ Rate limit exceeded. Please wait before sending more commands.")
		b.api.Send(msg)
		return
	}

	// Handle message or callback
	if update.Message != nil {
		b.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(update.CallbackQuery)
	}
}

// getOrCreateSession gets or creates a user session
func (b *Bot) getOrCreateSession(userID int64, username string) *UserSession {
	b.mu.Lock()
	defer b.mu.Unlock()

	session, exists := b.sessions[userID]
	if !exists || time.Since(session.LastCommand) > time.Duration(b.config.Security.SessionTimeout)*time.Second {
		session = &UserSession{
			UserID:      userID,
			Username:    username,
			StartTime:   time.Now(),
			LastCommand: time.Now(),
			CurrentPath: b.config.Server.WorkingDir,
			IsAdmin:     false,
		}
		b.sessions[userID] = session
		b.logger.Infof("🔐 Created new session for @%s (ID: %d)", username, userID)
	} else {
		session.LastCommand = time.Now()
	}

	return session
}

// sendAlert sends monitoring alerts to all authorized users
func (b *Bot) sendAlert(alertType, message string) {
	alertMsg := fmt.Sprintf("🚨 **%s Alert**\n\n%s\n\n*Server: %s*\n*Time: %s*",
		strings.Title(alertType),
		message,
		b.config.Server.Name,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	for _, userID := range b.config.Telegram.AllowedUsers {
		msg := tgbotapi.NewMessage(userID, alertMsg)
		msg.ParseMode = "Markdown"
		b.api.Send(msg)
	}

	b.logger.Warnf("🚨 Alert sent: %s - %s", alertType, message)
}

// sendMessage sends a message to the user
func (b *Bot) sendMessage(chatID int64, text string) {
	if len(text) > 4096 {
		// Split long messages
		for len(text) > 4096 {
			msg := tgbotapi.NewMessage(chatID, text[:4096])
			b.api.Send(msg)
			text = text[4096:]
		}
	}
	if len(text) > 0 {
		msg := tgbotapi.NewMessage(chatID, text)
		b.api.Send(msg)
	}
}

// sendDocument sends a document to the user
func (b *Bot) sendDocument(chatID int64, filePath, caption string) error {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	if caption != "" {
		doc.Caption = caption
	}
	_, err := b.api.Send(doc)
	return err
}

// sendKeyboard sends a message with inline keyboard
func (b *Bot) sendKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// createMainKeyboard creates the main menu keyboard
func (b *Bot) createMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 System Info", "system_info"),
			tgbotapi.NewInlineKeyboardButtonData("📈 Resources", "resources"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Processes", "processes"),
			tgbotapi.NewInlineKeyboardButtonData("📁 Files", "files"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Services", "services"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Logs", "logs"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Admin", "admin"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "help"),
		),
	)
}