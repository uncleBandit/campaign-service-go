// internal/model/campaign.go
package model

import "time"

type Campaign struct {
    ID           int       `db:"id" json:"id"`
    Name         string    `db:"name" json:"name"`
    Channel      string    `db:"channel" json:"channel"`
    Status       string    `db:"status" json:"status"`
    BaseTemplate string    `db:"base_template" json:"base_template"`
    ScheduledAt  *time.Time `db:"scheduled_at" json:"scheduled_at,omitempty"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`

}
