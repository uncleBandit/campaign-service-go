package service_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/repository"

)

// Mock repositories
type MockCustomerRepo struct{}
type MockCampaignRepo struct{}

type CampaignService struct {
    CustomerRepo repository.CustomerRepositoryInterface
    CampaignRepo repository.CampaignRepositoryInterface
}




func (m *MockCustomerRepo) GetByID(id int) (*model.Customer, error) {
	switch id {
	case 1:
		return &model.Customer{
			FirstName:       "Alice",
			LastName:        "Smith",
			Location:        "Nairobi",
			PreferredProduct: "Shoes",
		}, nil
	case 2:
		return &model.Customer{
			FirstName: "", LastName: "", Location: "", PreferredProduct: "",
		}, nil
	}
	return nil, nil
}

func (m *MockCustomerRepo) ListAll() ([]model.Customer, error) {
	return []model.Customer{
		{FirstName: "Alice", LastName: "Smith", Location: "Nairobi", PreferredProduct: "Shoes"},
		{FirstName: "Bob", LastName: "Jones", Location: "Mombasa", PreferredProduct: "Hat"},
	}, nil
}


func (m *MockCampaignRepo) GetByID(id int) (*model.Campaign, error) {
	return &model.Campaign{
		BaseTemplate: "Hi {first_name} {last_name}, check out {preferred_product} in {location}!",
	}, nil
}

// Stub implementations to satisfy interface
func (m *MockCampaignRepo) Create(c *model.Campaign) error          { return nil }
func (m *MockCampaignRepo) Update(c *model.Campaign) error          { return nil }
func (m *MockCampaignRepo) ListCampaigns(offset, limit int, channel, status string) ([]*model.Campaign, int, error) {
	return []*model.Campaign{}, 0, nil
}
func (m *MockCampaignRepo) UpdateStatus(id int, status string) error { return nil }

func (s *CampaignService) RenderPreview(campaignID, customerID int, overrideTemplate *string) (string, error) {
    campaign, err := s.CampaignRepo.GetByID(campaignID)
    if err != nil {
        return "", err
    }

    customer, err := s.CustomerRepo.GetByID(customerID)
    if err != nil || customer == nil {
        return "", fmt.Errorf("customer not found")
    }

    template := campaign.BaseTemplate
    if overrideTemplate != nil {
        template = *overrideTemplate
    }

    // Use <unknown> for empty fields
    firstName := customer.FirstName
    if firstName == "" {
        firstName = "<unknown>"
    }
    lastName := customer.LastName
    if lastName == "" {
        lastName = "<unknown>"
    }
    location := customer.Location
    if location == "" {
        location = "<unknown>"
    }
    preferred := customer.PreferredProduct
    if preferred == "" {
        preferred = "<unknown>"
    }

    // Simple placeholder replacement
    message := template
    message = strings.ReplaceAll(message, "{first_name}", firstName)
	message = strings.ReplaceAll(message, "{last_name}", lastName)
	message = strings.ReplaceAll(message, "{location}", location)
	message = strings.ReplaceAll(message, "{preferred_product}", preferred)


    return message, nil
}


func strPtr(s string) *string { return &s }


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


