package api

import (
	"net/http"

	"github.com/dzahariev/respite/common"
)

// Home is an API root route controller
func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := common.GetLogger(ctx)
	logger.Debug("Home request received")
	JSON(w, http.StatusOK, server.Resources.Names())
}
