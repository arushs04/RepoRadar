package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"supplygraph/internal/api"
	"supplygraph/internal/db"
	"supplygraph/internal/scanjobs"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	repo := db.NewRepository(database)
	runner := scanjobs.NewRunner(repo)
	if err := runner.ResumeQueuedJobs(context.Background()); err != nil {
		log.Fatalf("resume queued scan jobs: %v", err)
	}

	server := api.NewServer(repo, runner)

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("starting api server on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("serve http: %v", err)
	}
}
