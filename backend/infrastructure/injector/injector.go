package injector

import (
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/handler"
	"github.com/renkonmaster/donguri/internal/repository"
	photo_service "github.com/renkonmaster/donguri/internal/service/photo"

	"github.com/jmoiron/sqlx"
)

func InjectServer(db *sqlx.DB) (*api.Server, error) {
	photo := photo_service.NewPhotoService()
	repo := repository.New(db)
	h := handler.New(photo, repo)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
