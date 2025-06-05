package auth

import (
	"context"

	"github.com/dzahariev/respite/basemodel"
)

type Client interface {
	RetrospectToken(ctx context.Context, accessToken string) error
	GetRolesFromToken(ctx context.Context, accessToken string) ([]string, error)
	GetUserFromToken(ctx context.Context, accessToken string) (*basemodel.User, error)
}
