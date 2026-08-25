package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"greenhouse/internal/console"
	"greenhouse/internal/store"
)

func main() {
	cfg := loadConfig()
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	st := store.New(cfg.DataDir)
	services := console.Build(st)
	server := console.NewServer(services, cfg.Version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runControlLoop(ctx, services)

	httpServer := &http.Server{
		Addr:              cfg.Listen,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("greenhouse listening on %s", cfg.Listen)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func runControlLoop(ctx context.Context, services *console.Services) {
	lastRelease := map[string]string{}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			sheds, err := services.Sheds.List("")
			if err != nil {
				continue
			}
			for _, sh := range sheds {
				_ = services.Monitor.Evaluate(sh.ID)
				plans, _ := services.Plans.List(sh.ID)
				for _, plan := range plans {
					if plan.Enabled {
						_, _ = services.Cycle.Run(sh.ID, plan.ID)
					}
				}
				day := now.Format("2006-01-02")
				if lastRelease[sh.ID] != day {
					_, _ = services.Quotas.Release(sh.ID)
					lastRelease[sh.ID] = day
				}
			}
		}
	}
}
