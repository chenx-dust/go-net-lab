package main

import (
	"flag"
	"log"

	"github.com/chenx-dust/go-net-lab/paracat/app"
	"github.com/chenx-dust/go-net-lab/paracat/config"
)

func main() {
	cfgFilename := flag.String("c", "config.yaml", "config file")
	flag.Parse()

	cfg, err := config.LoadFromFile(*cfgFilename)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var application app.App
	if cfg.Mode == config.ClientMode {
		application = app.NewClient(cfg)
	} else if cfg.Mode == config.ServerMode {
		application = app.NewServer(cfg)
	} else if cfg.Mode == config.RelayMode {
		application = app.NewRelay(cfg)
	} else {
		log.Fatalf("Invalid mode: %v", cfg.Mode)
	}

	err = application.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
