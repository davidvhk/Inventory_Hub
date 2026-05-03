package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type InventoryEvent struct {
	InventoryTimestamp string `json:"inventory_timestamp"`
	Hostname           string `json:"hostname"`
	IPAddress          string `json:"ip_address"`
	OS                 string `json:"os"`
	UUID               string `json:"uuid"`
	DeviceID           string `json:"device_id"`
}

func main() {
	// Config from Env
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "inventory_events"
	}
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "inv-consumer-group"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://inv_user:inv_password@localhost:5432/inventory_db?sslmode=disable"
	}

	// Init DB
	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", dbURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to DB, retrying in 5s... (%d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer db.Close()

	// Init Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaBrokers},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("Consumer started. Topic: %s, Brokers: %s, GroupID: %s", topic, kafkaBrokers, groupID)

	// Graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigchan
		log.Println("Shutting down...")
		cancel()
	}()

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event InventoryEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		log.Printf("Received event for device: %s", event.DeviceID)

		dbTime := parseInventoryTime(event.InventoryTimestamp)

		// Upsert into Postgres
		query := `
			INSERT INTO devices (inventory_timestamp, hostname, ip_address, os, uuid, device_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (device_id) DO UPDATE SET
				inventory_timestamp = EXCLUDED.inventory_timestamp,
				hostname = EXCLUDED.hostname,
				ip_address = EXCLUDED.ip_address,
				os = EXCLUDED.os,
				uuid = EXCLUDED.uuid,
				updated_at = NOW()
		`

		_, err = db.Exec(query,
			dbTime,
			event.Hostname,
			event.IPAddress,
			event.OS,
			event.UUID,
			event.DeviceID,
		)

		if err != nil {
			log.Printf("Error inserting into DB: %v", err)
		} else {
			log.Printf("Successfully saved device %s to DB", event.DeviceID)
		}
	}

	log.Println("Consumer stopped.")
}

func parseInventoryTime(ts string) time.Time {
	// Parse the timestamp. OCS often uses "Sun May  3 10:34" or similar.
	parsedTime, err := time.Parse("Mon Jan _2 15:04", ts)
	if err != nil {
		log.Printf("Warning: Could not parse timestamp '%s', using NOW(): %v", ts, err)
		return time.Now().Truncate(time.Second)
	}
	// Add current year since OCS format doesn't provide it
	return parsedTime.AddDate(time.Now().Year(), 0, 0).Truncate(time.Second)
}
