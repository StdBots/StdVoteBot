from pyrogram import filters
from bot import app

@app.on_message(filters.command("cancel"))
async def cancel_cmd(client, message):
    await message.reply("❌ Process Cancelled.")
