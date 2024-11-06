package app

import (
	"github.com/chenx-dust/go-net-lab/paracat/config"
)

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (server *Server) Run() error {
	return nil
}
