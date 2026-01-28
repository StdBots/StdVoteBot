from pyrogram import filters
from bot import app
from core.states import user_states, SETTING_CHANNEL

@app.on_message(filters.command("vote"))
async def vote_start(client, message):
    user_states[message.from_user.id] = SETTING_CHANNEL
    await message.reply("Send channel username with @")

@app.on_message(filters.text)
async def receive_channel(client, message):
    if user_states.get(message.from_user.id) != SETTING_CHANNEL:
        return

    channel = message.text.strip()
    user_states.pop(message.from_user.id, None)

    await message.reply(f"Poll will be created in {channel}")
