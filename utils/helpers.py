import random, string

def generate_vote_id(user_id):
    return ''.join(random.choices(string.ascii_uppercase + string.digits, k=3)) + str(user_id)
