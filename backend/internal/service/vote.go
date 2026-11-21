package service

import (
	"context"
	"errors"
	"time"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/realtime"
	"live-polling-app/backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrVoteInvalid = errors.New("invalid vote")

type VoteService struct {
	polls *repository.PollRepository
	redis *realtime.RedisStore
}

type PollResults struct {
	Poll   models.Poll      `json:"poll"`
	Counts map[string]int64 `json:"counts"`
	Total  int64            `json:"total"`
}

func NewVoteService(polls *repository.PollRepository, redis *realtime.RedisStore) *VoteService {
	return &VoteService{polls: polls, redis: redis}
}

func (s *VoteService) Results(ctx context.Context, slug string) (PollResults, error) {
	poll, err := s.polls.FindBySlug(ctx, slug)
	if err != nil {
		return PollResults{}, err
	}
	counts, err := s.counts(ctx, poll)
	if err != nil {
		return PollResults{}, err
	}
	return buildResults(poll, counts), nil
}

func (s *VoteService) Vote(ctx context.Context, slug, optionID, voterID string) (PollResults, error) {
	poll, err := s.polls.FindBySlug(ctx, slug)
	if err != nil {
		return PollResults{}, err
	}
	if !poll.IsOpen(time.Now()) {
		return PollResults{}, ErrPollClosed
	}
	if !hasOption(poll, optionID) || voterID == "" {
		return PollResults{}, ErrVoteInvalid
	}
	redisResult, err := s.redis.ApplyVote(ctx, poll.ID.Hex(), optionID, voterID)
	if errors.Is(err, realtime.ErrDuplicateVoter) {
		return PollResults{}, realtime.ErrDuplicateVoter
	}
	if err != nil {
		return PollResults{}, err
	}
	if err := s.polls.CreateVote(ctx, newVote(poll.ID, optionID, voterID)); err != nil {
		rollbackErr := s.redis.RollbackVote(ctx, poll.ID.Hex(), optionID, voterID)
		if errors.Is(err, repository.ErrVoteAlreadyExists) {
			if rollbackErr != nil {
				return PollResults{}, errors.Join(realtime.ErrDuplicateVoter, rollbackErr)
			}
			return PollResults{}, realtime.ErrDuplicateVoter
		}
		if rollbackErr != nil {
			return PollResults{}, errors.Join(err, rollbackErr)
		}
		return PollResults{}, err
	}
	return buildResults(poll, redisResult.Counts), nil
}

func (s *VoteService) counts(ctx context.Context, poll models.Poll) (map[string]int64, error) {
	counts, err := s.redis.GetCounts(ctx, poll.ID.Hex())
	if err != nil {
		return nil, err
	}
	if !hasAllOptionCounts(poll, counts) {
		counts, err = s.polls.FindVoteCounts(ctx, poll.ID)
		if err != nil {
			return nil, err
		}
		if err := s.redis.SetCounts(ctx, poll.ID.Hex(), counts); err != nil {
			return nil, err
		}
	}
	return counts, nil
}

func hasAllOptionCounts(poll models.Poll, counts map[string]int64) bool {
	for _, option := range poll.Options {
		if _, ok := counts[option.ID]; !ok {
			return false
		}
	}
	return true
}

func buildResults(poll models.Poll, counts map[string]int64) PollResults {
	normalized := make(map[string]int64, len(poll.Options))
	var total int64
	for _, option := range poll.Options {
		count := counts[option.ID]
		normalized[option.ID] = count
		total += count
	}
	return PollResults{Poll: poll.PublicView(), Counts: normalized, Total: total}
}

func hasOption(poll models.Poll, optionID string) bool {
	for _, option := range poll.Options {
		if option.ID == optionID {
			return true
		}
	}
	return false
}

func newVote(pollID bson.ObjectID, optionID, voterID string) models.Vote {
	return models.Vote{ID: bson.NewObjectID(), PollID: pollID, OptionID: optionID, VoterID: voterID, CreatedAt: time.Now().UTC()}
}
