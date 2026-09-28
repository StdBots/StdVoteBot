package bot

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/StdBots/StdVoteBot/internal/analytics"
	"github.com/StdBots/StdVoteBot/internal/broadcast"
	"github.com/StdBots/StdVoteBot/internal/config"
	"github.com/StdBots/StdVoteBot/internal/credit"
	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram Bot API client and all subsystem dependencies
type Bot struct {
	api             *tgbotapi.BotAPI
	cfg             *config.Config
	db              *database.MongoDB
	pollRepo        *database.PollRepo
	userRepo        *database.UserRepo
	cache           *database.MemoryCache
	wizard          *WizardManager
	middleware      *Middleware
	handler         *Handler
	callbackHandler *CallbackHandler
	inlineHandler   *InlineHandler
	broadcastEngine *broadcast.Engine
	analyticsEngine *analytics.Engine
}

// New creates and initializes the complete bot instance
func New(cfg *config.Config, db *database.MongoDB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	if cfg.Env == "development" {
		api.Debug = true
	}

	pollRepo := database.NewPollRepo(db)
	userRepo := database.NewUserRepo(db)
	cache := database.NewMemoryCache(10 * time.Minute)
	wizard := NewWizardManager(cache)
	middleware := NewMiddleware(cfg, userRepo, api)

	handler := NewHandler(api, pollRepo, userRepo, wizard, middleware)
	callbackHandler := NewCallbackHandler(api, pollRepo, userRepo, wizard, middleware, handler)
	inlineHandler := NewInlineHandler(api, pollRepo)
	broadcastEngine := broadcast.NewEngine(api, userRepo, cfg.OwnerID)
	analyticsEngine := analytics.NewEngine(api, pollRepo, userRepo, cfg.OwnerID)

	return &Bot{
		api:             api,
		cfg:             cfg,
		db:              db,
		pollRepo:        pollRepo,
		userRepo:        userRepo,
		cache:           cache,
		wizard:          wizard,
		middleware:      middleware,
		handler:         handler,
		callbackHandler: callbackHandler,
		inlineHandler:   inlineHandler,
		broadcastEngine: broadcastEngine,
		analyticsEngine: analyticsEngine,
	}, nil
}

// Start runs the update polling loop until context cancellation
func (b *Bot) Start(ctx context.Context) error {
	log.Printf("🤖 StdVoteBot authorized on account @%s (ID: %d)", b.api.Self.UserName, b.api.Self.ID)

	// Verify Credit Integrity
	intact, tampered := credit.VerifyIntegrity()
	if !intact {
		log.Printf("[SECURITY WARNING] Integrity violation detected: %v", tampered)
	}
	credit.ReportForkStatus(b.api.Self.UserName, intact)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Shutting down bot update loop...")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			go b.processUpdate(update)
		}
	}
}

// processUpdate dispatches an incoming Telegram update to appropriate handlers
func (b *Bot) processUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC RECOVERED] in update %d: %v", update.UpdateID, r)
		}
	}()

	// 1. Handle Inline Queries
	if update.InlineQuery != nil {
		b.inlineHandler.Handle(update.InlineQuery)
		return
	}

	// 2. Handle Callback Queries
	if update.CallbackQuery != nil {
		b.callbackHandler.Handle(update.CallbackQuery)
		return
	}

	// 3. Handle Messages
	if update.Message != nil {
		msg := update.Message

		// Check for commands
		if msg.IsCommand() {
			cmd := strings.ToLower(msg.Command())
			switch cmd {
			case "start":
				b.handler.HandleStart(msg)
			case "newpoll", "create", "vote":
				b.handler.HandleNewPoll(msg)
			case "mypolls":
				b.handler.HandleMyPolls(msg)
			case "cancel":
				b.handler.HandleCancel(msg)
			case "help":
				b.handler.HandleHelp(msg)
			case "stats", "std":
				b.analyticsEngine.HandleStats(msg)
			case "broadcast":
				b.broadcastEngine.HandleBroadcast(msg)
			default:
				reply := tgbotapi.NewMessage(msg.Chat.ID, "❓ Unknown command. Send /help to view available commands.")
				_, _ = b.api.Send(reply)
			}
			return
		}

		// Check if broadcast is waiting for content from admin
		if b.broadcastEngine.IsAwaitingBroadcast(msg.From.ID) {
			b.broadcastEngine.ExecuteBroadcast(msg)
			return
		}

		// Handle normal text in wizard
		if msg.Text != "" {
			b.handler.HandleTextMessage(msg)
		}
	}
}
