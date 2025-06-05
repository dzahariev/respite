package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dzahariev/respite/auth"
	"github.com/dzahariev/respite/basemodel"
	"github.com/dzahariev/respite/cfg"
	"github.com/dzahariev/respite/repo"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	READ  = "read"
	WRITE = "write"
)

// Server represent current API server
type Server struct {
	ServerConfig      cfg.Server
	DB                *gorm.DB
	Router            *mux.Router
	AuthClient        auth.Client
	ResourceFactory   *repo.ResourceFactory
	RoleToPermissions map[string][]string
}

func NewServer(serverConfig cfg.Server, logConfig cfg.Logger, dbConfig cfg.DataBase, modelObjects []basemodel.Object, authClient auth.Client, roleToPermissions map[string][]string) (*Server, error) {
	// Initialise server instance
	server := &Server{}
	// Keep configuration
	server.ServerConfig = serverConfig
	// Initialise logger
	server.initLogger(logConfig)
	// Initialise global configurations
	repo.MaxPageSize = serverConfig.MaxPageSize
	repo.MinPageSize = serverConfig.MinPageSize
	// Store Auth Client
	server.AuthClient = authClient
	// Initlaise roles to permissions mapping
	server.RoleToPermissions = roleToPermissions
	// Initialise DB connection
	err := server.initDB(dbConfig)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return nil, err
	}
	// Register all resources
	server.initResourceFactory(modelObjects)
	// Initialise router and register all routes
	server.initRouter()
	slog.Info("Server initialized", "port", server.ServerConfig.Port, "db", dbConfig.DatabaseName)
	return server, nil
}

func (server *Server) initLogger(logConfig cfg.Logger) {
	var logLevel slog.Leveler
	switch logConfig.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelDebug
	}
	var logHandler slog.Handler
	if logConfig.Format == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel, AddSource: true})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel, AddSource: true})
	}
	slog.SetDefault(slog.New(logHandler))
	slog.Info("Logger initialized", "level", logConfig.Level, "format", logConfig.Format)
}

func (server *Server) initDB(dbConfig cfg.DataBase) error {
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DatabaseName, dbConfig.Password)
	var err error
	server.DB, err = gorm.Open(postgres.Open(DBURL), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return fmt.Errorf("cannot connect to database: %w", err)
	}
	slog.Info("Database connection established", "host", dbConfig.Host, "port", dbConfig.Port, "dbname", dbConfig.DatabaseName)
	return nil
}

// initResourceFactory is used to register all resources
func (server *Server) initResourceFactory(modelObjects []basemodel.Object) {
	server.ResourceFactory = &repo.ResourceFactory{Resources: map[string]repo.Resource{}}
	// Register user resource
	server.ResourceFactory.Register(&basemodel.User{})
	// Register all other provided resources
	for _, modelObject := range modelObjects {
		server.ResourceFactory.Register(modelObject)
	}
	slog.Info("Resource factory initialized", "resources", server.ResourceFactory.Names())
}

// initRouter is used to register routes
func (server *Server) initRouter() {
	server.Router = mux.NewRouter()
	server.Router.Use(loggerMiddleware)

	// Unsecured Home Route
	server.Router.HandleFunc(fmt.Sprintf("/%s/", server.ServerConfig.APIPath), server.Public(ContentTypeJSON(server.Home))).Methods(http.MethodGet)
	// Register all resource routes
	for _, resource := range server.ResourceFactory.Resources {
		server.Router.HandleFunc(fmt.Sprintf("/%s/%s", server.ServerConfig.APIPath, resource.Name), server.Protected(ContentTypeJSON(server.Create(resource.Name)), resource, WRITE)).Methods(http.MethodPost)
		server.Router.HandleFunc(fmt.Sprintf("/%s/%s", server.ServerConfig.APIPath, resource.Name), server.Protected(ContentTypeJSON(server.GetAll(resource.Name)), resource, READ)).Methods(http.MethodGet)
		server.Router.HandleFunc(fmt.Sprintf("/%s/%s/{id}", server.ServerConfig.APIPath, resource.Name), server.Protected(ContentTypeJSON(server.Get(resource.Name)), resource, READ)).Methods(http.MethodGet)
		server.Router.HandleFunc(fmt.Sprintf("/%s/%s/{id}", server.ServerConfig.APIPath, resource.Name), server.Protected(ContentTypeJSON(server.Update(resource.Name)), resource, WRITE)).Methods(http.MethodPut)
		server.Router.HandleFunc(fmt.Sprintf("/%s/%s/{id}", server.ServerConfig.APIPath, resource.Name), server.Protected(ContentTypeJSON(server.Delete(resource.Name)), resource, WRITE)).Methods(http.MethodDelete)
	}
	// Static Route
	server.Router.PathPrefix("/").Handler(server.Static())
	slog.Info("Router initialized", "routes", server.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		slog.Info("Registered route", "path", path, "methods", methods)
		return nil
	}))
}

// Run starts the http server
func (server *Server) Run() {
	addr := fmt.Sprintf("0.0.0.0:%s", server.ServerConfig.Port)
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: server.ServerConfig.WriteTimeout,
		ReadTimeout:  server.ServerConfig.ReadTimeout,
		IdleTimeout:  server.ServerConfig.IdleTimeout,
		Handler:      server.Router,
	}

	go func() {
		slog.Info("Listening on port", "port", server.ServerConfig.Port)
		err := srv.ListenAndServe()
		if err != nil {
			slog.Info("Error while serving", "error", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// Block until we receive termination signal.
	<-c
	// Wait for a deadline for termination.
	ctx, cancel := context.WithTimeout(context.Background(), server.ServerConfig.DeadlineOnInterrupt)
	defer cancel()
	slog.Info("Shutting down")
	srv.Shutdown(ctx)
	os.Exit(0)
}
