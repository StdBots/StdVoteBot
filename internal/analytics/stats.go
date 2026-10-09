package analytics

import (
	"fmt"
	"runtime"
	"time"

	"github.com/StdBots/StdVoteBot/internal/credit"
	"github.com/StdBots/StdVoteBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var startTime = time.Now()

// Engine provides metrics and reporting
type Engine struct {
	api      *tgbotapi.BotAPI
	pollRepo *database.PollRepo
	userRepo *database.UserRepo
	ownerID  int64
}

// NewEngine creates a new analytics engine
func NewEngine(api *tgbotapi.BotAPI, pollRepo *database.PollRepo, userRepo *database.UserRepo, ownerID int64) *Engine {
	return &Engine{
		api:      api,
		pollRepo: pollRepo,
		userRepo: userRepo,
		ownerID:  ownerID,
	}
}

// HandleStats renders the system and bot statistics
func (e *Engine) HandleStats(msg *tgbotapi.Message) {
	if msg.From.ID != e.ownerID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ Unauthorized: Admin statistics only.")
		_, _ = e.api.Send(reply)
		return
	}

	totalUsers, _ := e.userRepo.CountUsers()
	totalPolls, _ := e.pollRepo.CountTotalPolls()
	totalVotes, _ := e.pollRepo.CountTotalVotes()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime).Round(time.Second)

	statsText := fmt.Sprintf(
		"📈 <b>StdVoteBot Statistics & Health:</b>\n\n"+
			"👥 <b>Total Users:</b> %d\n"+
			"🗳️ <b>Total Polls:</b> %d\n"+
			"🔢 <b>Total Votes Cast:</b> %d\n"+
			"⏱️ <b>Uptime:</b> %s\n\n"+
			"⚙️ <b>System Metrics:</b>\n"+
			"• <b>Go Version:</b> %s\n"+
			"• <b>Goroutines:</b> %d\n"+
			"• <b>Memory Alloc:</b> %.2f MB\n"+
			"• <b>Memory Sys:</b> %.2f MB\n"+
			"• <b>Garbage Collections:</b> %d\n\n"+
			"%s",
		totalUsers,
		totalPolls,
		totalVotes,
		uptime,
		runtime.Version(),
		runtime.NumGoroutine(),
		float64(m.Alloc)/(1024*1024),
		float64(m.Sys)/(1024*1024),
		m.NumGC,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(statsText))
	reply.ParseMode = "HTML"

	btnRefresh := tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh Stats", "stats:refresh")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btnRefresh))

	_, _ = e.api.Send(reply)
}
