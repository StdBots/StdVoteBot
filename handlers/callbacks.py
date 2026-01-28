from pyrogram.types import CallbackQuery
from bot import app
from services.vote_service import VoteService

@app.on_callback_query()
async def handle_callbacks(client, query: CallbackQuery):
    if not query.data.startswith("vote:"):
        return

    poll_id = query.data.split(":")[1]
    success, result = VoteService.add_vote(poll_id, query.from_user.id)

    if success:
        await query.answer(f"Vote Count: {result}", show_alert=True)
    else:
        await query.answer(result, show_alert=True)
