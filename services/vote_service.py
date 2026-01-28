from utils.db import load_json, save_json

POLL_FILE = "data/polls.json"

class VoteService:

    @staticmethod
    def add_vote(poll_id, user_id):
        polls = load_json(POLL_FILE, {})
        poll = polls.get(poll_id)

        if not poll:
            return False, "Poll not found"

        if user_id in poll["voters"]:
            return False, "Already voted"

        poll["votes"] += 1
        poll["voters"].append(user_id)
        save_json(POLL_FILE, polls)

        return True, poll["votes"]
