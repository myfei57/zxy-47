package main

import (
	"flag"
	"os"
)

type Config struct {
	DataDir string
	Listen  string
	Version string
}

func loadConfig() Config {
	cfg := Config{DataDir: "data", Listen: ":8080", Version: "1.0.0"}
	flag.StringVar(&cfg.DataDir, "data", cfg.DataDir, "data dir")
	flag.StringVar(&cfg.Listen, "listen", cfg.Listen, "listen address")
	flag.StringVar(&cfg.Version, "version", cfg.Version, "version")
	flag.Parse()
	if value := os.Getenv("GH_DATA_DIR"); value != "" {
		cfg.DataDir = value
	}
	if value := os.Getenv("GH_LISTEN"); value != "" {
		cfg.Listen = value
	}
	return cfg
}
