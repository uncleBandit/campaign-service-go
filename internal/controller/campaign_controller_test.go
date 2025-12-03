package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/unclebandit/smsleopard-backend/internal/controller"
	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/service"
)

// --- Mock Repositories ---

type MockCustomerRepo struct{}
type MockCampaignRepoForPagination struct {
	campaigns []*model.Campaign
}


func (m *MockCustomerRepo) GetByID(id int) (*model.Customer, error) {
	return &model.Customer{
		FirstName:       "Alice",
		LastName:        "Smith",
		Location:        "Nairobi",
		PreferredProduct: "Shoes",
	}, nil
}

func (m *MockCustomerRepo) ListAll() ([]model.Customer, error) {
	return []model.Customer{}, nil
}

type MockCampaignRepo struct{}

func (m *MockCampaignRepo) GetByID(id int) (*model.Campaign, error) {
	return &model.Campaign{
		BaseTemplate: "Hi {first_name} {last_name}, check out {preferred_product} in {location}!",
	}, nil
}

func (m *MockCampaignRepo) Create(c *model.Campaign) error                    { return nil }
func (m *MockCampaignRepo) Update(c *model.Campaign) error                    { return nil }
func (m *MockCampaignRepo) ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error) {
	return []*model.Campaign{}, 0, nil
}
func (m *MockCampaignRepo) UpdateStatus(id int, status string) error { return nil }

// --- Test Function ---

func TestPersonalizedPreviewHandler(t *testing.T) {
	// Initialize service with mocks
	svc := &service.CampaignService{
		CampaignRepo: &MockCampaignRepo{},
		CustomerRepo: &MockCustomerRepo{},
	}

	ctrl := &controller.CampaignController{
		CampaignService: svc,
	}

	// Create request body
	body := map[string]interface{}{"customer_id": 1}
	b, _ := json.Marshal(body)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/campaigns/1/personalized-preview", bytes.NewReader(b))
	w := httptest.NewRecorder()

	// Call the controller handler
	ctrl.PersonalizedPreview(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	msg, ok := res["rendered_message"].(string)
	if !ok {
		t.Fatalf("rendered_message not found or not a string")
	}

	if !strings.Contains(msg, "Alice") {
		t.Errorf("expected 'Alice' in message, got %q", msg)
	}
}


func (m *MockCampaignRepoForPagination) ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error) {
	var filtered []*model.Campaign
	for _, c := range m.campaigns {
		if channel != "" && c.Channel != channel {
			continue
		}
		if status != "" && c.Status != status {
			continue
		}
		filtered = append(filtered, c)
	}
	total := len(filtered)

	// Simulate pagination
	start := offset
	end := offset + limit
	if start > total {
		return []*model.Campaign{}, total, nil
	}
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

// --- Test Function ---

func (m *MockCampaignRepoForPagination) GetByID(id int) (*model.Campaign, error) {
	// Return first campaign matching ID or nil
	for _, c := range m.campaigns {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockCampaignRepoForPagination) Create(c *model.Campaign) error {
	// no-op for test
	return nil
}

func (m *MockCampaignRepoForPagination) Update(c *model.Campaign) error {
	return nil
}

func (m *MockCampaignRepoForPagination) UpdateStatus(id int, status string) error {
	return nil
}

func TestListCampaignsPagination(t *testing.T) {
	// --- Seed only campaigns that match the filter ---
	totalCampaigns := 25 // total sms & draft campaigns
	campaigns := []*model.Campaign{}
	for i := 1; i <= totalCampaigns; i++ {
		campaigns = append(campaigns, &model.Campaign{
			ID:      i,
			Name:    "Campaign " + strconv.Itoa(i),
			Channel: "sms",
			Status:  "draft",
		})
	}

	// Initialize repo, service, controller
	repo := &MockCampaignRepoForPagination{campaigns: campaigns}
	svc := &service.CampaignService{CampaignRepo: repo}
	ctrl := &controller.CampaignController{CampaignService: svc}

	pageSize := 10
	seen := map[int]bool{}

	// Calculate total pages
	totalPages := (totalCampaigns + pageSize - 1) / pageSize

	for page := 1; page <= totalPages; page++ {
		// Build request
		req := httptest.NewRequest(
			"GET",
			"/campaigns?page="+strconv.Itoa(page)+
				"&page_size="+strconv.Itoa(pageSize)+
				"&channel=sms&status=draft",
			nil,
		)
		w := httptest.NewRecorder()

		// Call controller
		ctrl.ListCampaigns(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		// Decode response
		var res struct {
			Data       []model.Campaign `json:"data"`
			Pagination struct {
				Page       int `json:"page"`
				PageSize   int `json:"page_size"`
				TotalCount int `json:"total_count"`
			} `json:"pagination"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// --- Check pagination info ---
		if res.Pagination.Page != page {
			t.Errorf("expected page %d, got %d", page, res.Pagination.Page)
		}
		if res.Pagination.PageSize != pageSize {
			t.Errorf("expected page size %d, got %d", pageSize, res.Pagination.PageSize)
		}
		if res.Pagination.TotalCount != totalCampaigns {
			t.Errorf("expected total count %d, got %d", totalCampaigns, res.Pagination.TotalCount)
		}

		// --- Check data ---
		for _, c := range res.Data {
			// No duplicates
			if seen[c.ID] {
				t.Errorf("duplicate campaign ID %d across pages", c.ID)
			}
			seen[c.ID] = true

			// Filters
			if c.Channel != "sms" {
				t.Errorf("expected channel sms, got %s", c.Channel)
			}
			if c.Status != "draft" {
				t.Errorf("expected status draft, got %s", c.Status)
			}
		}
	}

	// --- Ensure all campaigns are returned ---
	if len(seen) != totalCampaigns {
		t.Errorf("expected %d unique campaigns, got %d", totalCampaigns, len(seen))
	}
}

func (m *MockCampaignRepo) CreateOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
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

func (m *MockCampaignRepo) GetOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
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

func (m *MockCampaignRepo) UpdateOutboundMessageStatus(id int, status, lastError string) error {
	return nil
}

func (m *MockCampaignRepo) GetCampaignStats(campaignID int) (map[string]int, error) {
	return map[string]int{"pending": 1, "sent": 0, "failed": 0}, nil
}

func (m *MockCampaignRepoForPagination) CreateOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
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

func (m *MockCampaignRepoForPagination) GetOutboundMessage(campaignID, customerID int) (*model.OutboundMessage, error) {
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

func (m *MockCampaignRepoForPagination) UpdateOutboundMessageStatus(id int, status, lastError string) error {
	return nil
}

func (m *MockCampaignRepoForPagination) GetCampaignStats(campaignID int) (map[string]int, error) {
	return map[string]int{"pending": 1, "sent": 0, "failed": 0}, nil
}

func (m *MockCampaignRepo) GetOutboundMessageByID(id int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID: id,
        RenderedContent: "Test message",
    }, nil
}



func (m *MockCampaignRepo) UpdateOutboundMessageContent(id int, content string) error {
    return nil
}

func (m *MockCampaignRepoForPagination) GetOutboundMessageByID(id int) (*model.OutboundMessage, error) {
    return &model.OutboundMessage{
        ID: id,
        RenderedContent: "Test message",
    }, nil
}

func (m *MockCampaignRepoForPagination) UpdateOutboundMessageContent(id int, content string) error {
    // no-op stub
    return nil
}
