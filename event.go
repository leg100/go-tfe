package tfe

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

const (
	EventOrganizationCreated   EventType = "organization_created"
	EventOrganizationDeleted   EventType = "organization_deleted"
	EventWorkspaceCreated      EventType = "workspace_created"
	EventWorkspaceDeleted      EventType = "workspace_deleted"
	EventRunCreated            EventType = "run_created"
	EventRunCompleted          EventType = "run_completed"
	EventRunCanceled           EventType = "run_canceled"
	EventRunApplied            EventType = "run_applied"
	EventRunPlanned            EventType = "run_planned"
	EventRunPlannedAndFinished EventType = "run_planned_and_finished"
	EventPlanQueued            EventType = "plan_queued"
	EventApplyQueued           EventType = "apply_queued"
	EventError                 EventType = "error"
)

type EventType string

type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// Events provides methods for sending and receiving events in real-time.
type Events interface {
	Subscribe(id string) (Subscription, error)
}

// Subscription represents a stream of events for a subscriber
type Subscription interface {
	// Event stream for all subscriber's event.
	C() <-chan Event

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

// events implements Events.
type events struct {
	client *Client
}

type subscription struct {
	conn *websocket.Conn
	ch   chan Event
}

func (e *events) Subscribe(id string) (Subscription, error) {
	u := url.URL{Scheme: "wss", Host: e.client.baseURL.Host, Path: "/events"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	ch := make(chan Event)

	go func() {
		defer c.Close()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				ch <- Event{Type: EventError, Payload: fmt.Sprintf("websocket read error: %s\n", err.Error())}
				return
			}

			var ev Event
			if err := json.Unmarshal(msg, &ev); err != nil {
				ch <- Event{Type: EventError, Payload: fmt.Sprintf("websocket decode error: %s\n", err.Error())}
				return
			}

			ch <- ev
		}
	}()

	return &subscription{conn: c, ch: ch}, nil
}

func (s *subscription) C() <-chan Event {
	return s.ch
}

func (s *subscription) Close() error {
	// Cleanly close the connection by sending a close message and then waiting
	// (with timeout) for the server to close the connection.
	err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return nil
}
