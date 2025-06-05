package api

import (
	"net/http"
)

// Home is an API root route controller
func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := GetLogger(ctx)
	logger.Debug("Home request received")
	JSON(w, http.StatusOK, server.ResourceFactory.Names())
}
