package e2e_test_helpers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"auth/internal/events"

	"github.com/nats-io/nats.go"
)

type EventCapture struct {
	nc     *nats.Conn
	sub    *nats.Subscription
	mu     sync.Mutex
	events []events.Event
	done   chan struct{}
}

func NewEventCapture(nc *nats.Conn) *EventCapture {
	ec := &EventCapture{
		nc:   nc,
		done: make(chan struct{}),
	}

	var err error
	ec.sub, err = nc.Subscribe(">", func(msg *nats.Msg) {
		ec.appendMsg(msg)
	})
	if err != nil {
		panic("nats subscribe: " + err.Error())
	}
	_ = nc.Flush()

	return ec
}

func (ec *EventCapture) appendMsg(msg *nats.Msg) {
	var event events.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = append(ec.events, event)
}

func (ec *EventCapture) Wait(ctx context.Context, eventType events.EventType, matcher func(events.Event) bool) error {
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return &EventNotPublishedError{EventType: eventType}
		default:
		}

		ec.mu.Lock()
		for _, e := range ec.events {
			if e.Type == eventType {
				if matcher == nil || matcher(e) {
					ec.mu.Unlock()
					return nil
				}
			}
		}
		ec.mu.Unlock()

		time.Sleep(10 * time.Millisecond)
	}
}

func (ec *EventCapture) AssertPublished(t *testing.T, eventType events.EventType) {
	t.Helper()
	if err := ec.Wait(context.Background(), eventType, nil); err != nil {
		t.Fatalf("event %q was not published: %v", eventType, err)
	}
}

func (ec *EventCapture) Close() {
	if ec.sub != nil {
		_ = ec.sub.Unsubscribe()
	}
}

type EventNotPublishedError struct {
	EventType events.EventType
}

func (e *EventNotPublishedError) Error() string {
	return "event not published: " + string(e.EventType)
}
