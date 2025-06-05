package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dzahariev/respite/repo"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
)

// GetAll retrieves all objects
func (server *Server) GetAll(resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		logger.Debug("GetAll request received", "resource", resourceName)
		repository := repo.NewRepositoryFromRequest(r, server.DB, resourceName, server.Resources)
		logger.Debug("Repository created", "resource", resourceName, "dbscopes", repository.DBScopes)
		list, err := repository.GetAll(ctx, resourceName)
		if err != nil {
			logger.Error("Error getting all objects", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}
		logger.Debug("Objects retrieved successfully", "resource", resourceName, "count", len(list.Data))
		JSON(w, http.StatusOK, list)
	}
}

// Get loads an object by given ID
func (server *Server) Get(resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		logger.Debug("Get request received", "resource", resourceName)
		vars := mux.Vars(r)
		uid, err := uuid.FromString(vars["id"])
		if err != nil {
			logger.Error("Error parsing UUID from request", "error", err)
			ERROR(w, http.StatusBadRequest, err)
			return
		}

		repository := repo.NewRepositoryFromRequest(r, server.DB, resourceName, server.Resources)
		object, err := repository.Get(ctx, resourceName, uid)
		if err != nil {
			//TODO If the object is not found, return 404 otherwise return 500
			logger.Error("Error getting object", "error", err)
			ERROR(w, http.StatusNotFound, err)
			return
		}
		logger.Debug("Object retrieved successfully", "resource", resourceName, "id", uid)
		JSON(w, http.StatusOK, object)
	}
}

// Create is caled to create an object
func (server *Server) Create(resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		logger.Debug("Create request received", "resource", resourceName)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Error reading request body", "error", err)
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		repository := repo.NewRepositoryFromRequest(r, server.DB, resourceName, server.Resources)
		object, err := repository.Create(ctx, resourceName, body)
		if err != nil {
			logger.Error("Error creating object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("%s%s/%v", r.Host, r.RequestURI, object.GetID()))
		logger.Debug("Object created successfully", "resource", resourceName, "id", object.GetID())
		JSON(w, http.StatusCreated, object)
	}
}

// UpdateBook updates existing object
func (server *Server) Update(resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		logger.Debug("Update request received", "resource", resourceName)
		vars := mux.Vars(r)
		uid, err := uuid.FromString(vars["id"])
		if err != nil {
			logger.Error("Error parsing UUID from request", "error", err)
			ERROR(w, http.StatusBadRequest, err)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Error reading request body", "error", err)
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		repository := repo.NewRepositoryFromRequest(r, server.DB, resourceName, server.Resources)
		object, err := repository.Update(ctx, resourceName, uid, body)
		if err != nil {
			logger.Error("Error updating object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}
		logger.Debug("Object updated successfully", "resource", resourceName, "id", uid)
		JSON(w, http.StatusOK, object)
	}
}

// Delete deletes an object
func (server *Server) Delete(resourceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		logger.Debug("Delete request received", "resource", resourceName)
		vars := mux.Vars(r)

		uid, err := uuid.FromString(vars["id"])
		if err != nil {
			logger.Error("Error parsing UUID from request", "error", err)
			ERROR(w, http.StatusBadRequest, err)
			return
		}
		repository := repo.NewRepositoryFromRequest(r, server.DB, resourceName, server.Resources)
		err = repository.Delete(ctx, resourceName, uid)
		if err != nil {
			logger.Error("Error deleting object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Entity", fmt.Sprintf("%s", uid))
		logger.Debug("Object deleted successfully", "resource", resourceName, "id", uid)
		JSON(w, http.StatusNoContent, "")
	}
}
