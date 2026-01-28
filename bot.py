from pyrogram import Client
from config import API_ID, API_HASH, BOT_TOKEN
from core.logger import setup_logger

setup_logger()

app = Client("StdVoteBot", api_id=API_ID, api_hash=API_HASH, bot_token=BOT_TOKEN)

from handlers import start, vote, callbacks, stats, cancel, admin

print("Bot Started")
app.run()
