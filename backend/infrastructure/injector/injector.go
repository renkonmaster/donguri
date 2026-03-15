package injector

import (
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/handler"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"

	"gorm.io/gorm"
)

func InjectServer(db *gorm.DB) (*api.Server, error) {
	hub := stream.NewHub()
	repo := repository.New(db)
	h := handler.New(repo, hub)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
