package service

import (
	"log"
	"github.com/unclebandit/smsleopard-backend/internal/model"
)

// OutboundRepo defines the methods the worker needs
type OutboundRepository interface {
	GetByID(id int) (*model.OutboundMessage, error)
	Update(msg *model.OutboundMessage) error
}

// Worker processes outbound message jobs
type Worker struct {
	OutboundRepo OutboundRepository
	JobChan      <-chan int
	SendFunc     func(msg string) bool
}

// Constructor
func NewWorker(repo OutboundRepository, jobChan <-chan int, sendFunc func(msg string) bool) *Worker {
	return &Worker{
		OutboundRepo: repo,
		JobChan:      jobChan,
		SendFunc:     sendFunc,
	}
}

// Start begins processing jobs
func (w *Worker) Start() {
	for jobID := range w.JobChan {
		msg, err := w.OutboundRepo.GetByID(jobID)
		if err != nil {
			log.Println("Failed to get message:", err)
			continue
		}

		// simulate sending
		success := w.SendFunc("mock message")
		if success {
			msg.Status = "sent"
		} else {
			msg.Status = "failed"
		}

		w.OutboundRepo.Update(msg)
	}
}
