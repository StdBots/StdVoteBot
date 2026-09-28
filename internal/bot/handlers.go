package bot

import (
	"fmt"
	"strings"

	"github.com/StdBots/StdVoteBot/internal/credit"
	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler manages bot commands and user messages
type Handler struct {
	bot        *tgbotapi.BotAPI
	pollRepo   *database.PollRepo
	userRepo   *database.UserRepo
	wizard     *WizardManager
	middleware *Middleware
}

// NewHandler initializes a new Handler
func NewHandler(bot *tgbotapi.BotAPI, pollRepo *database.PollRepo, userRepo *database.UserRepo, wizard *WizardManager, middleware *Middleware) *Handler {
	return &Handler{
		bot:        bot,
		pollRepo:   pollRepo,
		userRepo:   userRepo,
		wizard:     wizard,
		middleware: middleware,
	}
}

// HandleStart handles the /start command (including deep links)
func (h *Handler) HandleStart(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	args := strings.TrimSpace(msg.CommandArguments())
	if args != "" {
		// Deep link: /start poll_<id> or /start <id>
		pollID := strings.TrimPrefix(args, "poll_")
		poll, err := h.pollRepo.GetPoll(pollID)
		if err == nil && poll != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, FormatPollMessage(poll))
			reply.ParseMode = "HTML"
			reply.ReplyMarkup = BuildPollKeyboard(poll, h.bot.Self.UserName)
			_, _ = h.bot.Send(reply)
			return
		}
	}

	welcomeText := fmt.Sprintf(
		"🗳️ <b>Welcome to StdVoteBot!</b>\n\n"+
			"The most powerful, concurrent voting & poll bot on Telegram.\n\n"+
			"✨ <b>Features:</b>\n"+
			"• Real-time live button vote updating\n"+
			"• Share polls anywhere via Inline Queries (<code>@%s</code>)\n"+
			"• Multi-option custom polls or instant Quick Polls\n"+
			"• Anti-cheat one-vote protection\n"+
			"• Fast MongoDB persistence\n\n"+
			"⚡ <i>Engineered by STD DEEPANSHU (%s)</i>\n%s",
		h.bot.Self.UserName,
		credit.Domain,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(welcomeText))
	reply.ParseMode = "HTML"

	btnNewPoll := tgbotapi.NewInlineKeyboardButtonData("➕ Create Poll", "wizard:start")
	btnMyPolls := tgbotapi.NewInlineKeyboardButtonData("📊 My Polls", "mypolls:0")
	btnChannel := tgbotapi.NewInlineKeyboardButtonURL("📢 Updates", "https://t.me/StdBots")
	btnDev := tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Developer", "https://"+credit.Domain)

	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnNewPoll, btnMyPolls),
		tgbotapi.NewInlineKeyboardRow(btnChannel, btnDev),
	)

	_, _ = h.bot.Send(reply)
}

// HandleHelp handles the /help command
func (h *Handler) HandleHelp(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	helpText := fmt.Sprintf(
		"📖 <b>StdVoteBot Help & Guide</b>\n\n"+
			"<b>Commands:</b>\n"+
			"• /start — Launch bot and view main menu\n"+
			"• /newpoll — Launch interactive poll creator\n"+
			"• /mypolls — View and manage your created polls\n"+
			"• /cancel — Abort active poll creation\n"+
			"• /help — Show this help message\n\n"+
			"<b>💡 How to Share Polls:</b>\n"+
			"1. Create a poll using /newpoll.\n"+
			"2. In any chat, type: <code>@%s &lt;poll_id&gt;</code>\n"+
			"3. Tap on the preview to send the live interactive poll into the chat!\n\n"+
			"%s",
		h.bot.Self.UserName,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(helpText))
	reply.ParseMode = "HTML"
	_, _ = h.bot.Send(reply)
}

// HandleNewPoll begins the poll creation wizard
func (h *Handler) HandleNewPoll(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)
	h.wizard.StartWizard(msg.From.ID)

	prompt := "📝 <b>Step 1 of 2: Poll Question</b>\n\n" +
		"Please send the <b>Question or Title</b> for your poll.\n" +
		"<i>(Or send /cancel to abort at any time)</i>"

	reply := tgbotapi.NewMessage(msg.Chat.ID, prompt)
	reply.ParseMode = "HTML"
	_, _ = h.bot.Send(reply)
}

// HandleCancel cancels active operations
func (h *Handler) HandleCancel(msg *tgbotapi.Message) {
	h.wizard.Cancel(msg.From.ID)

	reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Operation cancelled. You are back to the main menu.")
	_, _ = h.bot.Send(reply)
}

