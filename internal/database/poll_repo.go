package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPollNotFound = errors.New("poll not found")
	ErrPollClosed   = errors.New("this poll is closed")
	ErrAlreadyVoted = errors.New("you have already voted in this poll")
)

// PollRepo handles database operations for polls
type PollRepo struct {
	coll  *mongo.Collection
	users *mongo.Collection
	mu    sync.Mutex
}

// NewPollRepo creates a new Poll repository
func NewPollRepo(db *MongoDB) *PollRepo {
	return &PollRepo{
		coll:  db.Polls,
		users: db.Users,
	}
}

// GeneratePollID creates an 8-character alphanumeric unique ID
func GeneratePollID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())[:8]
	}
	return hex.EncodeToString(bytes)
}

// CreatePoll inserts a new poll into MongoDB
func (r *PollRepo) CreatePoll(poll *Poll) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if poll.ID == "" {
		poll.ID = GeneratePollID()
	}
	poll.CreatedAt = time.Now()
	poll.UpdatedAt = time.Now()
	if poll.Voters == nil {
		poll.Voters = []VoteRecord{}
	}

	_, err := r.coll.InsertOne(ctx, poll)
	if err != nil {
		return fmt.Errorf("failed to insert poll: %w", err)
	}

	// Increment user's total polls created
	_, _ = r.users.UpdateOne(ctx,
		bson.M{"_id": poll.CreatorID},
		bson.M{"$inc": bson.M{"total_polls_created": 1}},
	)

	return nil
}

// GetPoll fetches a poll by ID
func (r *PollRepo) GetPoll(pollID string) (*Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll Poll
	err := r.coll.FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPollNotFound
		}
		return nil, err
	}
	return &poll, nil
}

// CastVote handles voting with anti-cheat and optional vote-switching
func (r *PollRepo) CastVote(pollID string, userID int64, optionID int) (bool, string, *Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var poll Poll
	err := r.coll.FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, "❌ Poll does not exist.", nil, ErrPollNotFound
		}
		return false, "❌ Database error.", nil, err
	}

	if poll.IsClosed {
		return false, "⚠️ This poll is already closed.", &poll, ErrPollClosed
	}

	if optionID < 0 || optionID >= len(poll.Options) {
		return false, "❌ Invalid poll option.", &poll, errors.New("invalid option")
	}

	// Check previous vote by this user
	existingVoteIndex := -1
	for idx, v := range poll.Voters {
		if v.UserID == userID {
			existingVoteIndex = idx
			break
		}
	}

	if existingVoteIndex != -1 {
		prevOptionID := poll.Voters[existingVoteIndex].OptionID

		// User clicked the exact same option they already voted for
		if prevOptionID == optionID {
			return false, "ℹ️ You already voted for this option!", &poll, nil
		}

		// User voted for a different option
		if !poll.AllowChangeVote {
			return false, "⚠️ You have already voted and cannot change your vote.", &poll, ErrAlreadyVoted
		}

		// User is changing their vote: decrement old option, increment new option
		poll.Options[prevOptionID].VotesCount--
		if poll.Options[prevOptionID].VotesCount < 0 {
			poll.Options[prevOptionID].VotesCount = 0
		}
		poll.Options[optionID].VotesCount++
		poll.Voters[existingVoteIndex].OptionID = optionID
		poll.Voters[existingVoteIndex].VotedAt = time.Now()
		poll.UpdatedAt = time.Now()

		_, err = r.coll.ReplaceOne(ctx, bson.M{"_id": pollID}, poll)
		if err != nil {
			return false, "❌ Failed to update vote.", nil, err
		}

		msg := fmt.Sprintf("✅ Your vote was changed to: %s", poll.Options[optionID].Text)
		return true, msg, &poll, nil
	}

	// Fresh new vote
	poll.Options[optionID].VotesCount++
	poll.TotalVotes++
	poll.Voters = append(poll.Voters, VoteRecord{
		UserID:   userID,
		OptionID: optionID,
		VotedAt:  time.Now(),
	})
	poll.UpdatedAt = time.Now()

	_, err = r.coll.ReplaceOne(ctx, bson.M{"_id": pollID}, poll)
	if err != nil {
		return false, "❌ Failed to cast vote.", nil, err
	}

	// Track total votes cast for user
	_, _ = r.users.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"total_votes_cast": 1}},
	)

	msg := fmt.Sprintf("✅ Vote recorded for: %s", poll.Options[optionID].Text)
	return true, msg, &poll, nil
}

// GetUserPolls returns polls created by a specific user
func (r *PollRepo) GetUserPolls(creatorID int64, limit int64) ([]*Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cursor, err := r.coll.Find(ctx, bson.M{"creator_id": creatorID}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []*Poll
	if err = cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

// ClosePoll marks a poll as closed
func (r *PollRepo) ClosePoll(pollID string, requesterID int64, isAdmin bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": pollID}
	if !isAdmin {
		filter["creator_id"] = requesterID
	}

	res, err := r.coll.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"is_closed":  true,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}

// DeletePoll deletes a poll permanently
func (r *PollRepo) DeletePoll(pollID string, requesterID int64, isAdmin bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": pollID}
	if !isAdmin {
		filter["creator_id"] = requesterID
	}

	res, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}

// CountTotalPolls returns the total number of polls
func (r *PollRepo) CountTotalPolls() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.coll.CountDocuments(ctx, bson.M{})
}

// CountTotalVotes aggregates total votes across all polls
func (r *PollRepo) CountTotalVotes() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$total_votes"}}},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int64 `bson:"total"`
	}
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return 0, nil
	}
	return results[0].Total, nil
}
