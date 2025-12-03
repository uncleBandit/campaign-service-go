package service_test

import (
	"testing"
	"time"

	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/service"
)

// ✅ Mock Campaign Repository for pagination
type MockCampaignPaginationRepo struct{}

func (m *MockCampaignPaginationRepo) ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error) {
	all := []*model.Campaign{
		{ID: 5, Name: "C5"},
		{ID: 4, Name: "C4"},
		{ID: 3, Name: "C3"},
		{ID: 2, Name: "C2"},
		{ID: 1, Name: "C1"},
	}

	start := offset
	end := offset + limit

	if start >= len(all) {
		return []*model.Campaign{}, len(all), nil
	}
	if end > len(all) {
		end = len(all)
	}

	return all[start:end], len(all), nil
}

// Stub implementations to satisfy the interface
func (m *MockCampaignPaginationRepo) Create(c *model.Campaign) error {
	c.ID = 999 // fake ID
	c.CreatedAt = time.Now()
	return nil
}

func (m *MockCampaignPaginationRepo) GetByID(id int) (*model.Campaign, error) {
	return &model.Campaign{ID: id, Name: "Mock"}, nil
}

func (m *MockCampaignPaginationRepo) Update(c *model.Campaign) error {
	return nil
}

func (m *MockCampaignPaginationRepo) UpdateStatus(id int, status string) error {
	// do nothing, just stub
	return nil
}

func TestPagination(t *testing.T) {
    svc := &service.CampaignService{
        CampaignRepo: &MockCampaignPaginationRepo{},
    }

    pageSize := 2

    // Call ListCampaigns with 4 arguments (page, pageSize, channel, status)
    page1, pagination1, _ := svc.ListCampaigns(1, pageSize, "", "")
    page2, _, _ := svc.ListCampaigns(2, pageSize, "", "")

    expectedTotal := 5
    if pagination1["total_count"] != expectedTotal {
        t.Errorf("expected total_count %d, got %d", expectedTotal, pagination1["total_count"])
    }

    if len(page1) != 2 || len(page2) != 2 {
        t.Fatalf("expected full pages, got %d and %d", len(page1), len(page2))
    }

    // Check descending order
    if page1[0].ID <= page1[1].ID {
        t.Errorf("expected descending order in page 1")
    }
    if page2[0].ID <= page2[1].ID {
        t.Errorf("expected descending order in page 2")
    }

    // Check no duplicates between pages
    if page1[1].ID == page2[0].ID {
        t.Errorf("duplicate entry between pages: %v", page1[1].ID)
    }

    // Optional: check last page
    page3, pagination3, _ := svc.ListCampaigns(3, pageSize, "", "")
    if len(page3) != 1 {
        t.Errorf("expected last page to have 1 item, got %d", len(page3))
    }

    if pagination3["total_count"] != expectedTotal {
        t.Errorf("expected total_count %d, got %d", expectedTotal, pagination3["total_count"])
    }
}

func (m *MockCampaignPaginationRepo) CreateOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID:         1,
        CampaignID: campaignID,
        CustomerID: customerID,
        Status:     "pending",
        RetryCount: 0,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }, nil
}

func (m *MockCampaignPaginationRepo) GetOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID:         1,
        CampaignID: campaignID,
        CustomerID: customerID,
        Status:     "pending",
        RetryCount: 0,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }, nil
}

func (m *MockCampaignPaginationRepo) UpdateOutboundMessageStatus(id int, status, lastError string) error {
    return nil
}

func (m *MockCampaignPaginationRepo) GetCampaignStats(campaignID int) (map[string]int, error) {
    return map[string]int{"pending": 1, "sent": 0, "failed": 0}, nil
}


func (m *MockCampaignRepo) GetOutboundMessageByID(id int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID: id,
        RenderedContent: "Test message",
    }, nil
}

func (m *MockCampaignPaginationRepo) GetOutboundMessageByID(id int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID: id,
        RenderedContent: "Test message",
    }, nil
}

func (m *MockCampaignPaginationRepo) UpdateOutboundMessageContent(id int, content string) error {
    return nil
}
