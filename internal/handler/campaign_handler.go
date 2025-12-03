// internal/handler/campaign_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/unclebandit/smsleopard-backend/internal/model"
	"github.com/unclebandit/smsleopard-backend/internal/repository"
	"github.com/unclebandit/smsleopard-backend/internal/service"

)

// CampaignHandler holds the dependencies for campaign-related HTTP handlers
type CampaignHandler struct {
	Repo    *repository.CampaignRepository
	Service *service.CampaignService
}


// NewCampaignHandler creates a new CampaignHandler with the given repository
func NewCampaignHandler(repo *repository.CampaignRepository) *CampaignHandler {
	return &CampaignHandler{
		Repo: repo,
	}
}

// CreateCampaignHandler handles creating a new campaign
func (h *CampaignHandler) CreateCampaignHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name         string     `json:"name"`
		Channel      string     `json:"channel"`
		BaseTemplate string     `json:"base_template"`
		ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	campaign := &model.Campaign{
		Name:         payload.Name,
		Channel:      payload.Channel,
		BaseTemplate: payload.BaseTemplate,
		Status:       "draft",
		ScheduledAt:  payload.ScheduledAt,
		CreatedAt:    time.Now(),
	}

	if err := h.Repo.Create(campaign); err != nil {
		http.Error(w, "failed to create campaign: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaign)
}

// ListCampaignsHandler returns a paginated list of campaigns
func (h *CampaignHandler) ListCampaignsHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	channel := r.URL.Query().Get("channel")
	status := r.URL.Query().Get("status")

	campaigns, pagination, err := h.Service.ListCampaigns(page, pageSize, channel, status)
	if err != nil {
		http.Error(w, "failed to fetch campaigns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":       campaigns,
		"pagination": pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


// GetCampaignHandler returns details of a single campaign by ID
func (h *CampaignHandler) GetCampaignHandler(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "invalid campaign id", http.StatusBadRequest)
        return
    }

    details, err := h.Service.GetCampaignDetailsWithStats(id)
    if err != nil {
        http.Error(w, "failed to fetch campaign: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(details)
}

func (h *CampaignHandler) GetCampaignHandlerWithStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	log.Println("📥 Handler called for campaign ID:", id)

	details, err := h.Service.GetCampaignDetailsWithStats(id)
	if err != nil {
		log.Println("❌ Error fetching campaign:", err)
		http.Error(w, "failed to fetch campaign: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Returning campaign details with stats: %+v\n", details)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}



