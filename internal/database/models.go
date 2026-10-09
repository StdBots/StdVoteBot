package database

import "time"

// PollOption represents a single option in a poll
type PollOption struct {
	ID         int    `bson:"id" json:"id"`
	Text       string `bson:"text" json:"text"`
	VotesCount int    `bson:"votes_count" json:"votes_count"`
}

// VoteRecord stores an individual user's vote
type VoteRecord struct {
	UserID    int64     `bson:"user_id" json:"user_id"`
	OptionID  int       `bson:"option_id" json:"option_id"`
	VotedAt   time.Time `bson:"voted_at" json:"voted_at"`
}

// Poll represents a voting poll instance
type Poll struct {
	ID              string       `bson:"_id" json:"id"`
	CreatorID       int64        `bson:"creator_id" json:"creator_id"`
	CreatorUsername string       `bson:"creator_username" json:"creator_username"`
	Question        string       `bson:"question" json:"question"`
	Options         []PollOption `bson:"options" json:"options"`
	TotalVotes      int          `bson:"total_votes" json:"total_votes"`
	Voters          []VoteRecord `bson:"voters" json:"voters"`
	IsClosed        bool         `bson:"is_closed" json:"is_closed"`
	AllowChangeVote bool         `bson:"allow_change_vote" json:"allow_change_vote"`
	RequireForceSub bool         `bson:"require_force_sub" json:"require_force_sub"`
	CreatedAt       time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time    `bson:"updated_at" json:"updated_at"`
}

// User represents a registered bot user
type User struct {
	ID                int64     `bson:"_id" json:"id"`
	Username          string    `bson:"username" json:"username"`
	FirstName         string    `bson:"first_name" json:"first_name"`
	LastName          string    `bson:"last_name" json:"last_name"`
	TotalPollsCreated int       `bson:"total_polls_created" json:"total_polls_created"`
	TotalVotesCast    int       `bson:"total_votes_cast" json:"total_votes_cast"`
	JoinedAt          time.Time `bson:"joined_at" json:"joined_at"`
	LastActive        time.Time `bson:"last_active" json:"last_active"`
}

// WizardState represents a user's progress creating a poll
type WizardState struct {
	Step     int      // 1: Question, 2: Options
	Question string
	Options  []string
}
