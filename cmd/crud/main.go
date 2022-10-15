package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/a-romancev/crud_task/company"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", ".", "PATH to config folder")
	flag.Parse()

	var conf Config
	conf = conf.WithFile(confPath)
	err := conf.Validate()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s/%s",
		conf.Mongo.Host,
		conf.Mongo.Password,
		conf.Mongo.Host,
		conf.Mongo.Database,
	)))
	if err != nil {
		log.Fatal("Failed to connect to mongo.")
	}
	mongoDB := mongoClient.Database("company")
	companyMongo := company.NewMongo(mongoDB)
	companyCRUD := company.NewCRUD(companyMongo)

	webServer := &http.Server{
		Addr:    conf.ListenWebAddress,
		Handler: NewHandler(companyCRUD),
	}
	go func() {
		log.Printf("Web listening on %s.", conf.ListenWebAddress)
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
