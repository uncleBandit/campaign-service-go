package repository

import (
    "database/sql"
    "fmt"
    "time"
    "log"

    appErrors "github.com/unclebandit/smsleopard-backend/internal/errors"
    "github.com/unclebandit/smsleopard-backend/internal/model"
)

type CampaignRepositoryInterface interface {
    // Campaign CRUD
    ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error)
    GetByID(id int) (*model.Campaign, error)
    UpdateStatus(campaignID int, status string) error
    Update(c *model.Campaign) error
    Create(c *model.Campaign) error

    // Outbound messages
    CreateOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error)
    GetOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error)
    UpdateOutboundMessageStatus(id int, status, lastError string) error
    GetCampaignStats(campaignID int) (map[string]int, error)
    UpdateOutboundMessageContent(id int, content string) error
    GetOutboundMessageByID(id int) (*model.OutboundMessage, error)
}

type CampaignRepository struct {
    DB *sql.DB
}

// ====================== Campaign CRUD ======================

func (r *CampaignRepository) Create(c *model.Campaign) error {
    c.CreatedAt = time.Now()
    if c.Status == "" {
        c.Status = "draft"
    }
    query := `
        INSERT INTO campaigns (name, channel, status, base_template, scheduled_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
    return r.DB.QueryRow(query, c.Name, c.Channel, c.Status, c.BaseTemplate, c.ScheduledAt, c.CreatedAt).Scan(&c.ID)
}

func (r *CampaignRepository) Update(c *model.Campaign) error {
    query := `
        UPDATE campaigns
        SET name=$1, base_template=$2, status=$3, updated_at=NOW()
        WHERE id=$4
    `
    _, err := r.DB.Exec(query, c.Name, c.BaseTemplate, c.Status, c.ID)
    return err
}

func (r *CampaignRepository) UpdateStatus(campaignID int, status string) error {
    query := `UPDATE campaigns SET status=$1, updated_at=$2 WHERE id=$3`
    _, err := r.DB.Exec(query, status, time.Now(), campaignID)
    return err
}

func (r *CampaignRepository) GetByID(id int) (*model.Campaign, error) {
    query := `
        SELECT id, name, channel, status, base_template, scheduled_at, created_at, updated_at
        FROM campaigns WHERE id=$1
    `
    var c model.Campaign
    err := r.DB.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.Channel, &c.Status, &c.BaseTemplate, &c.ScheduledAt, &c.CreatedAt, &c.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, appErrors.NewCampaignNotFound(id)
        }
        return nil, err
    }
    return &c, nil
}

func (r *CampaignRepository) ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error) {
    campaigns := []*model.Campaign{}
    query := `SELECT id, name, channel, status, base_template, scheduled_at, created_at, updated_at FROM campaigns WHERE 1=1`
    args := []interface{}{}
    argPos := 1

    if channel != "" {
        query += fmt.Sprintf(" AND channel=$%d", argPos)
        args = append(args, channel)
        argPos++
    }
    if status != "" {
        query += fmt.Sprintf(" AND status=$%d", argPos)
        args = append(args, status)
        argPos++
    }

    query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
    args = append(args, limit, offset)

    rows, err := r.DB.Query(query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    for rows.Next() {
        c := &model.Campaign{}
        if err := rows.Scan(&c.ID, &c.Name, &c.Channel, &c.Status, &c.BaseTemplate, &c.ScheduledAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
            return nil, 0, err
        }
        campaigns = append(campaigns, c)
    }

    // Count total
    countQuery := `SELECT COUNT(*) FROM campaigns WHERE 1=1`
    argsCount := []interface{}{}
    argPosCount := 1
    if channel != "" {
        countQuery += fmt.Sprintf(" AND channel=$%d", argPosCount)
        argsCount = append(argsCount, channel)
        argPosCount++
    }
    if status != "" {
        countQuery += fmt.Sprintf(" AND status=$%d", argPosCount)
        argsCount = append(argsCount, status)
    }

    var total int
    if err := r.DB.QueryRow(countQuery, argsCount...).Scan(&total); err != nil {
        return nil, 0, err
    }

    return campaigns, total, nil
}

// ====================== Outbound Messages ======================

// Idempotent insert
func (r *CampaignRepository) CreateOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
    // 1. Check if message already exists
    existing, err := r.GetOutboundMessage(campaignID, customerID)
    if err != nil {
        return nil, err
    }
    if existing != nil {
        return existing, nil // return the existing one
    }

    // 2. Insert new message
    query := `
        INSERT INTO outbound_messages (campaign_id, customer_id, status, retry_count, created_at)
        VALUES ($1, $2, 'pending', 0, NOW())
        RETURNING id, status, retry_count, created_at, updated_at,rendered_content
    `
    var msg model.OutboundMessage
    err = r.DB.QueryRow(query, campaignID, customerID).Scan(&msg.ID, &msg.Status, &msg.RetryCount, &msg.CreatedAt, &msg.UpdatedAt)
    if err != nil {
        return nil, err
    }

    msg.CampaignID = campaignID
    msg.CustomerID = customerID
    return &msg, nil
}


func (r *CampaignRepository) GetOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
    query := `SELECT id, campaign_id, customer_id, status, rendered_content, last_error, retry_count, created_at, updated_at
              FROM outbound_messages
              WHERE campaign_id=$1 AND customer_id=$2`
    var msg model.OutboundMessage
    err := r.DB.QueryRow(query, campaignID, customerID).Scan(
        &msg.ID, &msg.CampaignID, &msg.CustomerID, &msg.Status,
        &msg.RenderedContent, &msg.LastError, &msg.RetryCount,
        &msg.CreatedAt, &msg.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &msg, nil
}

func (r *CampaignRepository) UpdateOutboundMessageStatus(id int, status, lastError string) error {
    query := `UPDATE outbound_messages SET status=$1, last_error=$2, retry_count=retry_count+1, updated_at=NOW() WHERE id=$3`
    _, err := r.DB.Exec(query, status, lastError, id)
    return err
}

func (r *CampaignRepository) GetCampaignStats(campaignID int) (map[string]int, error) {
    query := `SELECT status, COUNT(*) FROM outbound_messages WHERE campaign_id=$1 GROUP BY status`
    rows, err := r.DB.Query(query, campaignID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    stats := map[string]int{"pending": 0, "sent": 0, "failed": 0}
    for rows.Next() {
        var status string
        var count int
        if err := rows.Scan(&status, &count); err != nil {
            return nil, err
        }
        stats[status] = count
    }
    return stats, nil
}

func (r *CampaignRepository) Exists(campaignID, customerID int) (bool, error) {
    var count int
    err := r.DB.QueryRow(`
        SELECT COUNT(*) 
        FROM outbound_messages 
        WHERE campaign_id = $1 AND customer_id = $2`, campaignID, customerID).Scan(&count)
    if err != nil {
        log.Println("⚠️ Exists query error:", err)
        return false, err
    }
    return count > 0, nil
}

func (r *CampaignRepository) UpdateOutboundMessageContent(id int, content string) error {
    query := `UPDATE outbound_messages SET rendered_content=$1, updated_at=NOW() WHERE id=$2`
    _, err := r.DB.Exec(query, content, id)
    return err
}

func (r *CampaignRepository) GetOutboundMessageByID(id int) (*model.OutboundMessage, error) {
    query := `
        SELECT id, campaign_id, customer_id, status, rendered_content, last_error, retry_count, created_at, updated_at
        FROM outbound_messages
        WHERE id=$1
    `
    var msg model.OutboundMessage
    err := r.DB.QueryRow(query, id).Scan(
        &msg.ID, &msg.CampaignID, &msg.CustomerID, &msg.Status,
        &msg.RenderedContent, &msg.LastError, &msg.RetryCount,
        &msg.CreatedAt, &msg.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &msg, nil
}


var _ CampaignRepositoryInterface = (*CampaignRepository)(nil)
