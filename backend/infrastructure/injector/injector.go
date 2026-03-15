package injector

import (
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/handler"
	"github.com/renkonmaster/donguri/internal/repository"

	"gorm.io/gorm"
)

func InjectServer(db *gorm.DB) (*api.Server, error) {
	repo := repository.New(db)
	h := handler.New(repo)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
