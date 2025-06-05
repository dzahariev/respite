package repo

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/dzahariev/respite/basemodel"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

// Define a custom type for context keys
type contextKey string

const (
	LOGGER                   contextKey = "logger"
	CURRENT_USER_ID          contextKey = "currentUserID"
	CURRENT_USER_PERMISSIONS contextKey = "currentUserPermissions"
)

type Repository struct {
	DB        *gorm.DB
	DBScopes  DBScopes
	Resources *Resources
	RequestID uuid.UUID
}

// GetLogger is a helper to get logger from context or fallback
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LOGGER).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// NewRepository creates a new repository instance
func NewRepository(pageSize, pageNumber, offset int, userID *uuid.UUID, resourceName string, dataBase *gorm.DB, resources *Resources, currentUserPermissions []string) Repository {
	isGlobal := resources.IsGlobal(resourceName)
	dbScopes := NewDBScopes(pageSize, pageNumber, offset, userID, isGlobal)
	requestDatabase := dataBase.Scopes(dbScopes.Paginate())
	// If resource is not global and user do not have global permissions,
	// we scope the database to only owned resources
	if !isGlobal && !haveGlobalPermission(resourceName, currentUserPermissions) {
		requestDatabase = dataBase.Scopes(dbScopes.Owned(), dbScopes.Paginate())
	}

	return Repository{
		DB:        requestDatabase,
		DBScopes:  dbScopes,
		Resources: resources,
		RequestID: uuid.Must(uuid.NewV4()),
	}
}

func NewRepositoryFromRequest(request *http.Request, dataBase *gorm.DB, resourceName string, resources *Resources) Repository {
	isGlobal := resources.IsGlobal(resourceName)
	dbScopes := NewDBScopesFromRequest(request, isGlobal)
	currentUserPermissions := getCurrentUserPermissions(request)
	logger := GetLogger(request.Context())
	logger.Debug("Creating new repository", "resource", resourceName, "dbScopes", dbScopes, "userID", dbScopes.UserID, "global", isGlobal, "permissions", currentUserPermissions)
	return NewRepository(dbScopes.PageSize, dbScopes.Page, dbScopes.Offset, dbScopes.UserID, resourceName, dataBase, resources, currentUserPermissions)
}

// GetAll retrieves all objects
func (repository *Repository) GetAll(ctx context.Context, resourceName string) (*basemodel.List, error) {
	var err error
	object, err := repository.Resources.New(resourceName)
	if err != nil {
		return nil, err
	}

	count, err := object.Count(ctx, repository.DB, object)
	if err != nil {
		return nil, err
	}

	data, err := object.FindAll(ctx, repository.DB, object)
	if err != nil {
		return nil, err
	}

	list := &basemodel.List{
		Count:    count,
		PageSize: repository.DBScopes.PageSize,
		Page:     repository.DBScopes.Page,
		Data:     *data,
	}

	return list, nil
}

// Get loads an object by given ID
func (repository *Repository) Get(ctx context.Context, resourceName string, uid uuid.UUID) (basemodel.Object, error) {
	object, err := repository.Resources.New(resourceName)
	if err != nil {
		return nil, err
	}

	err = object.FindByID(ctx, repository.DB, object, uid)
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Create is caled to create an object
func (repository *Repository) Create(ctx context.Context, resourceName string, jsonObject []byte) (basemodel.Object, error) {
	object, err := repository.Resources.New(resourceName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonObject, object)
	if err != nil {
		return nil, err
	}

	err = object.Validate(ctx)
	if err != nil {
		return nil, err
	}

	if !repository.DBScopes.Global {
		ownerUUID := repository.DBScopes.UserID
		if ownerUUID == nil {
			return nil, err
		}
		objectAsLocalObject := object.(basemodel.LocalObject)
		objectAsLocalObject.SetUserID(*ownerUUID)
	}

	err = object.Save(ctx, repository.DB, object)

	if err != nil {
		return nil, err
	}

	return object, nil
}

// UpdateBook updates existing object
func (repository *Repository) Update(ctx context.Context, resourceName string, uid uuid.UUID, jsonObject []byte) (basemodel.Object, error) {
	object, err := repository.Resources.New(resourceName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonObject, &object)
	if err != nil {
		return nil, err
	}

	err = object.Validate(ctx)
	if err != nil {
		return nil, err
	}

	recordExisting := reflect.New(reflect.TypeOf(object).Elem()).Interface().(basemodel.Object)
	err = recordExisting.FindByID(ctx, repository.DB, recordExisting, uid)
	if err != nil {
		return nil, err
	}

	object.SetID(uid)

	err = object.Update(ctx, repository.DB, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Delete deletes an object
func (repository *Repository) Delete(ctx context.Context, resourceName string, uid uuid.UUID) error {
	object, err := repository.Resources.New(resourceName)
	if err != nil {
		return err
	}

	err = object.FindByID(ctx, repository.DB, object, uid)
	if err != nil {
		return err
	}

	err = object.Delete(ctx, repository.DB, object)
	if err != nil {
		return err
	}
	return nil
}
