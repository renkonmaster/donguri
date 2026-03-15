package handler

import (
	"context"
	"net/http"

	"github.com/ras0q/go-backend-template/internal/api"
)

// POST /api/v1/users
func (h *Handler) CreateUser(_ context.Context, _ *api.CreateUserReq) (*api.CreateUser, error) {
	return nil, &api.ErrorStatusCode{
		StatusCode: http.StatusNotImplemented,
		Response: api.Error{
			Message: "template endpoint is disabled",
		},
	}
}

// GET /api/v1/users/:userID
func (h *Handler) GetUser(_ context.Context, _ api.GetUserParams) (*api.User, error) {
	return nil, &api.ErrorStatusCode{
		StatusCode: http.StatusNotImplemented,
		Response: api.Error{
			Message: "template endpoint is disabled",
		},
	}
}

// GET /api/v1/users
func (h *Handler) GetUsers(_ context.Context) ([]api.User, error) {
	return nil, &api.ErrorStatusCode{
		StatusCode: http.StatusNotImplemented,
		Response: api.Error{
			Message: "template endpoint is disabled",
		},
	}
}
