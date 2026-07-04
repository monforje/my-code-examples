package natsrepo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"auth/internal/events"
	natsrepo "auth/internal/repository/nats"

	"github.com/nats-io/nats.go"
)

func TestProducer_IntegrationPublishIdentityCreated(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL,
		nats.Timeout(5*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
	)
	if err != nil {
		t.Skipf("NATS not available at %s: %v", natsURL, err)
	}
	t.Cleanup(func() { nc.Drain() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	producer := natsrepo.NewProducer(nc)
	err = producer.PublishIdentityCreated(ctx, events.IdentityCreatedPayload{
		IdentityID: "integration-identity-1",
		Email:      "integration@example.com",
	})
	if err != nil {
		t.Fatalf("PublishIdentityCreated() error = %v", err)
	}
}
