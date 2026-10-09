package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallbackHandler processes all inline keyboard interactions
type CallbackHandler struct {
	bot        *tgbotapi.BotAPI
	pollRepo   *database.PollRepo
	userRepo   *database.UserRepo
	wizard     *WizardManager
	middleware *Middleware
	handler    *Handler
}

// NewCallbackHandler initializes callback query handler
func NewCallbackHandler(bot *tgbotapi.BotAPI, pollRepo *database.PollRepo, userRepo *database.UserRepo, wizard *WizardManager, middleware *Middleware, handler *Handler) *CallbackHandler {
	return &CallbackHandler{
		bot:        bot,
		pollRepo:   pollRepo,
		userRepo:   userRepo,
		wizard:     wizard,
		middleware: middleware,
		handler:    handler,
	}
}

// Handle processes incoming callback queries
func (c *CallbackHandler) Handle(query *tgbotapi.CallbackQuery) {
	c.middleware.TrackUser(query.From)
	data := query.Data

	switch {
	case strings.HasPrefix(data, "vote:"):
		c.handleVote(query)
	case strings.HasPrefix(data, "quick:"):
		c.handleQuickPoll(query)
	case data == "wizard:start":
		c.wizard.StartWizard(query.From.ID)
		c.answer(query.ID, "", false)
		prompt := "📝 <b>Step 1 of 2: Poll Question</b>\n\nPlease send the <b>Question or Title</b> for your poll:"
		if query.Message != nil {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, prompt)
			msg.ParseMode = "HTML"
			_, _ = c.bot.Send(msg)
		}
	case data == "wizard:cancel":
		c.wizard.Cancel(query.From.ID)
		c.answer(query.ID, "❌ Poll creation cancelled.", true)
		if query.Message != nil {
			del := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
			_, _ = c.bot.Request(del)
		}
	case strings.HasPrefix(data, "refresh:"):
		pollID := strings.TrimPrefix(data, "refresh:")
		c.refreshPoll(query, pollID)
	case strings.HasPrefix(data, "view:"):
		pollID := strings.TrimPrefix(data, "view:")
		c.managePollView(query, pollID)
	case strings.HasPrefix(data, "close:"):
		pollID := strings.TrimPrefix(data, "close:")
		c.closePoll(query, pollID)
	case strings.HasPrefix(data, "delete:"):
		pollID := strings.TrimPrefix(data, "delete:")
		c.deletePoll(query, pollID)
	case strings.HasPrefix(data, "fsub_check:"):
		pollID := strings.TrimPrefix(data, "fsub_check:")
		isMember, _ := c.middleware.CheckForceSub(query.From.ID)
		if isMember {
			c.answer(query.ID, "✅ Subscription verified! You can now vote.", true)
			c.refreshPoll(query, pollID)
		} else {
			c.answer(query.ID, "❌ You have not joined the channel yet!", true)
		}
	case strings.HasPrefix(data, "mypolls:"):
		c.answer(query.ID, "", false)
		if query.Message != nil {
			c.handler.HandleMyPolls(query.Message)
		}
	default:
		c.answer(query.ID, "Action not recognized.", false)
	}
}

// handleVote processes a vote attempt
func (c *CallbackHandler) handleVote(query *tgbotapi.CallbackQuery) {
	parts := strings.Split(query.Data, ":")
	if len(parts) != 3 {
		c.answer(query.ID, "Invalid vote data.", true)
		return
	}

	pollID := parts[1]
	optionID, err := strconv.Atoi(parts[2])
	if err != nil {
		c.answer(query.ID, "Invalid option.", true)
		return
	}

	// 1. Force-Sub Check
	isMember, _ := c.middleware.CheckForceSub(query.From.ID)
	if !isMember {
		c.answer(query.ID, fmt.Sprintf("⚠️ Please join @%s to vote!", c.middleware.cfg.ForceSubChannel), true)
		return
	}

	// 2. Cast vote atomically
	success, alertMsg, updatedPoll, _ := c.pollRepo.CastVote(pollID, query.From.ID, optionID)
	c.answer(query.ID, alertMsg, true)

	if !success || updatedPoll == nil {
		return
	}

	// 3. Update dynamic keyboard and message body live!
	newMarkup := BuildPollKeyboard(updatedPoll, c.bot.Self.UserName)
	newText := FormatPollMessage(updatedPoll)

	if query.InlineMessageID != "" {
		// Poll was sent via inline query in a group or channel
		editMarkup := tgbotapi.EditMessageReplyMarkupConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: query.InlineMessageID,
				ReplyMarkup:     &newMarkup,
			},
		}
		_, _ = c.bot.Request(editMarkup)

		editText := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: query.InlineMessageID,
			},
			Text:      newText,
			ParseMode: "HTML",
		}
		_, _ = c.bot.Request(editText)
	} else if query.Message != nil {
		// Poll was sent directly in a chat with the bot
		editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, newText)
		editMsg.ParseMode = "HTML"
		editMsg.ReplyMarkup = &newMarkup
		_, _ = c.bot.Send(editMsg)
	}
}

