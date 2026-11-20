package service

import (
	"testing"
	"time"

	"live-polling-app/backend/internal/models"
)

func TestValidateCreatePollRejectsDuplicateOptions(t *testing.T) {
	_, _, _, err := validateCreatePoll(CreatePollInput{
		Slug:     "team-lunch",
		Question: "Where should we eat?",
		Options:  []string{"Pizza", " pizza "},
	})
	if err != ErrPollInvalidInput {
		t.Fatalf("duplicate option error = %v", err)
	}
}

func TestValidateCreatePollTrimsOptionsAndAssignsStableIDs(t *testing.T) {
	_, _, options, err := validateCreatePoll(CreatePollInput{
		Slug:     "team-lunch",
		Question: "Where should we eat?",
		Options:  []string{" Pizza ", "Sushi"},
		ClosesAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("validateCreatePoll() error = %v", err)
	}
	if options[0].ID != "option-1" || options[0].Text != "Pizza" {
		t.Fatalf("first option = %+v", options[0])
	}
}

func TestPollIsOpenHonorsStatusAndExpiry(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	poll := models.Poll{Status: models.PollOpen, ClosesAt: &past}
	if poll.IsOpen(time.Now()) {
		t.Fatal("expired poll should not be open")
	}
}
