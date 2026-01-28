from pyrogram import filters
from bot import app
from config import UPDATE_CHANNEL
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton

@app.on_message(filters.command("start"))
async def start_cmd(client, message):
    keyboard = InlineKeyboardMarkup([
        [InlineKeyboardButton("Join Channel", url=f"https://t.me/{UPDATE_CHANNEL}")]
    ])
    await message.reply("Welcome to StdVoteBot 🚀", reply_markup=keyboard)
