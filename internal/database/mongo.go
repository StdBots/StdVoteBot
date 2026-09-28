/*
 * Copyright (C) 2024-2026 STD DEEPANSHU <https://deepanshu.in>
 * STD BOTS - Telegram: @STD_DEEPANSHU, @STDBOTS
 *
 * This file is part of StdVoteBot.
 * Licensed under AGPL-3.0.
 */

package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds client, database, and collections for StdVoteBot
type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
	Polls  *mongo.Collection
	Users  *mongo.Collection
}

// NewMongoDB connects to MongoDB and initializes indexes
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(100).SetMinPoolSize(10)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping database
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(dbName)
	m := &MongoDB{
		client: client,
		db:     db,
		Polls:  db.Collection("polls"),
		Users:  db.Collection("users"),
	}

	// Create Indexes
	m.initIndexes()

	log.Printf("Connected to MongoDB successfully: db=%s", dbName)
	return m, nil
}

// initIndexes ensures optimal indexes for fast queries
func (m *MongoDB) initIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Polls index on creator_id and created_at
	_, _ = m.Polls.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "creator_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "voters.user_id", Value: 1}},
		},
	})

	// Users index
	_, _ = m.Users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "joined_at", Value: -1}},
	})
}

// Close disconnects MongoDB
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}
