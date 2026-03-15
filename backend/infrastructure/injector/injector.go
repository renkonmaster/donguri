package injector

import (
	"github.com/ras0q/go-backend-template/internal/api"
	"github.com/ras0q/go-backend-template/internal/handler"
	"github.com/ras0q/go-backend-template/internal/repository"
	photo_service "github.com/ras0q/go-backend-template/internal/service/photo"

	"gorm.io/gorm"
)

func InjectServer(db *gorm.DB) (*api.Server, error) {
	photo := photo_service.NewPhotoService()
	repo := repository.New(db)
	h := handler.New(photo, repo)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
