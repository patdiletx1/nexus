package audit

import "time"

type Event struct {
	ID          string
	CompanyID   string
	ActorUserID string
	EventType   string
	EntityType  string
	EntityID    string
	Payload     map[string]any
	CreatedAt   time.Time
}

type EventQuery struct {
	CompanyID       string
	EntityType      string
	EntityID        string
	Limit           int
	EventType       string
	From            *time.Time
	To              *time.Time
	BeforeCreatedAt *time.Time
	BeforeEventID   string
}

type Service interface {
	LogEvent(event Event)
	ListEvents(query EventQuery) []Event
	CountEvents(query EventQuery) (int64, bool)
}
