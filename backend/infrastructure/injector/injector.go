package injector

import (
	"net/http"

	"github.com/ras0q/go-backend-template/internal/api"
	"github.com/ras0q/go-backend-template/internal/handler"
	"github.com/ras0q/go-backend-template/internal/repository"
	"github.com/ras0q/go-backend-template/internal/service/stream"

	"gorm.io/gorm"
)

func InjectServer(db *gorm.DB) (http.Handler, error) {
	hub := stream.NewHub()
	repo := repository.New(db)
	h := handler.New(repo, hub)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", s)
	mux.HandleFunc("/api/rooms/", h.StreamRoom)

	return mux, nil
}
