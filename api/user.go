package api

import (
	"context"

	"github.com/dzahariev/respite/model"
	"github.com/gofrs/uuid/v5"
)

// DBLoadUser loads an user by given ID
func (server *Server) DBLoadUser(ctx context.Context, userID string) (*model.User, error) {
	logger := GetLogger(ctx)
	logger.Debug("DBLoadUser request received", "userID", userID)
	uid, err := uuid.FromString(userID)
	if err != nil {
		return nil, err
	}

	user := &model.User{}
	err = user.FindByID(ctx, server.DB, user, uid)
	if err != nil {
		return nil, err
	}
	logger.Debug("User loaded successfully", "userID", userID, "user", user)
	return user, nil
}

// DBSaveUser is caled to save an user
func (server *Server) DBSaveUser(ctx context.Context, user *model.User) error {
	logger := GetLogger(ctx)
	logger.Debug("DBSaveUser request received", "user", user)
	err := user.Save(ctx, server.DB, user)

	if err != nil {
		return err
	}
	logger.Debug("User saved successfully", "userID", user.ID, "user", user)
	return nil
}
