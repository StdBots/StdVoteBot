package bot

import (
	"fmt"
	"log"

	"github.com/StdBots/StdVoteBot/internal/config"
	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Middleware provides security, authorization, and force-sub checks
type Middleware struct {
	cfg      *config.Config
	userRepo *database.UserRepo
	bot      *tgbotapi.BotAPI
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware(cfg *config.Config, userRepo *database.UserRepo, bot *tgbotapi.BotAPI) *Middleware {
	return &Middleware{
		cfg:      cfg,
		userRepo: userRepo,
		bot:      bot,
	}
}

// TrackUser registers or updates the user profile in database
func (m *Middleware) TrackUser(from *tgbotapi.User) {
	if from == nil {
		return
	}
	go func() {
		err := m.userRepo.RegisterOrUpdate(from.ID, from.UserName, from.FirstName, from.LastName)
		if err != nil {
			log.Printf("Error tracking user %d: %v", from.ID, err)
		}
	}()
}

// IsOwner checks if user is the bot owner
func (m *Middleware) IsOwner(userID int64) bool {
	return userID == m.cfg.OwnerID
}

// CheckForceSub checks if the user is a member of the mandatory channel
func (m *Middleware) CheckForceSub(userID int64) (bool, error) {
	if m.cfg.ForceSubChannel == "" {
		return true, nil
	}

	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: "@" + m.cfg.ForceSubChannel,
			UserID:             userID,
		},
	}

	member, err := m.bot.GetChatMember(chatConfig)
	if err != nil {
		// If bot is not admin in the channel or channel doesn't exist, fail open
		return true, nil
	}

	status := member.Status
	if status == "creator" || status == "administrator" || status == "member" || status == "restricted" {
		return true, nil
	}

	return false, nil
}

// GetForceSubMarkup returns the keyboard asking the user to join the channel
func (m *Middleware) GetForceSubMarkup(pollID string) tgbotapi.InlineKeyboardMarkup {
	channelURL := fmt.Sprintf("https://t.me/%s", m.cfg.ForceSubChannel)
	btnChannel := tgbotapi.NewInlineKeyboardButtonURL("📢 Join Channel", channelURL)
	btnRetry := tgbotapi.NewInlineKeyboardButtonData("🔄 Joined & Try Again", fmt.Sprintf("fsub_check:%s", pollID))

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnChannel),
		tgbotapi.NewInlineKeyboardRow(btnRetry),
	)
}
