from pyrogram import filters
from bot import app
from utils.db import load_json

@app.on_message(filters.command("stats"))
async def stats_cmd(client, message):
    polls = load_json("data/polls.json", {})
    await message.reply(f"Total Polls: {len(polls)}")
