package auth

import (
	"context"
	"errors"

	"github.com/Nerzal/gocloak/v13"
	"github.com/Nerzal/gocloak/v13/pkg/jwx"
	"github.com/dzahariev/respite/basemodel"
	"github.com/dzahariev/respite/cfg"
	"github.com/gofrs/uuid/v5"
)

type KeycloakClient struct {
	Client       *gocloak.GoCloak
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
}

// NewClient is used to init a client for Keycloak authentication
func NewClient(cfg *cfg.Keycloak) Client {
	return &KeycloakClient{
		Client:       gocloak.NewClient(cfg.AuthURL),
		URL:          cfg.AuthURL,
		Realm:        cfg.AuthRealm,
		ClientID:     cfg.AuthClientID,
		ClientSecret: cfg.AuthClientSecret,
	}
}

func (authClient *KeycloakClient) RetrospectToken(ctx context.Context, accessToken string) error {
	rptResult, err := authClient.Client.RetrospectToken(ctx, accessToken, authClient.ClientID, authClient.ClientSecret, authClient.Realm)
	if err != nil {
		return err
	}
	if !*rptResult.Active {
		return errors.New("token is not active")
	}

	return nil
}

func (authClient *KeycloakClient) GetRolesFromToken(ctx context.Context, accessToken string) ([]string, error) {
	jwxClaims := &jwx.Claims{}
	_, err := authClient.Client.DecodeAccessTokenCustomClaims(ctx, accessToken, authClient.Realm, jwxClaims)
	if err != nil {
		result := make([]string, 0)
		return result, err
	}
	return jwxClaims.RealmAccess.Roles, nil
}

// GetUserFromToken creates user entity from user info in token
func (authClient *KeycloakClient) GetUserFromToken(ctx context.Context, accessToken string) (*basemodel.User, error) {
	jwxClaims := &jwx.Claims{}
	_, err := authClient.Client.DecodeAccessTokenCustomClaims(ctx, accessToken, authClient.Realm, jwxClaims)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.FromString(jwxClaims.Subject)
	if err != nil {
		return nil, err
	}

	user := &basemodel.User{
		Base: basemodel.Base{
			ID: uid,
		},
		PreferedUserName: jwxClaims.PreferredUsername,
		GivenName:        jwxClaims.GivenName,
		FamilyName:       jwxClaims.FamilyName,
		Email:            jwxClaims.Email,
	}

	return user, nil
}
