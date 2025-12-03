package queue

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/unclebandit/smsleopard-backend/internal/repository"
)

// Queue interface
type Queue interface {
	Publish(topic string, payload any) error
	Subscribe(topic string, handler func(payload any) error) error
}

// InMemoryQueue is a production-ready in-memory queue with retry
type InMemoryQueue struct {
	mu       sync.Mutex
	handlers map[string][]func(payload any) error
}

// NewInMemoryQueue creates a new queue
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		handlers: make(map[string][]func(payload any) error),
	}
}

// JobPayload wraps a message payload with retry info
type JobPayload struct {
	Payload    any
	RetryCount int
	MaxRetries int
}

// Publish sends a message to all subscribers
func (q *InMemoryQueue) Publish(topic string, payload any) error {
	q.mu.Lock()
	handlers := q.handlers[topic]
	q.mu.Unlock()

	if len(handlers) == 0 {
		return fmt.Errorf("no subscribers for topic %s", topic)
	}

	job := JobPayload{
		Payload:    payload,
		RetryCount: 0,
		MaxRetries: 3,
	}

	for _, handler := range handlers {
		go q.processJob(handler, job)
	}

	return nil
}

// processJob handles retries and errors
func (q *InMemoryQueue) processJob(handler func(payload any) error, job JobPayload) {
	for job.RetryCount <= job.MaxRetries {
		err := handler(job.Payload)
		if err == nil {
			log.Printf("Job processed successfully: %+v\n", job.Payload)
			return // ACK
		}

		job.RetryCount++
		log.Printf("Job failed (attempt %d/%d): %+v, error: %v\n", job.RetryCount, job.MaxRetries, job.Payload, err)

		if job.RetryCount > job.MaxRetries {
			log.Printf("Job permanently failed after %d attempts: %+v\n", job.MaxRetries, job.Payload)
			return // No requeue
		}

		// Exponential backoff before retry
		time.Sleep(time.Duration(job.RetryCount*500) * time.Millisecond)
	}
}

// Subscribe adds a handler for a topic
func (q *InMemoryQueue) Subscribe(topic string, handler func(payload any) error) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.handlers[topic] = append(q.handlers[topic], handler)
	return nil
}

func StartCampaignSendSubscriber(q Queue, campaignRepo repository.CampaignRepositoryInterface) {
    go func() {
        err := q.Subscribe("campaign_sends", func(payload any) error {
            // Type assertion: payload should be an int (OutboundMessage ID)
            msgID, ok := payload.(int)
            if !ok {
                log.Println("⚠️ Invalid payload type, expected int")
                return nil // or return error to trigger retry
            }

            log.Println("📩 Processing queued outbound message ID:", msgID)

            // Fetch message details from DB
            msg, err := campaignRepo.GetOutboundMessageByID(msgID)
            if err != nil {
                log.Println("⚠️ Failed to fetch message:", err)
                return err
            }
            if msg == nil {
                log.Println("⚠️ Message not found for ID:", msgID)
                return nil // no retry
            }

            // TODO: Replace MockSender with actual SMS/email sending logic
            err = MockSender(msg.RenderedContent)
            if err != nil {
                log.Println("⚠️ Failed to send message:", err)
                // Update message status to "failed"
                _ = campaignRepo.UpdateOutboundMessageStatus(msgID, "failed", err.Error())
                return err // triggers retry in queue
            }

            // Update message status to "sent"
            err = campaignRepo.UpdateOutboundMessageStatus(msgID, "sent", "")
            if err != nil {
                log.Println("⚠️ Failed to update message status:", err)
                return err // retry
            }

            log.Println("✅ Message processed successfully:", msgID)
            return nil
        })

        if err != nil {
            log.Println("⚠️ Failed to start subscriber for campaign_sends:", err)
        }
    }()
}

//////////////////////////
// Example Mock Sender  //
//////////////////////////

// MockSender simulates sending messages with 90% success
func MockSender(payload any) error {
	r := rand.Float64()
	if r < 0.9 {
		return nil // success
	}
	return fmt.Errorf("mock sending failed")
}
