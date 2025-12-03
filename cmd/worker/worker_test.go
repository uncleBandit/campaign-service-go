package main

import (
	"sync"
	"testing"

	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/service"
)

// MockOutboundRepo stores messages in memory
type MockOutboundRepo struct {
	msgs map[int]*model.OutboundMessage
	mu   sync.Mutex
}

func (m *MockOutboundRepo) GetByID(id int) (*model.OutboundMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.msgs[id], nil
}

func (m *MockOutboundRepo) Update(msg *model.OutboundMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msgs[msg.ID] = msg
	return nil
}

// Mock sender function always succeeds
func MockSender(msg string) bool {
	return true
}

func TestWorker(t *testing.T) {
	repo := &MockOutboundRepo{
		msgs: map[int]*model.OutboundMessage{
			1: {ID: 1, Status: "pending", CampaignID: 1, CustomerID: 1},
		},
	}

	jobChan := make(chan int, 1)
	jobChan <- 1 // enqueue job

	var wg sync.WaitGroup
	wg.Add(1)

	worker := service.NewWorker(repo, jobChan, func(msg string) bool {
		success := MockSender(msg)
		wg.Done() // signal that job is processed
		return success
	})

	// Start worker
	go worker.Start()

	// Wait until worker processes the job
	wg.Wait()

	// Verify status
	msg, _ := repo.GetByID(1)
	if msg.Status != "sent" {
		t.Errorf("expected sent, got %s", msg.Status)
	}
}
