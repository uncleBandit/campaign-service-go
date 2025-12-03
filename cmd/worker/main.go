package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/repository"
	"github.com/unclebandit/smsleopard-backend/internal/service"
)

type QueueJob struct {
    OutboundMessageID int `json:"outbound_message_id"`
}

func main() {
    // Connect to DB
    db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/smsleopard?sslmode=disable")
    if err != nil {
        log.Fatal("failed to connect to DB:", err)
    }

    // Repositories
    customerRepo := &repository.CustomerRepository{DB: db}
    campaignRepo := &repository.CampaignRepository{DB: db}
    outboundRepo := &repository.OutboundMessageRepository{DB: db}

    campaignService := &service.CampaignService{
        CampaignRepo:  campaignRepo,
        CustomerRepo:  customerRepo,
        OutboundRepo:  outboundRepo,
    }

    // Connect to RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal("Failed to connect to RabbitMQ:", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatal("Failed to open a channel:", err)
    }
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "campaign_sends", // name
        true,             // durable
        false,            // delete when unused
        false,            // exclusive
        false,            // no-wait
        nil,              // arguments
    )
    if err != nil {
        log.Fatal("Failed to declare queue:", err)
    }

    msgs, err := ch.Consume(
        q.Name,
        "",
        false, // autoAck = false for reliability
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatal("Failed to register consumer:", err)
    }

    forever := make(chan bool)

    go func() {
        for d := range msgs {
            var job QueueJob
            if err := json.Unmarshal(d.Body, &job); err != nil {
                log.Println("Invalid job:", err)
                d.Ack(false)
                continue
            }

            // Process the message
            err := processMessage(job.OutboundMessageID, campaignService)
            if err != nil {
                log.Println("Failed to send message:", err)
                // Retry logic: requeue up to 3 times
                var retryCount int
                if d.Headers["x-retry-count"] != nil {
                    retryCount = d.Headers["x-retry-count"].(int)
                }
                if retryCount < 3 {
                    d.Nack(false, true) // requeue
                    continue
                }
            }

            d.Ack(false)
        }
    }()

    log.Println("Worker running, waiting for messages...")
    <-forever
}

func processMessage(outboundID int, svc *service.CampaignService) error {
    // Fetch outbound message + customer + campaign
    msg, err := svc.OutboundRepo.GetByID(outboundID)
    if err != nil {
        return err
    }

    customer, err := svc.CustomerRepo.GetByID(msg.CustomerID)
    if err != nil {
        return err
    }

    campaign, err := svc.CampaignRepo.GetByID(msg.CampaignID)
    if err != nil {
        return err
    }

    // Render message
    rendered := renderTemplate(campaign.BaseTemplate, customer)

    // Mock sending
    success := mockSend(rendered)
    if success {
        msg.Status = "sent"
        msg.RenderedContent = rendered
        msg.LastError = ""
    } else {
        msg.Status = "failed"
        msg.LastError = "mock send failed"
        msg.RetryCount += 1
    }

    return svc.OutboundRepo.Update(msg)
}

func renderTemplate(template string, customer *model.Customer) string {
    result := template
    placeholders := map[string]string{
        "first_name":       customer.FirstName,
        "last_name":        customer.LastName,
        "location":         customer.Location,
        "preferred_product": customer.PreferredProduct,
    }
    for k, v := range placeholders {
        result = replacePlaceholder(result, k, v)
    }
    return result
}

func replacePlaceholder(template, key, value string) string {
    placeholder := "{" + key + "}"
    if value == "" {
        value = "N/A"
    }
    return stringReplace(template, placeholder, value)
}

// Simple string replace helper
func stringReplace(s, old, new string) string {
    return fmt.Sprintf("%s", strings.ReplaceAll(s, old, new))
}

// Mock sender: 90% chance of success
func mockSend(msg string) bool {
    rand.Seed(time.Now().UnixNano())
    return rand.Intn(100) < 90
}
