package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	PollOpen   = "open"
	PollClosed = "closed"
)

type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug      string        `bson:"slug" json:"slug"`
	Question  string        `bson:"question" json:"question"`
	Options   []PollOption  `bson:"options" json:"options"`
	OwnerID   bson.ObjectID `bson:"ownerId" json:"-"`
	Status    string        `bson:"status" json:"status"`
	ClosesAt  *time.Time    `bson:"closesAt,omitempty" json:"closesAt,omitempty"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

func (p Poll) IsOpen(now time.Time) bool {
	return p.Status == PollOpen && (p.ClosesAt == nil || now.Before(*p.ClosesAt))
}

func (p Poll) PublicView() Poll {
	p.OwnerID = bson.NilObjectID
	return p
}

func NormalizeOptionText(value string) string {
	return strings.TrimSpace(value)
}