// HandleMyPolls lists the user's polls
func (h *Handler) HandleMyPolls(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	polls, err := h.pollRepo.GetUserPolls(msg.From.ID, 10)
	if err != nil || len(polls) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "ℹ️ You haven't created any polls yet.\nSend /newpoll to create your first poll!")
		btnNew := tgbotapi.NewInlineKeyboardButtonData("➕ Create Poll", "wizard:start")
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btnNew))
		_, _ = h.bot.Send(reply)
		return
	}

	var sb strings.Builder
	sb.WriteString("📊 <b>Your Recent Polls:</b>\n\n")

	var rows [][]tgbotapi.InlineKeyboardButton
	for idx, p := range polls {
		status := "🟢 Active"
		if p.IsClosed {
			status = "🔴 Closed"
		}
		sb.WriteString(fmt.Sprintf("%d. <b>%s</b>\n   Status: %s | Votes: %d\n", idx+1, p.Question, status, p.TotalVotes))

		btnView := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🔍 #%d %s", idx+1, p.Question[:min(len(p.Question), 15)]), fmt.Sprintf("view:%s", p.ID))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnView))
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, sb.String())
	reply.ParseMode = "HTML"
	reply.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	_, _ = h.bot.Send(reply)
}

// HandleTextMessage processes interactive steps in the wizard
func (h *Handler) HandleTextMessage(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	state, ok := h.wizard.GetState(msg.From.ID)
	if !ok {
		return // Ignore normal text messages not in wizard
	}

	if state.Step == 1 {
		// User sent question
		question := strings.TrimSpace(msg.Text)
		if len(question) < 3 {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Question is too short. Please send a valid question:")
			_, _ = h.bot.Send(reply)
			return
		}

		h.wizard.SetQuestion(msg.From.ID, question)

		prompt := fmt.Sprintf(
			"✅ <b>Question Set:</b>\n<i>\"%s\"</i>\n\n"+
				"📝 <b>Step 2 of 2: Poll Options</b>\n\n"+
				"Send your options separated by <b>new lines</b> (2 to 10 options).\n\n"+
				"<b>Example:</b>\n"+
				"Option A\n"+
				"Option B\n"+
				"Option C\n\n"+
				"<i>Or pick an instant quick template below:</i>",
			question,
		)

		reply := tgbotapi.NewMessage(msg.Chat.ID, prompt)
		reply.ParseMode = "HTML"

		btnThumbs := tgbotapi.NewInlineKeyboardButtonData("👍 / 👎 Quick Thumbs", "quick:thumbs")
		btnYesNo := tgbotapi.NewInlineKeyboardButtonData("✅ / ❌ Yes / No", "quick:yesno")
		btnCancel := tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "wizard:cancel")

		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btnThumbs, btnYesNo),
			tgbotapi.NewInlineKeyboardRow(btnCancel),
		)

		_, _ = h.bot.Send(reply)
		return
	}

	if state.Step == 2 {
		// User sent options list
		lines := strings.Split(msg.Text, "\n")
		var options []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if trimmed != "" {
				options = append(options, trimmed)
			}
		}

		if len(options) < 2 {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Please provide at least 2 options (separated by newlines):")
			_, _ = h.bot.Send(reply)
			return
		}
		if len(options) > 10 {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ A maximum of 10 options are supported. Please send again with 10 or fewer options:")
			_, _ = h.bot.Send(reply)
			return
		}

		h.CreateAndSendPoll(msg.Chat.ID, msg.From, state.Question, options)
		h.wizard.Cancel(msg.From.ID)
	}
}

// CreateAndSendPoll builds poll in DB and outputs the live message
func (h *Handler) CreateAndSendPoll(chatID int64, from *tgbotapi.User, question string, optionTexts []string) {
	var pollOptions []database.PollOption
	for idx, text := range optionTexts {
		pollOptions = append(pollOptions, database.PollOption{
			ID:         idx,
			Text:       text,
			VotesCount: 0,
		})
	}

	newPoll := &database.Poll{
		CreatorID:       from.ID,
		CreatorUsername: from.UserName,
		Question:        question,
		Options:         pollOptions,
		TotalVotes:      0,
		IsClosed:        false,
		AllowChangeVote: true,
	}

	err := h.pollRepo.CreatePoll(newPoll)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "❌ Failed to create poll in database. Please try again.")
		_, _ = h.bot.Send(reply)
		return
	}

	// Send poll preview
	msg := tgbotapi.NewMessage(chatID, FormatPollMessage(newPoll))
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = BuildPollKeyboard(newPoll, h.bot.Self.UserName)

	_, _ = h.bot.Send(msg)

	// Send helper tips
	tips := fmt.Sprintf(
		"🎉 <b>Poll Created Successfully!</b>\n\n"+
			"🆔 <b>Poll ID:</b> <code>%s</code>\n\n"+
			"📢 <b>How to Share:</b>\n"+
			"• Tap <b>'🔗 Share Poll in Chat'</b> button to post this poll directly to any group or channel!\n"+
			"• Or type <code>@%s %s</code> in any chat.",
		newPoll.ID,
		h.bot.Self.UserName,
		newPoll.ID,
	)
	tipMsg := tgbotapi.NewMessage(chatID, tips)
	tipMsg.ParseMode = "HTML"
	_, _ = h.bot.Send(tipMsg)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
