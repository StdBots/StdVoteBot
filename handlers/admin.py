from pyrogram import filters
from bot import app
from config import ADMIN_ID

@app.on_message(filters.command("std"))
async def admin_cmd(client, message):
    if message.from_user.id != ADMIN_ID:
        return
    await message.reply("Admin access granted.")
