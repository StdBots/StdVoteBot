from utils.db import load_json, save_json

POLL_FILE = "data/polls.json"

class PollService:

    @staticmethod
    def create_poll(poll_id, channel, owner):
        polls = load_json(POLL_FILE, {})
        polls[poll_id] = {
            "channel": channel,
            "owner": owner,
            "votes": 0,
            "voters": []
        }
        save_json(POLL_FILE, polls)

    @staticmethod
    def get_poll(poll_id):
        return load_json(POLL_FILE, {}).get(poll_id)
