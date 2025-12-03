// cmd/server/main.go
package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/go-chi/chi/v5"

	"github.com/unclebandit/smsleopard-backend/internal/controller"
	"github.com/unclebandit/smsleopard-backend/internal/db"
	"github.com/unclebandit/smsleopard-backend/internal/repository"
	"github.com/unclebandit/smsleopard-backend/internal/service"
    "github.com/unclebandit/smsleopard-backend/internal/queue"
    "github.com/unclebandit/smsleopard-backend/internal/handler"

)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, relying on OS environment variables")
	}

	// Init DB
	db.Init()
    q := queue.NewInMemoryQueue()

	customerRepo := &repository.CustomerRepository{DB: db.DB}
	campaignRepo := &repository.CampaignRepository{DB: db.DB}
    outboundRepo := &repository.OutboundMessageRepository{DB: db.DB}
    queue.StartCampaignSendSubscriber(q, campaignRepo)

	campaignService := &service.CampaignService{
		CampaignRepo: campaignRepo,
		CustomerRepo: customerRepo,
        OutboundRepo: outboundRepo,
        Queue:        q,  
	}

	campaignController := &controller.CampaignController{
		CampaignService: campaignService,
	}

    campaignHandler := &handler.CampaignHandler{
	Repo: campaignRepo,
	Service: campaignService, 
    }

    


	r := chi.NewRouter()

	// Campaign routes
	r.Post("/campaigns", campaignController.CreateCampaign)
	r.Get("/campaigns", campaignController.ListCampaigns)
	//r.Get("/campaigns/{id}", campaignController.GetCampaignDetails)
	r.Post("/campaigns/{id}/send", campaignController.SendCampaign)
	r.Post("/campaigns/{id}/personalized-preview", campaignController.PersonalizedPreview)
    r.Get("/campaigns/{id}", campaignHandler.GetCampaignHandlerWithStats)


	log.Println("🚀 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
