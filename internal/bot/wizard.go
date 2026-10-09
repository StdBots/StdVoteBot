package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/StdBots/StdVoteBot/internal/credit"
	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WizardManager handles interactive poll creation steps
type WizardManager struct {
	cache *database.MemoryCache
}

// NewWizardManager creates a new wizard manager
func NewWizardManager(cache *database.MemoryCache) *WizardManager {
	return &WizardManager{cache: cache}
}

func (w *WizardManager) cacheKey(userID int64) string {
	return fmt.Sprintf("wizard:%d", userID)
}

// StartWizard initializes poll creation for a user
func (w *WizardManager) StartWizard(userID int64) {
	state := &database.WizardState{
		Step:    1,
		Options: []string{},
	}
	w.cache.Set(w.cacheKey(userID), state, 30*time.Minute)
}

// GetState retrieves current wizard state
func (w *WizardManager) GetState(userID int64) (*database.WizardState, bool) {
	val, ok := w.cache.Get(w.cacheKey(userID))
	if !ok {
		return nil, false
	}
	state, ok := val.(*database.WizardState)
	return state, ok
}

// SetQuestion records the question and advances to option selection
func (w *WizardManager) SetQuestion(userID int64, question string) {
	state, ok := w.GetState(userID)
	if !ok {
		state = &database.WizardState{}
	}
	state.Question = strings.TrimSpace(question)
	state.Step = 2
	w.cache.Set(w.cacheKey(userID), state, 30*time.Minute)
}

// Cancel removes active wizard state
func (w *WizardManager) Cancel(userID int64) {
	w.cache.Delete(w.cacheKey(userID))
}

// FormatPollMessage formats the Telegram message body for a poll
func FormatPollMessage(poll *database.Poll) string {
	var sb strings.Builder

	sb.WriteString("🗳️ ")
	sb.WriteString(poll.Question)
	sb.WriteString("\n\n")

	status := "🟢 Active Poll"
	if poll.IsClosed {
		status = "🔴 Poll Closed"
	}
	sb.WriteString(fmt.Sprintf("📊 <b>Status:</b> %s | <b>Total Votes:</b> %d\n", status, poll.TotalVotes))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	for _, opt := range poll.Options {
		pct := 0.0
		if poll.TotalVotes > 0 {
			pct = (float64(opt.VotesCount) / float64(poll.TotalVotes)) * 100.0
		}
		sb.WriteString(fmt.Sprintf("▪️ <b>%s</b> — %d vote(s) (%.1f%%)\n", opt.Text, opt.VotesCount, pct))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(credit.GetFooter())

	return credit.GetWatermarked(sb.String())
}

// BuildPollKeyboard creates the dynamic inline buttons with real-time counters
func BuildPollKeyboard(poll *database.Poll, botUsername string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, opt := range poll.Options {
		pct := 0.0
		if poll.TotalVotes > 0 {
			pct = (float64(opt.VotesCount) / float64(poll.TotalVotes)) * 100.0
		}

		label := fmt.Sprintf("%s [%d]", opt.Text, opt.VotesCount)
		if poll.TotalVotes > 0 {
			label = fmt.Sprintf("%s [%.0f%% | %d]", opt.Text, pct, opt.VotesCount)
		}

		callbackData := fmt.Sprintf("vote:%s:%d", poll.ID, opt.ID)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Share and Manage buttons row
	btnShare := tgbotapi.NewInlineKeyboardButtonSwitch(
		"🔗 Share Poll in Chat",
		poll.ID,
	)
	btnRefresh := tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", fmt.Sprintf("refresh:%s", poll.ID))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnShare, btnRefresh))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
