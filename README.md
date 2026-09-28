# 🗳️ StdVoteBot — Ultra-Fast Telegram Voting & Poll Bot

<p align="center">
  <img src="https://graph.org/file/00ea4effe5d2dfbb8d5be.jpg" alt="StdVoteBot Banner" width="450"/>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Language-Go%201.22+-00ADD8?style=for-the-badge&logo=go" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL%20v3-blue?style=for-the-badge" alt="License"/></a>
  <a href="https://t.me/STDBOTS"><img src="https://img.shields.io/badge/Channel-%40STDBOTS-2CA5E0?style=for-the-badge&logo=telegram" alt="Telegram Channel"/></a>
  <a href="https://deepanshu.in"><img src="https://img.shields.io/badge/Author-STD%20DEEPANSHU-FF4500?style=for-the-badge" alt="Author"/></a>
</p>

<p align="center">
  <a href="https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdVoteBot">
    <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-7056bf?style=for-the-badge&logo=heroku" alt="Deploy to Heroku"/>
  </a>
  <a href="https://railway.app/template/new?template=https://github.com/StdBots/StdVoteBot">
    <img src="https://img.shields.io/badge/Deploy%20On%20Railway-0B0D0E?style=for-the-badge&logo=railway" alt="Deploy on Railway"/>
  </a>
</p>

---

## ⚡ Overview

**StdVoteBot** is a high-concurrency Telegram Voting and Poll Bot engineered in **Go (Golang)** with **MongoDB** persistence. It replaces slow and error-prone polling scripts with atomic counter operations, real-time dynamic button label updates, anti-cheat protection, inline query sharing anywhere on Telegram, and channel force-subscription checks.

Developed by **[STD DEEPANSHU](https://deepanshu.in)** as part of the **[STD BOTS Ecosystem](https://t.me/STDBOTS)**.

---

## ✨ Features

- ⚡ **Go 1.22+ Architecture:** Blazing speed, minimal memory usage (<25MB RAM).
- 🍃 **Atomic Concurrency:** MongoDB `$inc` and `$addToSet` ensure zero race conditions on concurrent votes.
- 🎨 **Real-Time Dynamic Buttons:** Button vote counters update live as users cast votes (`Option A [45]`, `Option B [12]`).
- 🔗 **Universal Inline Sharing:** Type `@StdVoteBot <poll_id>` in any group, channel, or direct message to publish interactive polls.
- 🛡️ **Anti-Cheat System:** One vote per user, duplicate detection, and optional vote toggle (switching vote).
- 📢 **Force Subscribe (FSUB):** Verify channel membership before users can vote.
- 📊 **Poll Creator Dashboard (`/mypolls`):** Real-time voter metrics, close/reopen polls, or delete polls.
- 📣 **Admin Broadcast Engine:** Ultra-fast bulk broadcast to all bot users with live progress.
- 🔒 **7-Layer Credit Protection:** AGPL-3.0 integrity validation, zero-width watermarks, and brand protection.

---

## 🚀 One-Click Deployments

### 🟣 Deploy to Heroku
Click the button below to deploy your instance to Heroku in 60 seconds:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdVoteBot)

### 🚂 Deploy on Railway
Click the button below to deploy on Railway with container support:

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/new?template=https://github.com/StdBots/StdVoteBot)

### 🖥️ 1-Command VPS Deployment (Linux / Ubuntu / Debian)
Run this single command on your VPS as root:
```bash
curl -fsSL https://raw.githubusercontent.com/StdBots/StdVoteBot/main/scripts/install_vps.sh | bash
```

---

## 🐳 Docker & Manual VPS Setup

```bash
# 1. Clone repository
git clone https://github.com/StdBots/StdVoteBot.git
cd StdVoteBot

# 2. Configure environment
cp .env.example .env
nano .env

# 3. Start with Docker Compose
docker compose up -d --build
```

---

## ⚙️ Environment Variables

| Variable | Description | Required | Default |
|---|---|---|---|
| `BOT_TOKEN` | Telegram Bot Token from [@BotFather](https://t.me/BotFather) | **Yes** | — |
| `OWNER_ID` | Telegram User ID of the primary administrator | **Yes** | `7394590844` |
| `MONGO_URI` | MongoDB Connection String (Atlas or Local) | **Yes** | — |
| `DB_NAME` | Database name | No | `stdvotebot` |
| `FORCE_SUB_CHANNEL` | Channel username without `@` for force-sub | No | `StdBots` |
| `LOG_CHANNEL_ID` | Telegram Channel ID for logging events | No | `0` |
| `ENV` | Environment mode (`development`/`production`) | No | `production` |

---

## 🤖 Commands

| Command | Description |
|---|---|
| `/start` | Launch the bot and view features & credits |
| `/newpoll` | Start interactive poll creation wizard |
| `/mypolls` | View your active polls, analytics, or close them |
| `/cancel` | Cancel current wizard operation |
| `/help` | Detailed help guide |
| `/stats` | Global bot analytics (Admin only) |
| `/broadcast` | Broadcast message to all registered users (Admin only) |

---

## 📄 License & Attribution

Licensed under the [GNU Affero General Public License v3 (AGPL-3.0)](LICENSE).

Mandatory Attribution: Derivative works, forks, and hosted instances must preserve all visible and embedded credits pointing to **STD DEEPANSHU** ([https://deepanshu.in](https://deepanshu.in)) and **STD BOTS** ([@STDBOTS](https://t.me/STDBOTS)).
