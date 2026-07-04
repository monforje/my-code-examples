package kafkarepo_test

import (
	kafkarepo "auth/internal/repository/kafka"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func TestProducer_IntegrationPublishUserCreated(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		t.Fatalf("kgo.NewClient() error = %v", err)
	}
	t.Cleanup(client.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	producer := kafkarepo.NewProducer(client)
	err = producer.PublishIdentityCreated(ctx, kafkarepo.IdentityCreatedPayload{
		IdentityID: "integration-identity-1",
		Email:      "integration@example.com",
	})
	if err != nil {
		t.Fatalf("PublishIdentityCreated() error = %v", err)
	}
}
