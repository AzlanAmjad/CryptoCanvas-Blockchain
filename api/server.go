package api

import (
	"net"
	"net/http"

	network "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/peer-to-peer-network"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
)

// This is the API Server for the blockchain
// We use the Echo web framework to build our API here

// server configuration struct
type ServerConfig struct {
	// ListenAddr is the address the server listens on
	ListenAddr net.Addr
	Logger     log.Logger
	Node       *network.Server
}

// Server is the API Server for the blockchain
type Server struct {
	config *ServerConfig
}

// NewServer creates a new Server
func NewServer(config *ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

// Start starts the server
func (s *Server) Start() error {

	// create a new Echo instance
	e := echo.New()

	// routes for the API
	e.GET("/block/:index", s.getBlocks)

	// start the server
	err := e.Start(s.config.ListenAddr.String())
	if err != nil {
		return err
	}

	return nil
}

// getBlocks is the handler for the /block/:index route
func (s *Server) getBlocks(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"message": "getBlocks route hit"})
}
