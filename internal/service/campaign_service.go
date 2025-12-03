// internal/service/campaign_service.go
package service

import (
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/unclebandit/smsleopard-backend/internal/model"
    "github.com/unclebandit/smsleopard-backend/internal/repository"
    "github.com/unclebandit/smsleopard-backend/internal/queue"
)

type CampaignService struct {
    CampaignRepo repository.CampaignRepositoryInterface
    CustomerRepo repository.CustomerRepositoryInterface
    OutboundRepo *repository.OutboundMessageRepository
    Queue        queue.Queue
}

// Result struct for SendCampaign
type SendCampaignResult struct {
    CampaignID     int
    MessagesQueued int
    Status         string
    MessageIDs     []int
}

type CampaignDetails struct {
    ID           int               `json:"id"`
    Name         string            `json:"name"`
    Channel      string            `json:"channel"`
    Status       string            `json:"status"`
    BaseTemplate string            `json:"base_template"`
    ScheduledAt  *time.Time        `json:"scheduled_at,omitempty"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    *time.Time         `json:"updated_at"`
    Stats        map[string]int    `json:"stats"`
}




func (s *CampaignService) RenderPreview(campaignID, customerID int, overrideTemplate *string) (string, error) {

    campaign, err := s.CampaignRepo.GetByID(campaignID)
    if err != nil {
        return "", err
    }
    if campaign == nil {
        return "", fmt.Errorf("campaign not found")
    }

    customer, err := s.CustomerRepo.GetByID(customerID)
    if err != nil {
        return "", err
    }
    if customer == nil {
        return "", fmt.Errorf("customer not found")
    }

    template := campaign.BaseTemplate

    if overrideTemplate != nil && strings.TrimSpace(*overrideTemplate) != "" {
        template = *overrideTemplate
    }

    if strings.TrimSpace(template) == "" {
        return "", fmt.Errorf("template cannot be empty")
    }

    message := template
    message = strings.ReplaceAll(message, "{first_name}", customer.FirstName)
    message = strings.ReplaceAll(message, "{last_name}", customer.LastName)
    message = strings.ReplaceAll(message, "{location}", customer.Location)
    message = strings.ReplaceAll(message, "{preferred_product}", customer.PreferredProduct)

    return message, nil
}


func replace(template, placeholder, value string) string {
    if value == "" {
        value = "<unknown>"
    }
    return strings.ReplaceAll(template, placeholder, value)
}


func (s *CampaignService) SendCampaign(campaignID int, customerIDs []int) (*SendCampaignResult, error) {
    campaign, err := s.CampaignRepo.GetByID(campaignID)
    if err != nil {
        return nil, err
    }

    if campaign.Status != "draft" && campaign.Status != "scheduled" && campaign.Status != "sending" {
        return nil, fmt.Errorf("campaign cannot be sent in status: %s", campaign.Status)
    }

    result := &SendCampaignResult{
        CampaignID:     campaignID,
        MessagesQueued: 0,
        Status:         "sending",
        MessageIDs:     []int{},
    }

    for _, customerID := range customerIDs {
        // Idempotent create (returns existing if already exists)
        msg, err := s.CampaignRepo.CreateOutboundMessage(campaignID, customerID)
        if err != nil {
            log.Println("⚠️ failed to create/get outbound message:", err)
            continue
        }
        if msg == nil {
            log.Println("⚠️ message already exists but could not be fetched")
            continue
        }

        // Render content if empty
        if msg.RenderedContent == "" {
            rendered, err := s.RenderPreview(campaignID, customerID, nil)
            if err != nil {
                log.Println("⚠️ failed to render message for customer", customerID, ":", err)
                continue
            }

            if err := s.CampaignRepo.UpdateOutboundMessageContent(msg.ID, rendered); err != nil {
                log.Println("⚠️ failed to update rendered content:", err)
                continue
            }
            msg.RenderedContent = rendered
        }


        // Always queue the message
        if err := s.Queue.Publish("campaign_sends", msg.ID); err != nil {
            log.Println("⚠️ failed to enqueue message ID", msg.ID, ":", err)
            continue
        }

        result.MessageIDs = append(result.MessageIDs, msg.ID)
        result.MessagesQueued++
    }


    if campaign.Status != "sending" {
        if err := s.CampaignRepo.UpdateStatus(campaignID, "sending"); err != nil {
            return result, err
        }
    }

    return result, nil
}




func (s *CampaignService) CreateCampaign(name, channel, baseTemplate string, scheduledAt *string) (*model.Campaign, error) {
    c := &model.Campaign{
        Name:         name,
        Channel:      channel,
        BaseTemplate: baseTemplate,
        Status:       "draft",
    }

    if scheduledAt != nil {
        // parse scheduledAt string into time.Time
        t, err := time.Parse(time.RFC3339, *scheduledAt)
        if err != nil {
            return nil, err
        }
        c.ScheduledAt = &t
    }

    if err := s.CampaignRepo.Create(c); err != nil {
        return nil, err
    }

    return c, nil
}

// ListCampaigns fetches campaigns with pagination
func (s *CampaignService) ListCampaigns(page, pageSize int, channel, status string) ([]model.Campaign, map[string]int, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 100 {
        pageSize = 100
    }
    offset := (page - 1) * pageSize

    ptrs, total, err := s.CampaignRepo.ListCampaigns(offset, pageSize, channel, status)
    if err != nil {
        return nil, nil, err
    }

    campaigns := make([]model.Campaign, len(ptrs))
    for i, c := range ptrs {
        campaigns[i] = *c
    }

    totalPages := (total + pageSize - 1) / pageSize
    pagination := map[string]int{
        "page":        page,
        "page_size":   pageSize,
        "total_count": total,
        "total_pages": totalPages,
    }

    return campaigns, pagination, nil
}

// GetCampaignDetails fetches a campaign by ID
func (s *CampaignService) GetCampaignDetails(id int) (*model.Campaign, error) {
    return s.CampaignRepo.GetByID(id)
}

func (s *CampaignService) GetCampaignDetailsWithStats(campaignID int) (*CampaignDetails, error) {
    log.Println("Fetching campaign with ID:", campaignID)

    // Fetch the campaign
    campaign, err := s.CampaignRepo.GetByID(campaignID)
    if err != nil {
        log.Println("Failed to fetch campaign:", err)
        return nil, err
    }
    log.Printf("Campaign fetched: %+v\n", campaign)

    // Fetch outbound message counts by status
    query := `
        SELECT status, COUNT(*) 
        FROM outbound_messages
        WHERE campaign_id = $1
        GROUP BY status
    `
    rows, err := s.OutboundRepo.DB.Query(query, campaignID)
    if err != nil {
        log.Println("Failed to query outbound messages:", err)
        return nil, err
    }
    defer rows.Close()

    // initialize stats map
    stats := map[string]int{
        "total":   0,
        "pending": 0,
        "sending": 0,
        "sent":    0,
        "failed":  0,
    }

    for rows.Next() {
        var status string
        var count int
        if err := rows.Scan(&status, &count); err != nil {
            log.Println("Failed to scan row:", err)
            return nil, err
        }
        log.Printf("Status row: %s => %d\n", status, count)

        if _, ok := stats[status]; ok {
            stats[status] = count
        }
        stats["total"] += count
    }

    log.Printf("Final stats map: %+v\n", stats)

    return &CampaignDetails{
        ID:           campaign.ID,
        Name:         campaign.Name,
        Channel:      campaign.Channel,
        Status:       campaign.Status,
        BaseTemplate: campaign.BaseTemplate,
        ScheduledAt:  campaign.ScheduledAt,
        CreatedAt:    campaign.CreatedAt,
        UpdatedAt:    campaign.UpdatedAt,
        Stats:        stats,
    }, nil
}





