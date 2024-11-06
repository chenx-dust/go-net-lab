package app

import (
	"github.com/chenx-dust/go-net-lab/paracat/config"
)

type Client struct {
	cfg *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg}
}

func (client *Client) Run() error {
	return nil
}