// handleQuickPoll creates a quick Yes/No or Thumbs poll
func (c *CallbackHandler) handleQuickPoll(query *tgbotapi.CallbackQuery) {
	state, ok := c.wizard.GetState(query.From.ID)
	if !ok || state.Question == "" {
		c.answer(query.ID, "Session expired. Send /newpoll to start over.", true)
		return
	}

	var options []string
	if query.Data == "quick:thumbs" {
		options = []string{"👍 Agree", "👎 Disagree"}
	} else if query.Data == "quick:yesno" {
		options = []string{"✅ Yes", "❌ No"}
	}

	c.answer(query.ID, "Creating Quick Poll...", false)
	c.handler.CreateAndSendPoll(query.Message.Chat.ID, query.From, state.Question, options)
	c.wizard.Cancel(query.From.ID)

	// Clean up wizard prompt message
	del := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
	_, _ = c.bot.Request(del)
}

// refreshPoll updates poll message markup
func (c *CallbackHandler) refreshPoll(query *tgbotapi.CallbackQuery, pollID string) {
	poll, err := c.pollRepo.GetPoll(pollID)
	if err != nil || poll == nil {
		c.answer(query.ID, "Poll not found.", true)
		return
	}

	newMarkup := BuildPollKeyboard(poll, c.bot.Self.UserName)
	newText := FormatPollMessage(poll)

	if query.InlineMessageID != "" {
		editMarkup := tgbotapi.EditMessageReplyMarkupConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: query.InlineMessageID,
				ReplyMarkup:     &newMarkup,
			},
		}
		_, _ = c.bot.Request(editMarkup)
	} else if query.Message != nil {
		editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, newText)
		editMsg.ParseMode = "HTML"
		editMsg.ReplyMarkup = &newMarkup
		_, _ = c.bot.Send(editMsg)
	}
	c.answer(query.ID, "🔄 Poll refreshed!", false)
}

// managePollView shows poll management options
func (c *CallbackHandler) managePollView(query *tgbotapi.CallbackQuery, pollID string) {
	poll, err := c.pollRepo.GetPoll(pollID)
	if err != nil || poll == nil {
		c.answer(query.ID, "Poll not found.", true)
		return
	}

	c.answer(query.ID, "", false)

	status := "🟢 Active"
	if poll.IsClosed {
		status = "🔴 Closed"
	}

	text := fmt.Sprintf(
		"⚙️ <b>Manage Poll:</b>\n\n"+
			"<b>Question:</b> %s\n"+
			"<b>Status:</b> %s\n"+
			"<b>Total Votes:</b> %d\n"+
			"<b>ID:</b> <code>%s</code>\n",
		poll.Question, status, poll.TotalVotes, poll.ID,
	)

	var rows [][]tgbotapi.InlineKeyboardButton
	if !poll.IsClosed {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 Close Poll", fmt.Sprintf("close:%s", poll.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🗑️ Delete Poll", fmt.Sprintf("delete:%s", poll.ID)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Back to My Polls", "mypolls:0"),
	))

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	_, _ = c.bot.Send(msg)
}

// closePoll marks a poll closed
func (c *CallbackHandler) closePoll(query *tgbotapi.CallbackQuery, pollID string) {
	isAdmin := c.middleware.IsOwner(query.From.ID)
	err := c.pollRepo.ClosePoll(pollID, query.From.ID, isAdmin)
	if err != nil {
		c.answer(query.ID, "Failed to close poll or unauthorized.", true)
		return
	}
	c.answer(query.ID, "🛑 Poll has been closed!", true)
	c.refreshPoll(query, pollID)
}

// deletePoll deletes a poll permanently
func (c *CallbackHandler) deletePoll(query *tgbotapi.CallbackQuery, pollID string) {
	isAdmin := c.middleware.IsOwner(query.From.ID)
	err := c.pollRepo.DeletePoll(pollID, query.From.ID, isAdmin)
	if err != nil {
		c.answer(query.ID, "Failed to delete poll or unauthorized.", true)
		return
	}
	c.answer(query.ID, "🗑️ Poll deleted permanently.", true)
	if query.Message != nil {
		del := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
		_, _ = c.bot.Request(del)
	}
}

func (c *CallbackHandler) answer(queryID string, text string, showAlert bool) {
	callbackConfig := tgbotapi.NewCallback(queryID, text)
	callbackConfig.ShowAlert = showAlert
	_, _ = c.bot.Request(callbackConfig)
}
