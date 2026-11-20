package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrPollInvalidInput = errors.New("invalid poll input")
	ErrPollClosed       = errors.New("poll is closed")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type PollService struct {
	polls *repository.PollRepository
}

type CreatePollInput struct {
	Slug     string
	Question string
	Options  []string
	ClosesAt *time.Time
}

func NewPollService(polls *repository.PollRepository) *PollService {
	return &PollService{polls: polls}
}

func (s *PollService) Create(ctx context.Context, ownerID bson.ObjectID, input CreatePollInput) (models.Poll, error) {
	slug, question, options, err := validateCreatePoll(input)
	if err != nil {
		return models.Poll{}, err
	}
	poll := models.Poll{
		ID:        bson.NewObjectID(),
		Slug:      slug,
		Question:  question,
		Options:   options,
		OwnerID:   ownerID,
		Status:    models.PollOpen,
		ClosesAt:  input.ClosesAt,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.polls.Create(ctx, poll); err != nil {
		return models.Poll{}, err
	}
	return poll, nil
}

func (s *PollService) ListMine(ctx context.Context, ownerID bson.ObjectID) ([]models.Poll, error) {
	return s.polls.FindByOwner(ctx, ownerID)
}

func (s *PollService) PublicBySlug(ctx context.Context, slug string) (models.Poll, error) {
	if !slugPattern.MatchString(slug) {
		return models.Poll{}, repository.ErrPollNotFound
	}
	poll, err := s.polls.FindBySlug(ctx, slug)
	if err != nil {
		return models.Poll{}, err
	}
	return poll.PublicView(), nil
}

func (s *PollService) Close(ctx context.Context, ownerID bson.ObjectID, slug string) error {
	return s.polls.Close(ctx, slug, ownerID)
}

func (s *PollService) Delete(ctx context.Context, ownerID bson.ObjectID, slug string) error {
	return s.polls.Delete(ctx, slug, ownerID)
}

func validateCreatePoll(input CreatePollInput) (string, string, []models.PollOption, error) {
	slug := strings.TrimSpace(input.Slug)
	question := strings.TrimSpace(input.Question)
	if len(slug) < 3 || len(slug) > 80 || !slugPattern.MatchString(slug) {
		return "", "", nil, ErrPollInvalidInput
	}
	if len(question) < 3 || len(question) > 200 {
		return "", "", nil, ErrPollInvalidInput
	}
	if len(input.Options) < 2 || len(input.Options) > 10 {
		return "", "", nil, ErrPollInvalidInput
	}
	if input.ClosesAt != nil && !input.ClosesAt.After(time.Now()) {
		return "", "", nil, ErrPollInvalidInput
	}

	options := make([]models.PollOption, 0, len(input.Options))
	seen := make(map[string]struct{}, len(input.Options))
	for index, rawOption := range input.Options {
		text := models.NormalizeOptionText(rawOption)
		key := strings.ToLower(text)
		if len(text) < 1 || len(text) > 100 {
			return "", "", nil, ErrPollInvalidInput
		}
		if _, exists := seen[key]; exists {
			return "", "", nil, ErrPollInvalidInput
		}
		seen[key] = struct{}{}
		options = append(options, models.PollOption{
			ID:   optionID(index),
			Text: text,
		})
	}
	return slug, question, options, nil
}

func optionID(index int) string {
	return "option-" + strconv.Itoa(index+1)
}
