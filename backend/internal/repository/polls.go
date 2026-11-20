package repository

import (
	"context"
	"errors"

	"live-polling-app/backend/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrPollNotFound = errors.New("poll not found")
	ErrSlugTaken    = errors.New("slug already exists")
)

type PollRepository struct {
	polls *mongo.Collection
	votes *mongo.Collection
}

func NewPollRepository(database *mongo.Database) *PollRepository {
	return &PollRepository{
		polls: database.Collection("polls"),
		votes: database.Collection("votes"),
	}
}

func (r *PollRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.polls.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "ownerId", Value: 1}}},
	}); err != nil {
		return err
	}
	_, err := r.votes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *PollRepository) Create(ctx context.Context, poll models.Poll) error {
	_, err := r.polls.InsertOne(ctx, poll)
	if isDuplicateKey(err) {
		return ErrSlugTaken
	}
	return err
}

func (r *PollRepository) FindBySlug(ctx context.Context, slug string) (models.Poll, error) {
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.D{{Key: "slug", Value: slug}}).Decode(&poll)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Poll{}, ErrPollNotFound
	}
	return poll, err
}

func (r *PollRepository) FindByOwner(ctx context.Context, ownerID bson.ObjectID) ([]models.Poll, error) {
	cursor, err := r.polls.Find(ctx, bson.D{{Key: "ownerId", Value: ownerID}}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) Close(ctx context.Context, slug string, ownerID bson.ObjectID) error {
	result, err := r.polls.UpdateOne(ctx,
		bson.D{{Key: "slug", Value: slug}, {Key: "ownerId", Value: ownerID}, {Key: "status", Value: models.PollOpen}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: models.PollClosed}}}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}

func (r *PollRepository) Delete(ctx context.Context, slug string, ownerID bson.ObjectID) error {
	result, err := r.polls.DeleteOne(ctx, bson.D{{Key: "slug", Value: slug}, {Key: "ownerId", Value: ownerID}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var writeErr mongo.WriteException
	return errors.As(err, &writeErr) && writeErr.HasErrorCode(11000)
}
