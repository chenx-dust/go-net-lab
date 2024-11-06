package app

import (
	"github.com/chenx-dust/go-net-lab/paracat/config"
)

type Relay struct {
	cfg *config.Config
}

func NewRelay(cfg *config.Config) *Relay {
	return &Relay{cfg: cfg}
}

func (relay *Relay) Run() error {
	return nil
}
