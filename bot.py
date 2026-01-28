from pyrogram import Client
from config import API_ID, API_HASH, BOT_TOKEN
from core.logger import setup_logger

setup_logger()

app = Client("StdVoteBot", api_id=API_ID, api_hash=API_HASH, bot_token=BOT_TOKEN)

import handlers.start
import handlers.vote
import handlers.callbacks
import handlers.stats
import handlers.cancel
import handlers.admin

print("Bot Started")
app.run()
