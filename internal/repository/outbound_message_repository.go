package repository

import (
	"database/sql"
	"time"
    

	"github.com/unclebandit/smsleopard-backend/internal/db"
	"github.com/unclebandit/smsleopard-backend/internal/model"
)

type OutboundMessageRepository struct {
    DB *sql.DB
}

// Create inserts a new outbound message into the database and returns the created ID
func (r *OutboundMessageRepository) Create(msg *model.OutboundMessage) error {
    now := time.Now()
    msg.CreatedAt = now
    msg.UpdatedAt = now

    query := `
        INSERT INTO outbound_messages 
        (campaign_id, customer_id, status, rendered_content, last_error, retry_count, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `
    return r.DB.QueryRow(
        query,
        msg.CampaignID,
        msg.CustomerID,
        msg.Status,
        msg.RenderedContent,
        msg.LastError,
        msg.RetryCount,
        msg.CreatedAt,
        msg.UpdatedAt,
    ).Scan(&msg.ID)
}

// Update updates an existing outbound message (e.g., status, last_error, retry_count)
func (r *OutboundMessageRepository) Update(msg *model.OutboundMessage) error {
    msg.UpdatedAt = time.Now()
    query := `
        UPDATE outbound_messages
        SET status=$1, last_error=$2, retry_count=$3, updated_at=$4
        WHERE id=$5
    `
    _, err := r.DB.Exec(query, msg.Status, msg.LastError, msg.RetryCount, msg.UpdatedAt, msg.ID)
    return err
}

// GetByID fetches an outbound message by its ID
func (r *OutboundMessageRepository) GetByID(id int) (*model.OutboundMessage, error) {
    query := `
        SELECT id, campaign_id, customer_id, status, rendered_content, last_error, retry_count, created_at, updated_at
        FROM outbound_messages
        WHERE id=$1
    `
    var msg model.OutboundMessage
    err := r.DB.QueryRow(query, id).Scan(
        &msg.ID,
        &msg.CampaignID,
        &msg.CustomerID,
        &msg.Status,
        &msg.RenderedContent,
        &msg.LastError,
        &msg.RetryCount,
        &msg.CreatedAt,
        &msg.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &msg, nil
}


func OutboundMessageExists(campaignID, customerID int) (bool, error) {
    query := `
        SELECT 1 FROM outbound_messages
        WHERE campaign_id = $1 AND customer_id = $2
        LIMIT 1
    `
    row := db.DB.QueryRow(query, campaignID, customerID)
    var tmp int
    err := row.Scan(&tmp)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, nil
        }
        return false, err
    }
    return true, nil
}


// Exists checks if an outbound message for the campaign & customer already exists
// Exists checks if an outbound message for the campaign & customer already exists

