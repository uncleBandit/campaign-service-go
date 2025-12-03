// internal/errors/errors.go
package appErrors

import "fmt"

// ErrCampaignNotFound is a sentinel error
type ErrCampaignNotFound struct {
    CampaignID int
}

func (e *ErrCampaignNotFound) Error() string {
    return fmt.Sprintf("campaign with ID %d not found", e.CampaignID)
}

// Helper constructor
func NewCampaignNotFound(id int) error {
    return &ErrCampaignNotFound{CampaignID: id}
}
