package auth

import (
	"context"

	"github.com/dzahariev/respite/model"
)

type Client interface {
	RetrospectToken(ctx context.Context, accessToken string) error
	GetRolesFromToken(ctx context.Context, accessToken string) ([]string, error)
	GetUserFromToken(ctx context.Context, accessToken string) (*model.User, error)
}
