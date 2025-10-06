package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/db"
	"github.com/censys/scan-takehome/pkg/scanning"
)

func main() {
	projectId := flag.String("project", "test-project", "GCP Project ID")
	subscriptionId := flag.String("subscription", "scan-sub", "GCP PubSub Subscription ID")
	dbUrl := flag.String(
		"db-url",
		"postgres://postgres:postgres@db:5432/scanning?sslmode=disable",
		"Database URL",
	)

	ctx := context.Background()

	log.Printf("connecting to database '%s'...\n", *dbUrl)
	db, err := db.NewDatabase(ctx, *dbUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected")

	if err := subscribeScans(ctx, db, *projectId, *subscriptionId); err != nil {
		log.Printf("failed to subscribe to read scans: %v", err)
		os.Exit(1)
	}
}

// subscribeScans subscribes to the pubsub topic and passes scan results to the persistence layer.
func subscribeScans(
	ctx context.Context,
	db db.Scanning,
	projectId string,
	subscriptionId string,
) error {
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		return fmt.Errorf("failed to connect to pubsub: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sub := client.Subscription(subscriptionId)
	if err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		if err := handleScan(ctx, db, m); err != nil {
			log.Printf("failed to handle scan: %v", err)
			m.Nack()
		} else {
			m.Ack()
		}
	}); err != nil {
		return fmt.Errorf("failed to receive from pubsub: %w", err)
	}

	return nil
}

// handleScan decodes a scan's results and passes them to the persistence layer.
func handleScan(ctx context.Context, db db.Scanning, m *pubsub.Message) error {
	var scan scanning.Scan
	if err := json.Unmarshal(m.Data, &scan); err != nil {
		return fmt.Errorf("failed to unmarshal scan: %w", err)
	}

	log.Printf("received scan: %+v\n", scan)

	switch scan.DataVersion {
	case scanning.V1:
		v1Data, ok := scan.Data.(string)
		if !ok {
			return fmt.Errorf("invalid scan V1Data format: %d", scan.DataVersion)
		}

		decoded, err := base64.StdEncoding.DecodeString(v1Data)
		if err != nil {
			return fmt.Errorf("failed to decode V1Data: %w", err)
		}
		scan.Data = string(decoded)
	case scanning.V2:
		v2Data, ok := scan.Data.(string)
		if !ok {
			return fmt.Errorf("invalid scan V2Data format: %d", scan.DataVersion)
		}

		scan.Data = v2Data
	default:
		return fmt.Errorf("unknown scan data version: %d", scan.DataVersion)
	}

	log.Printf("decoded scan: %+v\n", scan)
	if err := db.Upsert(ctx, scan); err != nil {
		return fmt.Errorf("failed to upsert scan: %w", err)
	}

	return nil
}
