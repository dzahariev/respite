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
func (server *Server) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		repository := repo.GetRequestContext(ctx)
		if repository == nil {
			logger.Error("Error reading repository from context")
			ERROR(w, http.StatusInternalServerError, fmt.Errorf("error reading repository from context"))
			return
		}
		logger.Debug("GetAll request received", "resource", repository.Resource.Name)

		list, err := repository.GetAll(ctx)
		if err != nil {
			logger.Error("Error getting all objects", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}
		logger.Debug("Objects retrieved successfully", "resource", repository.Resource.Name, "count", len(list.Data))
		JSON(w, http.StatusOK, list)
	}
}

// Get loads an object by given ID
func (server *Server) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)
		repository := repo.GetRequestContext(ctx)
		if repository == nil {
			logger.Error("Error reading repository from context")
			ERROR(w, http.StatusInternalServerError, fmt.Errorf("error reading repository from context"))
			return
		}
		logger.Debug("Get request received", "resource", repository.Resource.Name)

		vars := mux.Vars(r)
		uid, err := uuid.FromString(vars["id"])
		if err != nil {
			logger.Error("Error parsing UUID from request", "error", err)
			ERROR(w, http.StatusBadRequest, err)
			return
		}

		object, err := repository.Get(ctx, uid)
		if err != nil {
			//TODO If the object is not found, return 404 otherwise return 500
			logger.Error("Error getting object", "error", err)
			ERROR(w, http.StatusNotFound, err)
			return
		}
		logger.Debug("Object retrieved successfully", "resource", repository.Resource.Name, "id", uid)
		JSON(w, http.StatusOK, object)
	}
}

// Create is caled to create an object
func (server *Server) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)

		repository := repo.GetRequestContext(ctx)
		if repository == nil {
			logger.Error("Error reading repository from context")
			ERROR(w, http.StatusInternalServerError, fmt.Errorf("error reading repository from context"))
			return
		}
		logger.Debug("Create request received", "resource", repository.Resource)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Error reading request body", "error", err)
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		object, err := repository.Create(ctx, body)
		if err != nil {
			logger.Error("Error creating object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("%s%s/%v", r.Host, r.RequestURI, object.GetID()))
		logger.Debug("Object created successfully", "resource", repository.Resource.Name, "id", object.GetID())
		JSON(w, http.StatusCreated, object)
	}
}

// Update updates existing object
func (server *Server) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)

		repository := repo.GetRequestContext(ctx)
		if repository == nil {
			logger.Error("Error reading repository from context")
			ERROR(w, http.StatusInternalServerError, fmt.Errorf("error reading repository from context"))
			return
		}
		logger.Debug("Update request received", "resource", repository.Resource)

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
		object, err := repository.Update(ctx, uid, body)
		if err != nil {
			logger.Error("Error updating object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}
		logger.Debug("Object updated successfully", "resource", repository.Resource.Name, "id", uid)
		JSON(w, http.StatusOK, object)
	}
}

// Delete deletes an object
func (server *Server) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := repo.GetLogger(ctx)

		repository := repo.GetRequestContext(ctx)
		if repository == nil {
			logger.Error("Error reading repository from context")
			ERROR(w, http.StatusInternalServerError, fmt.Errorf("error reading repository from context"))
			return
		}
		logger.Debug("Delete request received", "resource", repository.Resource)

		vars := mux.Vars(r)
		uid, err := uuid.FromString(vars["id"])
		if err != nil {
			logger.Error("Error parsing UUID from request", "error", err)
			ERROR(w, http.StatusBadRequest, err)
			return
		}
		err = repository.Delete(ctx, uid)
		if err != nil {
			logger.Error("Error deleting object", "error", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Entity", fmt.Sprintf("%s", uid))
		logger.Debug("Object deleted successfully", "resource", repository.Resource.Name, "id", uid)
		JSON(w, http.StatusNoContent, "")
	}
}
