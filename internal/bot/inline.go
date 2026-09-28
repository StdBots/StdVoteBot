package bot

import (
	"fmt"
	"strings"

	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InlineHandler handles Telegram inline query requests for seamless poll sharing
type InlineHandler struct {
	bot      *tgbotapi.BotAPI
	pollRepo *database.PollRepo
}

// NewInlineHandler initializes inline query handler
func NewInlineHandler(bot *tgbotapi.BotAPI, pollRepo *database.PollRepo) *InlineHandler {
	return &InlineHandler{
		bot:      bot,
		pollRepo: pollRepo,
	}
}

// Handle processes incoming inline queries (@StdVoteBot <query>)
func (h *InlineHandler) Handle(query *tgbotapi.InlineQuery) {
	searchTerm := strings.TrimSpace(query.Query)
	var results []interface{}

	if searchTerm != "" {
		// User typed a specific poll ID
		pollID := strings.TrimPrefix(searchTerm, "poll_")
		poll, err := h.pollRepo.GetPoll(pollID)
		if err == nil && poll != nil {
			article := h.buildArticle(poll)
			results = append(results, article)
		}
	}

	// If no specific match or query is empty, show user's recent polls
	if len(results) == 0 {
		polls, err := h.pollRepo.GetUserPolls(query.From.ID, 10)
		if err == nil {
			for _, poll := range polls {
				article := h.buildArticle(poll)
				results = append(results, article)
			}
		}
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     1,
		Results:       results,
	}

	_, _ = h.bot.Request(inlineConf)
}

func (h *InlineHandler) buildArticle(poll *database.Poll) tgbotapi.InlineQueryResultArticle {
	markup := BuildPollKeyboard(poll, h.bot.Self.UserName)
	text := FormatPollMessage(poll)

	msgContent := tgbotapi.InputTextMessageContent{
		Text:      text,
		ParseMode: "HTML",
	}

	status := "Active"
	if poll.IsClosed {
		status = "Closed"
	}

	article := tgbotapi.NewInlineQueryResultArticleHTML(
		poll.ID,
		fmt.Sprintf("🗳️ %s", poll.Question),
		text,
	)
	article.Description = fmt.Sprintf("Votes: %d | Status: %s | ID: %s", poll.TotalVotes, status, poll.ID)
	article.ReplyMarkup = &markup
	article.InputMessageContent = msgContent

	return article
}
