package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/dzahariev/respite/domain"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type RequestContext struct {
	DB        *gorm.DB
	DBScopes  DBScopes
	Resource  Resource
	Resources *Resources
	RequestID uuid.UUID
}

// GetLogger is a helper to get logger from context or fallback
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// GetRequestContext is a helper to get RequestContext from context or nil if there is no such
func GetRequestContext(ctx context.Context) *RequestContext {
	if requestContext, ok := ctx.Value(RequestContextKey).(*RequestContext); ok {
		return requestContext
	}
	return nil
}

// NewRequestContextWithDetails creates a new RequestContext instance
func NewRequestContextWithDetails(pageSize, pageNumber, offset int, user *domain.User, resource Resource, dataBase *gorm.DB, resources *Resources, currentUserPermissions []string) *RequestContext {
	isGlobal := resources.IsGlobal(resource.Name)
	dbScopes := NewDBScopes(pageSize, pageNumber, offset, user, isGlobal)
	requestDatabase := dataBase.Scopes(dbScopes.Paginate())
	// If resource is not global and user do not have global permissions,
	// we scope the database to only owned resources
	if !isGlobal && !haveGlobalPermission(resource.Name, currentUserPermissions) {
		requestDatabase = dataBase.Scopes(dbScopes.Owned(), dbScopes.Paginate())
	}

	return &RequestContext{
		DB:        requestDatabase,
		DBScopes:  dbScopes,
		Resource:  resource,
		Resources: resources,
		RequestID: uuid.Must(uuid.NewV4()),
	}
}

func NewRequestContext(request *http.Request, dataBase *gorm.DB, resource Resource, resources *Resources) *RequestContext {
	isGlobal := resources.IsGlobal(resource.Name)
	dbScopes := NewDBScopesFromRequest(request, isGlobal)
	currentUserPermissions := getCurrentUserPermissions(request)
	logger := GetLogger(request.Context())
	logger.Debug("Creating new request context", "resource", resource.Name, "dbScopes", dbScopes, "userID", dbScopes.User, "global", isGlobal, "permissions", currentUserPermissions)
	return NewRequestContextWithDetails(dbScopes.PageSize, dbScopes.Page, dbScopes.Offset, dbScopes.User, resource, dataBase, resources, currentUserPermissions)
}

// GetAll retrieves all objects
func (requestContext *RequestContext) GetAll(ctx context.Context) (*domain.List, error) {
	var err error
	object, err := requestContext.Resources.New(requestContext.Resource.Name)
	if err != nil {
		return nil, err
	}

	count, err := object.Count(ctx, requestContext.DB, object)
	if err != nil {
		return nil, err
	}

	data, err := object.FindAll(ctx, requestContext.DB, object)
	if err != nil {
		return nil, err
	}

	list := &domain.List{
		Count:    count,
		PageSize: requestContext.DBScopes.PageSize,
		Page:     requestContext.DBScopes.Page,
		Data:     *data,
	}

	return list, nil
}

// Get loads an object by given ID
func (requestContext *RequestContext) Get(ctx context.Context, uid uuid.UUID) (domain.Object, error) {
	object, err := requestContext.Resources.New(requestContext.Resource.Name)
	if err != nil {
		return nil, err
	}

	err = object.FindByID(ctx, requestContext.DB, object, uid)
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Create is caled to create an object
func (requestContext *RequestContext) Create(ctx context.Context, jsonObject []byte) (domain.Object, error) {
	object, err := requestContext.Resources.New(requestContext.Resource.Name)
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

	if !requestContext.DBScopes.Global {
		ownerUser := requestContext.DBScopes.User
		if ownerUser == nil {
			return nil, err
		}
		objectAsLocalObject := object.(domain.LocalObject)
		objectAsLocalObject.SetUserID(ownerUser.ID)
	}

	err = object.Save(ctx, requestContext.DB, object)

	if err != nil {
		return nil, err
	}

	return object, nil
}

// Update updates existing object
func (requestContext *RequestContext) Update(ctx context.Context, uid uuid.UUID, jsonObject []byte) (domain.Object, error) {
	object, err := requestContext.Resources.New(requestContext.Resource.Name)
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

	recordExisting := reflect.New(reflect.TypeOf(object).Elem()).Interface().(domain.Object)
	err = recordExisting.FindByID(ctx, requestContext.DB, recordExisting, uid)
	if err != nil {
		return nil, err
	}

	object.SetID(uid)

	err = object.Update(ctx, requestContext.DB, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Delete deletes an object
func (requestContext *RequestContext) Delete(ctx context.Context, uid uuid.UUID) error {
	object, err := requestContext.Resources.New(requestContext.Resource.Name)
	if err != nil {
		return err
	}

	err = object.FindByID(ctx, requestContext.DB, object, uid)
	if err != nil {
		return err
	}

	err = object.Delete(ctx, requestContext.DB, object)
	if err != nil {
		return err
	}
	return nil
}

// getCurrentUserPermissions returns the current request user ID
func getCurrentUserPermissions(request *http.Request) []string {
	if request.Context().Value(CurrentUserPermissionsKey) == nil {
		return nil
	}
	if permissions, ok := request.Context().Value(CurrentUserPermissionsKey).([]string); ok {
		return permissions
	}
	return []string{}
}

// haveGlobalPermission is to check if the global permission for the resource is present in the list of permissions
func haveGlobalPermission(resource string, permissions []string) bool {
	for _, currentPermission := range permissions {
		resourcePermission := fmt.Sprintf("%s.%s", resource, GLOBAL)
		if strings.EqualFold(currentPermission, resourcePermission) {
			return true
		}
	}
	return false
}
