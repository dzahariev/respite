package repo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

const (
	GLOBAL = "global"
)

var (
	MaxPageSize = 500
	MinPageSize = 10
)

type DBScopes struct {
	PageSize int
	Page     int
	Offset   int
	UserID   *uuid.UUID
	Global   bool
}

func NewDBScopes(pageSize, pageNumber, offset int, userID *uuid.UUID, isGlobal bool) DBScopes {
	return DBScopes{
		PageSize: pageSize,
		Page:     pageNumber,
		Offset:   offset,
		UserID:   userID,
		Global:   isGlobal,
	}
}

func NewDBScopesFromRequest(request *http.Request, isGlobal bool) DBScopes {
	return DBScopes{
		PageSize: getPageSize(request),
		Page:     getPage(request),
		Offset:   getOffset(request),
		UserID:   getCurrentUserID(request),
		Global:   isGlobal,
	}
}

func (dbs *DBScopes) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(dbs.Offset).Limit(dbs.PageSize)
	}
}

func (dbs *DBScopes) Owned() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if dbs.Global {
			return db
		} else {
			return db.Where("user_id = ?", dbs.UserID)
		}
	}
}

// getCurrentUserID returns the current request user ID
func getCurrentUserID(request *http.Request) *uuid.UUID {
	if request.Context().Value(CURRENT_USER_ID) == nil {
		return nil
	}
	if userID, ok := request.Context().Value(CURRENT_USER_ID).(string); ok {
		if userID == "" {
			return nil
		}
		id, err := uuid.FromString(userID)
		if err != nil {
			return nil
		}
		return &id
	}
	return nil
}

func getPageSize(request *http.Request) int {
	query := request.URL.Query()
	pageSize, _ := strconv.Atoi(query.Get("page_size")) // Error is ignored because wrong or missing parameters are handled as 0
	switch {
	case pageSize > MaxPageSize:
		pageSize = MaxPageSize
	case pageSize <= 0:
		pageSize = MinPageSize
	}
	return pageSize
}

func getPage(request *http.Request) int {
	query := request.URL.Query()
	page, _ := strconv.Atoi(query.Get("page")) // Error is ignored because wrong or missing parameters are handled as 0
	if page <= 0 {
		page = 1
	}
	return page
}

func getOffset(request *http.Request) int {
	return (getPage(request) - 1) * getPageSize(request)
}

// getCurrentUserPermissions returns the current request user ID
func getCurrentUserPermissions(request *http.Request) []string {
	if request.Context().Value(CURRENT_USER_PERMISSIONS) == nil {
		return nil
	}
	if permissions, ok := request.Context().Value(CURRENT_USER_PERMISSIONS).([]string); ok {
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
