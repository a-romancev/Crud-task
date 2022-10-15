package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const webAddr = "0.0.0.0:9999"

func main() {

	webServer := &http.Server{
		Addr:    webAddr,
		Handler: NewHandler(),
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Web listening on %s.", webAddr)
		err := webServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal("Web server failed.")
		}
	}()

	<-ctx.Done()
	log.Print("Shutting down.")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = webServer.Shutdown(shutdownCtx)
	log.Print("Shutdown complete.")
}
