// internal/controller/campaign_controller.go
package controller

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    "github.com/unclebandit/smsleopard-backend/internal/service"

    "github.com/go-chi/chi/v5"
    "github.com/streadway/amqp"
)


type CampaignController struct {
    CampaignService *service.CampaignService
}

func (c *CampaignController) PersonalizedPreview(w http.ResponseWriter, r *http.Request) {
    campaignIDStr := chi.URLParam(r, "id")
    campaignID, _ := strconv.Atoi(campaignIDStr)

    var body struct {
        CustomerID       int     `json:"customer_id"`
        OverrideTemplate *string `json:"override_template"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    rendered, err := c.CampaignService.RenderPreview(campaignID, body.CustomerID, body.OverrideTemplate)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "rendered_message": rendered,
        "used_template":    body.OverrideTemplate,
        "customer_id":      body.CustomerID,
    })
}

func (c *CampaignController) CreateCampaign(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Name         string  `json:"name"`
        Channel      string  `json:"channel"`
        BaseTemplate string  `json:"base_template"`
        ScheduledAt  *string `json:"scheduled_at"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    campaign, err := c.CampaignService.CreateCampaign(body.Name, body.Channel, body.BaseTemplate, body.ScheduledAt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(campaign)
}


func (c *CampaignController) ListCampaigns(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    channel := r.URL.Query().Get("channel")
    status := r.URL.Query().Get("status")

    // Default values if missing
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }

    // Fetch campaigns and pagination info from service
    campaigns, pagination, err := c.CampaignService.ListCampaigns(page, pageSize, channel, status)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return JSON response
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data":       campaigns,
        "pagination": pagination, // already contains total_count, total_pages, page, page_size
    })
}




func (c *CampaignController) GetCampaignDetails(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.Atoi(idStr)

    campaign, err := c.CampaignService.GetCampaignDetails(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(campaign)
}


func (c *CampaignController) SendCampaign(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.Atoi(idStr)

    var body struct {
        CustomerIDs []int `json:"customer_ids"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    // Send campaign via service
    result, err := c.CampaignService.SendCampaign(id, body.CustomerIDs)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Connect to RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        http.Error(w, "Failed to connect to queue", http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        http.Error(w, "Failed to open queue channel", http.StatusInternalServerError)
        return
    }
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "campaign_sends",
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        http.Error(w, "Failed to declare queue", http.StatusInternalServerError)
        return
    }

    // Publish each message to the queue
    for _, msgID := range result.MessageIDs {
        body, _ := json.Marshal(map[string]int{"outbound_message_id": msgID})
        err = ch.Publish(
            "",
            q.Name,
            false,
            false,
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            },
        )
        if err != nil {
            log.Println("Failed to publish message:", err)
        }
    }

    // Return JSON response
    json.NewEncoder(w).Encode(map[string]interface{}{
        "campaign_id":     result.CampaignID,
        "messages_queued": result.MessagesQueued,
        "status":          result.Status,
    })
}




