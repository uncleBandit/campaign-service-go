// internal/model/outbound_message.go
package model

import "time"

type OutboundMessage struct {
    ID              int       `db:"id" json:"id"`
    CampaignID      int       `db:"campaign_id" json:"campaign_id"`
    CustomerID      int       `db:"customer_id" json:"customer_id"`
    Status          string    `db:"status" json:"status"` // pending, sent, failed
    RenderedContent string    `db:"rendered_content" json:"rendered_content"`
    LastError       string    `db:"last_error,omitempty" json:"last_error,omitempty"`
    RetryCount      int       `db:"retry_count" json:"retry_count"`
    CreatedAt       time.Time `db:"created_at" json:"created_at"`
    UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
