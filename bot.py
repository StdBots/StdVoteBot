import asyncio
import signal
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


async def main():
    await app.start()
    print("StdVoteBot Started 🚀")
    await idle()
    await app.stop()


async def idle():
    stop = asyncio.Event()

    def signal_handler(*_):
        stop.set()

    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    await stop.wait()


if __name__ == "__main__":
    asyncio.run(main())
