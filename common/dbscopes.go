package common

import (
	"net/http"
	"strconv"

	"github.com/dzahariev/respite/domain"
	"gorm.io/gorm"
)

var (
	MaxPageSize = 500
	MinPageSize = 10
)

type DBScopes struct {
	PageSize int
	Page     int
	Offset   int
	User     *domain.User
	Global   bool
}

func NewDBScopes(pageSize, pageNumber, offset int, user *domain.User, isGlobal bool) DBScopes {
	return DBScopes{
		PageSize: pageSize,
		Page:     pageNumber,
		Offset:   offset,
		User:     user,
		Global:   isGlobal,
	}
}

func NewDBScopesFromRequest(request *http.Request, isGlobal bool) DBScopes {
	return DBScopes{
		PageSize: getPageSize(request),
		Page:     getPage(request),
		Offset:   getOffset(request),
		User:     getCurrentUser(request),
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
			return db.Where("user_id = ?", dbs.User.ID.String())
		}
	}
}

// getCurrentUser returns the current request user ID
func getCurrentUser(request *http.Request) *domain.User {
	logger := GetLogger(request.Context())
	if request.Context().Value(CurrentUserKey) == nil {
		logger.Debug("Missing user in context")
		return nil
	}
	if user, ok := request.Context().Value(CurrentUserKey).(*domain.User); ok {
		return user
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
