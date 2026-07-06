package core

import "time"

type Event struct {
	Timestamp time.Time
	Source    string
	Service   string
	EventType string
	Level     string
	IP        string
	Message   string
	Metadata  map[string]string
}

type Trigger struct {
	Event  Event
	Reason string
	Count  int
}
